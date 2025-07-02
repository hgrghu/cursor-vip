package monitor

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	
	"github.com/kingparks/cursor-vip/tui/logger"
)

// Performance metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequests      int64
	HTTPErrors        int64
	HTTPSuccess       int64
	AvgResponseTime   int64 // microseconds
	
	// Application metrics
	MemoryUsage       uint64
	GoroutineCount    int
	CPUUsage          float64
	
	// Business metrics
	AuthAttempts      int64
	AuthSuccess       int64
	PaymentAttempts   int64
	PaymentSuccess    int64
	
	// Error tracking
	ErrorCount        int64
	LastError         string
	LastErrorTime     time.Time
	
	// Performance tracking
	StartTime         time.Time
	LastUpdate        time.Time
	
	mu sync.RWMutex
}

// Health status
type HealthStatus struct {
	Overall    string    `json:"overall"`
	HTTP       string    `json:"http"`
	Memory     string    `json:"memory"`
	Goroutines string    `json:"goroutines"`
	LastCheck  time.Time `json:"last_check"`
}

// Performance monitor
type Monitor struct {
	metrics     *Metrics
	ctx         context.Context
	cancel      context.CancelFunc
	interval    time.Duration
	healthChecks map[string]HealthChecker
	mu          sync.RWMutex
}

// Health checker interface
type HealthChecker interface {
	Check(ctx context.Context) error
	Name() string
}

// Global monitor instance
var GlobalMonitor *Monitor

// Initialize monitor
func InitMonitor(interval time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	
	GlobalMonitor = &Monitor{
		metrics: &Metrics{
			StartTime:  time.Now(),
			LastUpdate: time.Now(),
		},
		ctx:          ctx,
		cancel:       cancel,
		interval:     interval,
		healthChecks: make(map[string]HealthChecker),
	}
	
	// Start monitoring goroutine
	go GlobalMonitor.run()
	
	logger.Info("Performance monitor initialized")
}

// Stop monitor
func StopMonitor() {
	if GlobalMonitor != nil {
		GlobalMonitor.cancel()
		logger.Info("Performance monitor stopped")
	}
}

// Main monitoring loop
func (m *Monitor) run() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collectMetrics()
		}
	}
}

// Collect system metrics
func (m *Monitor) collectMetrics() {
	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()
	
	// Memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.metrics.MemoryUsage = memStats.Alloc
	
	// Goroutine count
	m.metrics.GoroutineCount = runtime.NumGoroutine()
	
	// Update timestamp
	m.metrics.LastUpdate = time.Now()
	
	// Log metrics periodically
	uptime := time.Since(m.metrics.StartTime)
	if uptime.Minutes() > 1 && int(uptime.Minutes())%5 == 0 {
		m.logMetrics()
	}
}

// Log current metrics
func (m *Monitor) logMetrics() {
	logger.InfoWithFields("Performance metrics", logger.Fields{
		"memory_mb":      m.metrics.MemoryUsage / 1024 / 1024,
		"goroutines":     m.metrics.GoroutineCount,
		"http_requests":  atomic.LoadInt64(&m.metrics.HTTPRequests),
		"http_errors":    atomic.LoadInt64(&m.metrics.HTTPErrors),
		"auth_attempts":  atomic.LoadInt64(&m.metrics.AuthAttempts),
		"uptime_hours":   time.Since(m.metrics.StartTime).Hours(),
	})
}

// HTTP request tracking
func (m *Monitor) RecordHTTPRequest(duration time.Duration, success bool) {
	atomic.AddInt64(&m.metrics.HTTPRequests, 1)
	
	if success {
		atomic.AddInt64(&m.metrics.HTTPSuccess, 1)
	} else {
		atomic.AddInt64(&m.metrics.HTTPErrors, 1)
	}
	
	// Update average response time
	durationMicros := duration.Microseconds()
	oldAvg := atomic.LoadInt64(&m.metrics.AvgResponseTime)
	requests := atomic.LoadInt64(&m.metrics.HTTPRequests)
	
	if requests > 1 {
		newAvg := (oldAvg*int64(requests-1) + durationMicros) / int64(requests)
		atomic.StoreInt64(&m.metrics.AvgResponseTime, newAvg)
	} else {
		atomic.StoreInt64(&m.metrics.AvgResponseTime, durationMicros)
	}
}

// Authentication tracking
func (m *Monitor) RecordAuth(success bool) {
	atomic.AddInt64(&m.metrics.AuthAttempts, 1)
	if success {
		atomic.AddInt64(&m.metrics.AuthSuccess, 1)
	}
	
	logger.LogAuth("attempt", "monitor", success)
}

// Payment tracking (开源版本已移除支付功能)
func (m *Monitor) RecordPayment(success bool) {
	// 开源版本不再记录支付相关指标
	return
}

// Error tracking
func (m *Monitor) RecordError(err error) {
	atomic.AddInt64(&m.metrics.ErrorCount, 1)
	
	m.metrics.mu.Lock()
	m.metrics.LastError = err.Error()
	m.metrics.LastErrorTime = time.Now()
	m.metrics.mu.Unlock()
	
	logger.Error("Recorded error: %v", err)
}

// Get current metrics
func (m *Monitor) GetMetrics() *Metrics {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()
	
	// Create a copy
	return &Metrics{
		HTTPRequests:    atomic.LoadInt64(&m.metrics.HTTPRequests),
		HTTPErrors:      atomic.LoadInt64(&m.metrics.HTTPErrors),
		HTTPSuccess:     atomic.LoadInt64(&m.metrics.HTTPSuccess),
		AvgResponseTime: atomic.LoadInt64(&m.metrics.AvgResponseTime),
		MemoryUsage:     m.metrics.MemoryUsage,
		GoroutineCount:  m.metrics.GoroutineCount,
		CPUUsage:        m.metrics.CPUUsage,
		AuthAttempts:    atomic.LoadInt64(&m.metrics.AuthAttempts),
		AuthSuccess:     atomic.LoadInt64(&m.metrics.AuthSuccess),
		PaymentAttempts: atomic.LoadInt64(&m.metrics.PaymentAttempts),
		PaymentSuccess:  atomic.LoadInt64(&m.metrics.PaymentSuccess),
		ErrorCount:      atomic.LoadInt64(&m.metrics.ErrorCount),
		LastError:       m.metrics.LastError,
		LastErrorTime:   m.metrics.LastErrorTime,
		StartTime:       m.metrics.StartTime,
		LastUpdate:      m.metrics.LastUpdate,
	}
}

// Health check system
func (m *Monitor) RegisterHealthCheck(checker HealthChecker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.healthChecks[checker.Name()] = checker
}

// Perform health checks
func (m *Monitor) HealthCheck(ctx context.Context) *HealthStatus {
	status := &HealthStatus{
		LastCheck: time.Now(),
	}
	
	// Check memory usage
	if m.metrics.MemoryUsage > 500*1024*1024 { // 500MB
		status.Memory = "WARNING"
	} else if m.metrics.MemoryUsage > 1024*1024*1024 { // 1GB
		status.Memory = "CRITICAL"
	} else {
		status.Memory = "OK"
	}
	
	// Check goroutine count
	if m.metrics.GoroutineCount > 100 {
		status.Goroutines = "WARNING"
	} else if m.metrics.GoroutineCount > 500 {
		status.Goroutines = "CRITICAL"
	} else {
		status.Goroutines = "OK"
	}
	
	// Check HTTP success rate
	httpRequests := atomic.LoadInt64(&m.metrics.HTTPRequests)
	httpErrors := atomic.LoadInt64(&m.metrics.HTTPErrors)
	
	if httpRequests > 0 {
		errorRate := float64(httpErrors) / float64(httpRequests)
		if errorRate > 0.1 { // 10% error rate
			status.HTTP = "WARNING"
		} else if errorRate > 0.2 { // 20% error rate
			status.HTTP = "CRITICAL"
		} else {
			status.HTTP = "OK"
		}
	} else {
		status.HTTP = "OK"
	}
	
	// Run custom health checks
	allHealthy := true
	for name, checker := range m.healthChecks {
		if err := checker.Check(ctx); err != nil {
			logger.Warn("Health check failed for %s: %v", name, err)
			allHealthy = false
		}
	}
	
	// Overall status
	if status.Memory == "CRITICAL" || status.Goroutines == "CRITICAL" || status.HTTP == "CRITICAL" || !allHealthy {
		status.Overall = "CRITICAL"
	} else if status.Memory == "WARNING" || status.Goroutines == "WARNING" || status.HTTP == "WARNING" {
		status.Overall = "WARNING"
	} else {
		status.Overall = "OK"
	}
	
	return status
}

// Memory optimization
func (m *Monitor) OptimizeMemory() {
	before := m.metrics.MemoryUsage
	
	// Force garbage collection
	runtime.GC()
	runtime.GC() // Second GC to clean up finalizers
	
	// Update metrics
	m.collectMetrics()
	after := m.metrics.MemoryUsage
	
	freed := int64(before) - int64(after)
	logger.Info("Memory optimization: freed %d bytes (%.2f MB)", freed, float64(freed)/1024/1024)
}

// Performance suggestions
func (m *Monitor) GetPerformanceSuggestions() []string {
	var suggestions []string
	
	// Memory suggestions
	if m.metrics.MemoryUsage > 512*1024*1024 { // 512MB
		suggestions = append(suggestions, "High memory usage detected. Consider restarting the application.")
	}
	
	// Goroutine suggestions
	if m.metrics.GoroutineCount > 100 {
		suggestions = append(suggestions, "High goroutine count detected. Check for goroutine leaks.")
	}
	
	// HTTP error rate suggestions
	httpRequests := atomic.LoadInt64(&m.metrics.HTTPRequests)
	httpErrors := atomic.LoadInt64(&m.metrics.HTTPErrors)
	
	if httpRequests > 10 && float64(httpErrors)/float64(httpRequests) > 0.1 {
		suggestions = append(suggestions, "High HTTP error rate. Check network connectivity and server status.")
	}
	
	// Response time suggestions
	avgResponseTime := atomic.LoadInt64(&m.metrics.AvgResponseTime)
	if avgResponseTime > 5000000 { // 5 seconds
		suggestions = append(suggestions, "Slow HTTP responses detected. Consider using a different server or check network.")
	}
	
	return suggestions
}

// Reset metrics
func (m *Monitor) Reset() {
	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()
	
	atomic.StoreInt64(&m.metrics.HTTPRequests, 0)
	atomic.StoreInt64(&m.metrics.HTTPErrors, 0)
	atomic.StoreInt64(&m.metrics.HTTPSuccess, 0)
	atomic.StoreInt64(&m.metrics.AvgResponseTime, 0)
	atomic.StoreInt64(&m.metrics.AuthAttempts, 0)
	atomic.StoreInt64(&m.metrics.AuthSuccess, 0)
	atomic.StoreInt64(&m.metrics.PaymentAttempts, 0)
	atomic.StoreInt64(&m.metrics.PaymentSuccess, 0)
	atomic.StoreInt64(&m.metrics.ErrorCount, 0)
	
	m.metrics.LastError = ""
	m.metrics.StartTime = time.Now()
	m.metrics.LastUpdate = time.Now()
	
	logger.Info("Performance metrics reset")
}

// Global convenience functions
func RecordHTTPRequest(duration time.Duration, success bool) {
	if GlobalMonitor != nil {
		GlobalMonitor.RecordHTTPRequest(duration, success)
	}
}

func RecordAuth(success bool) {
	if GlobalMonitor != nil {
		GlobalMonitor.RecordAuth(success)
	}
}

func RecordPayment(success bool) {
	if GlobalMonitor != nil {
		GlobalMonitor.RecordPayment(success)
	}
}

func RecordError(err error) {
	if GlobalMonitor != nil {
		GlobalMonitor.RecordError(err)
	}
}

func GetMetrics() *Metrics {
	if GlobalMonitor != nil {
		return GlobalMonitor.GetMetrics()
	}
	return nil
}

func OptimizeMemory() {
	if GlobalMonitor != nil {
		GlobalMonitor.OptimizeMemory()
	}
}

func GetHealthStatus() *HealthStatus {
	if GlobalMonitor != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return GlobalMonitor.HealthCheck(ctx)
	}
	return &HealthStatus{Overall: "UNKNOWN"}
}