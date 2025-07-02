package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
	
	"github.com/astaxie/beego/httplib"
	"github.com/kingparks/cursor-vip/auth/sign"
	"github.com/kingparks/cursor-vip/tui/params"
	"github.com/tidwall/gjson"
)

// HTTP client configuration
type HTTPConfig struct {
	MaxRetries      int
	RetryDelay      time.Duration
	Timeout         time.Duration
	ConnectTimeout  time.Duration
	CircuitBreaker  bool
	EnableGzip      bool
}

// Circuit breaker states
type CircuitBreakerState int

const (
	Closed CircuitBreakerState = iota
	HalfOpen
	Open
)

// Circuit breaker for HTTP requests
type CircuitBreaker struct {
	state          CircuitBreakerState
	failureCount   int
	lastFailTime   time.Time
	threshold      int
	timeout        time.Duration
	mutex          sync.RWMutex
}

// Enhanced HTTP manager
type HTTPManager struct {
	config         *HTTPConfig
	circuitBreaker *CircuitBreaker
	hosts          []string
	activeHost     string
	mutex          sync.RWMutex
}

// Request/Response types
type APIRequest struct {
	Method   string
	Path     string
	Headers  map[string]string
	Body     interface{}
	DeviceID string
}

type APIResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
	Error      error
}

// Initialize HTTP manager
func NewHTTPManager(hosts []string, config *HTTPConfig) *HTTPManager {
	if config == nil {
		config = &HTTPConfig{
			MaxRetries:     3,
			RetryDelay:     1 * time.Second,
			Timeout:        30 * time.Second,
			ConnectTimeout: 10 * time.Second,
			CircuitBreaker: true,
			EnableGzip:     true,
		}
	}
	
	cb := &CircuitBreaker{
		state:     Closed,
		threshold: 5,
		timeout:   30 * time.Second,
	}
	
	return &HTTPManager{
		config:         config,
		circuitBreaker: cb,
		hosts:          hosts,
	}
}

// Circuit breaker methods
func (cb *CircuitBreaker) CanRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	switch cb.state {
	case Closed:
		return true
	case HalfOpen:
		return true
	case Open:
		return time.Since(cb.lastFailTime) > cb.timeout
	default:
		return false
	}
}

func (cb *CircuitBreaker) OnSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.failureCount = 0
	cb.state = Closed
}

func (cb *CircuitBreaker) OnFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.failureCount++
	cb.lastFailTime = time.Now()
	
	if cb.failureCount >= cb.threshold {
		cb.state = Open
	} else if cb.state == HalfOpen {
		cb.state = Open
	}
}

// Host management
func (hm *HTTPManager) SetActiveHost() {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	
	// Try to find working host
	for _, host := range hm.hosts {
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		req := httplib.Get(host + "/health")
		req.SetTimeout(5*time.Second, 5*time.Second)
		
		if _, err := req.String(); err == nil {
			hm.activeHost = host
			cancel()
			return
		}
		cancel()
	}
	
	// Fall back to first host if none are reachable
	if len(hm.hosts) > 0 {
		hm.activeHost = hm.hosts[0]
	}
}

func (hm *HTTPManager) GetActiveHost() string {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	return hm.activeHost
}

// Enhanced request method with retry and circuit breaker
func (hm *HTTPManager) MakeRequest(ctx context.Context, request *APIRequest) (*APIResponse, error) {
	if hm.config.CircuitBreaker && !hm.circuitBreaker.CanRequest() {
		return nil, fmt.Errorf("circuit breaker is open")
	}
	
	var lastErr error
	maxRetries := hm.config.MaxRetries
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(hm.config.RetryDelay * time.Duration(attempt)):
			}
		}
		
		response, err := hm.makeHTTPRequest(ctx, request)
		if err == nil && response.StatusCode < 500 {
			if hm.config.CircuitBreaker {
				hm.circuitBreaker.OnSuccess()
			}
			return response, nil
		}
		
		lastErr = err
		if hm.config.CircuitBreaker {
			hm.circuitBreaker.OnFailure()
		}
		
		// Don't retry for client errors (4xx)
		if response != nil && response.StatusCode >= 400 && response.StatusCode < 500 {
			break
		}
	}
	
	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries+1, lastErr)
}

// Low-level HTTP request
func (hm *HTTPManager) makeHTTPRequest(ctx context.Context, request *APIRequest) (*APIResponse, error) {
	host := hm.GetActiveHost()
	if host == "" {
		return nil, fmt.Errorf("no active host available")
	}
	
	url := host + request.Path
	
	var req *httplib.BeegoHTTPRequest
	switch strings.ToUpper(request.Method) {
	case "GET":
		req = httplib.Get(url)
	case "POST":
		req = httplib.Post(url)
		if request.Body != nil {
			bodyData, err := json.Marshal(request.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			req.Body(bodyData)
			req.Header("Content-Type", "application/json")
		}
	case "PUT":
		req = httplib.Put(url)
		if request.Body != nil {
			bodyData, err := json.Marshal(request.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			req.Body(bodyData)
			req.Header("Content-Type", "application/json")
		}
	case "DELETE":
		req = httplib.Delete(url)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", request.Method)
	}
	
	// Set timeouts
	req.SetTimeout(hm.config.ConnectTimeout, hm.config.Timeout)
	
	// Set headers
	for key, value := range request.Headers {
		req.Header(key, value)
	}
	
	// Add authentication if device ID is provided
	if request.DeviceID != "" {
		req.Header("sign", sign.Sign(request.DeviceID))
	}
	
	// Enable gzip if configured
	if hm.config.EnableGzip {
		req.Setting(httplib.BeegoHTTPSettings{Gzip: true})
	}
	
	// Execute request with context
	respBody, err := req.String()
	response := &APIResponse{
		Body: respBody,
	}
	
	if err != nil {
		response.Error = err
		return response, err
	}
	
	return response, nil
}

// Convenience methods for common operations
func (hm *HTTPManager) Get(ctx context.Context, path string, deviceID string, headers map[string]string) (*APIResponse, error) {
	request := &APIRequest{
		Method:   "GET",
		Path:     path,
		Headers:  headers,
		DeviceID: deviceID,
	}
	return hm.MakeRequest(ctx, request)
}

func (hm *HTTPManager) Post(ctx context.Context, path string, body interface{}, deviceID string, headers map[string]string) (*APIResponse, error) {
	request := &APIRequest{
		Method:   "POST",
		Path:     path,
		Body:     body,
		Headers:  headers,
		DeviceID: deviceID,
	}
	return hm.MakeRequest(ctx, request)
}

func (hm *HTTPManager) GetJSON(ctx context.Context, path string, deviceID string) (gjson.Result, error) {
	response, err := hm.Get(ctx, path, deviceID, nil)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.Parse(response.Body), nil
}

func (hm *HTTPManager) PostJSON(ctx context.Context, path string, body interface{}, deviceID string) (gjson.Result, error) {
	response, err := hm.Post(ctx, path, body, deviceID, nil)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.Parse(response.Body), nil
}

// Health check endpoint
func (hm *HTTPManager) HealthCheck(ctx context.Context) error {
	_, err := hm.Get(ctx, "/health", "", nil)
	return err
}

// Set proxy configuration
func (hm *HTTPManager) SetProxy(lang string) {
	hm.SetActiveHost()
	
	proxy := httplib.BeegoHTTPSettings{}.Proxy
	proxyText := ""
	
	if os.Getenv("http_proxy") != "" {
		proxy = func(request *http.Request) (*url.URL, error) {
			return url.Parse(os.Getenv("http_proxy"))
		}
		proxyText = os.Getenv("http_proxy") + " " + params.Trr.Tr("经由") + " http_proxy " + params.Trr.Tr("代理访问")
	}
	if os.Getenv("https_proxy") != "" {
		proxy = func(request *http.Request) (*url.URL, error) {
			return url.Parse(os.Getenv("https_proxy"))
		}
		proxyText = os.Getenv("https_proxy") + " " + params.Trr.Tr("经由") + " https_proxy " + params.Trr.Tr("代理访问")
	}
	if os.Getenv("all_proxy") != "" {
		proxy = func(request *http.Request) (*url.URL, error) {
			return url.Parse(os.Getenv("all_proxy"))
		}
		proxyText = os.Getenv("all_proxy") + " " + params.Trr.Tr("经由") + " all_proxy " + params.Trr.Tr("代理访问")
	}
	
	userAgent := fmt.Sprintf(`{"lang":"%s","GOOS":"%s","ARCH":"%s","version":%d,"deviceID":"%s","machineID":"%s","sign":"%s","mode":%d}`,
		lang, runtime.GOOS, runtime.GOARCH, params.Version, params.DeviceID, params.MachineID, sign.Sign(params.DeviceID), params.Mode)
	
	httplib.SetDefaultSetting(httplib.BeegoHTTPSettings{
		Proxy:            proxy,
		ReadWriteTimeout: hm.config.Timeout,
		ConnectTimeout:   hm.config.ConnectTimeout,
		Gzip:             hm.config.EnableGzip,
		DumpBody:         true,
		UserAgent:        userAgent,
	})
	
	if len(proxyText) > 0 {
		_, _ = fmt.Fprintf(params.ColorOut, params.Yellow, proxyText)
	}
}

// Response helper methods
func (r *APIResponse) GetString(path string) string {
	return gjson.Get(r.Body, path).String()
}

func (r *APIResponse) GetInt(path string) int64 {
	return gjson.Get(r.Body, path).Int()
}

func (r *APIResponse) GetBool(path string) bool {
	return gjson.Get(r.Body, path).Bool()
}

func (r *APIResponse) HasError() bool {
	return r.Error != nil || r.StatusCode >= 400
}