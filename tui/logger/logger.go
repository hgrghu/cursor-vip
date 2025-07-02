package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Log levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Enhanced logger with file rotation
type Logger struct {
	level      LogLevel
	file       *os.File
	fileSize   int64
	maxSize    int64
	maxBackups int
	mu         sync.Mutex
	prefix     string
}

// Global logger instance
var globalLogger *Logger

// Initialize logger
func Init(level LogLevel, logDir string, maxSize int64, maxBackups int) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	
	logFile := filepath.Join(logDir, "cursor-vip.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat log file: %w", err)
	}
	
	globalLogger = &Logger{
		level:      level,
		file:       file,
		fileSize:   stat.Size(),
		maxSize:    maxSize,
		maxBackups: maxBackups,
		prefix:     "[cursor-vip]",
	}
	
	return nil
}

// Initialize with defaults
func InitDefault() error {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".cursor-vip", "logs")
	return Init(INFO, logDir, 10*1024*1024, 5) // 10MB max, 5 backups
}

// Close logger
func Close() {
	if globalLogger != nil {
		globalLogger.mu.Lock()
		defer globalLogger.mu.Unlock()
		if globalLogger.file != nil {
			globalLogger.file.Close()
		}
	}
}

// Get caller info
func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// Log with level
func (l *Logger) log(level LogLevel, msg string, args ...interface{}) {
	if level < l.level {
		return
	}
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Check if we need to rotate
	if l.fileSize > l.maxSize {
		l.rotate()
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	caller := getCallerInfo(3)
	levelName := levelNames[level]
	
	logLine := fmt.Sprintf("%s [%s] %s [%s] %s\n", 
		timestamp, levelName, l.prefix, caller, fmt.Sprintf(msg, args...))
	
	// Write to file
	if l.file != nil {
		n, _ := l.file.WriteString(logLine)
		l.fileSize += int64(n)
		l.file.Sync()
	}
	
	// Also write to stderr for errors and fatal
	if level >= ERROR {
		fmt.Fprint(os.Stderr, logLine)
	}
}

// Rotate log files
func (l *Logger) rotate() {
	if l.file == nil {
		return
	}
	
	l.file.Close()
	
	// Rotate existing backup files
	logPath := l.file.Name()
	for i := l.maxBackups - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", logPath, i)
		newPath := fmt.Sprintf("%s.%d", logPath, i+1)
		os.Rename(oldPath, newPath)
	}
	
	// Move current log to .1
	os.Rename(logPath, logPath+".1")
	
	// Create new log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("Failed to rotate log file: %v", err)
		return
	}
	
	l.file = file
	l.fileSize = 0
}

// Set log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Get log level
func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// Global logging functions
func Debug(msg string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(DEBUG, msg, args...)
	}
}

func Info(msg string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(INFO, msg, args...)
	}
}

func Warn(msg string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(WARN, msg, args...)
	}
}

func Error(msg string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(ERROR, msg, args...)
	}
}

func Fatal(msg string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(FATAL, msg, args...)
	}
	os.Exit(1)
}

// Structured logging
type Fields map[string]interface{}

func (f Fields) String() string {
	var result string
	for k, v := range f {
		if result != "" {
			result += " "
		}
		result += fmt.Sprintf("%s=%v", k, v)
	}
	return result
}

func DebugWithFields(msg string, fields Fields) {
	Debug("%s | %s", msg, fields.String())
}

func InfoWithFields(msg string, fields Fields) {
	Info("%s | %s", msg, fields.String())
}

func WarnWithFields(msg string, fields Fields) {
	Warn("%s | %s", msg, fields.String())
}

func ErrorWithFields(msg string, fields Fields) {
	Error("%s | %s", msg, fields.String())
}

// HTTP request logging
func LogHTTPRequest(method, url string, status int, duration time.Duration) {
	Info("HTTP %s %s - %d (%v)", method, url, status, duration)
}

// Log configuration changes
func LogConfigChange(key, oldValue, newValue string) {
	Info("Config changed: %s %s -> %s", key, oldValue, newValue)
}

// Log authentication events
func LogAuth(event, deviceID string, success bool) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}
	Info("Auth %s: device=%s status=%s", event, deviceID, status)
}

// Log payment events (开源版本已移除支付功能)
func LogPayment(event, orderID, deviceID string, amount float64) {
	// 开源版本不再记录支付事件
	return
}

// Performance monitoring
func LogPerformance(operation string, duration time.Duration, success bool) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}
	
	level := INFO
	if duration > 5*time.Second {
		level = WARN
	}
	if duration > 10*time.Second {
		level = ERROR
	}
	
	if globalLogger != nil {
		globalLogger.log(level, "Performance: %s took %v (%s)", operation, duration, status)
	}
}