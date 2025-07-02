package params

import (
	"github.com/unknwon/i18n"
	"io"
	"os"
	"sync"
	"time"
	"os/signal"
	"syscall"
	"github.com/mattn/go-colorable"
)

var Mode int64           // 1模式1 2模式2
var CursorVersion string // cursor版本号
var Lang string
var ExclusiveToken string
var M3c string
var Promotion string
var DeviceID string
var MachineID string
var ColorOut io.Writer
var Sigs chan os.Signal
var SigCountDown chan int
var Trr *Tr

// Enhanced configuration structure
type Config struct {
	Lang       string `json:"lang"`
	Mode       int64  `json:"mode"`
	Promotion  string `json:"promotion"`
	HTTPTimeout time.Duration `json:"http_timeout"`
	MaxRetries int    `json:"max_retries"`
	LogLevel   string `json:"log_level"`
	mutex      sync.RWMutex
}

// Application state management
type AppState struct {
	IsRunning    bool
	StartTime    time.Time
	ErrorCount   int
	LastError    error
	mutex        sync.RWMutex
}

// Enhanced error types
type AppError struct {
	Code    int
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Global application state
var (
	GlobalConfig *Config
	GlobalState  *AppState
	
	// Enhanced signal handling
	SigCountDown chan int
	Sigs         chan os.Signal
	
	// UI components
	ColorOut *colorable.Colorable
	Trr      *Tr
	
	// Application metadata
	DeviceID    string
	MachineID   string
	M3c         string
	ExclusiveToken string
)

type Tr struct {
	i18n.Locale
}

func (t *Tr) Tr(format string, args ...interface{}) string {
	return t.Locale.Tr(format, args...)
}

// Initialize global state
func init() {
	GlobalConfig = &Config{
		HTTPTimeout: 30 * time.Second,
		MaxRetries:  3,
		LogLevel:    "info",
	}
	
	GlobalState = &AppState{
		StartTime: time.Now(),
	}
	
	// Setup signal handling
	Sigs = make(chan os.Signal, 1)
	signal.Notify(Sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
}

// Thread-safe config access
func (c *Config) GetLang() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.Lang
}

func (c *Config) SetLang(lang string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Lang = lang
}

func (c *Config) GetMode() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.Mode
}

func (c *Config) SetMode(mode int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Mode = mode
}

// Thread-safe state management
func (s *AppState) SetRunning(running bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.IsRunning = running
}

func (s *AppState) IsAppRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.IsRunning
}

func (s *AppState) IncrementErrorCount() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ErrorCount++
}

func (s *AppState) SetLastError(err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.LastError = err
	s.ErrorCount++
}
