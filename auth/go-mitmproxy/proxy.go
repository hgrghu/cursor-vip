package mitmproxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// Proxy represents a MITM proxy server
type Proxy struct {
	Addr        string
	server      *http.Server
	handlers    []Handler
	certs       *CertManager
	mu          sync.RWMutex
	running     bool
}

// Handler defines the interface for request/response handlers
type Handler interface {
	HandleRequest(req *http.Request) *http.Request
	HandleResponse(resp *http.Response) *http.Response
}

// CertManager manages SSL certificates for HTTPS interception
type CertManager struct {
	cert tls.Certificate
	mu   sync.RWMutex
}

// Flow represents an HTTP request/response flow
type Flow struct {
	Request  *http.Request
	Response *http.Response
	Error    error
}

// NewProxy creates a new MITM proxy instance
func NewProxy(addr string) *Proxy {
	return &Proxy{
		Addr:     addr,
		handlers: make([]Handler, 0),
		certs:    &CertManager{},
	}
}

// AddHandler adds a request/response handler to the proxy
func (p *Proxy) AddHandler(handler Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers = append(p.handlers, handler)
}

// Start starts the proxy server
func (p *Proxy) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.running {
		return fmt.Errorf("proxy is already running")
	}
	
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleHTTP)
	
	p.server = &http.Server{
		Addr:    p.Addr,
		Handler: mux,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	
	listener, err := net.Listen("tcp", p.Addr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	
	p.running = true
	
	go func() {
		if err := p.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Proxy server error: %v\n", err)
		}
	}()
	
	return nil
}

// Stop stops the proxy server
func (p *Proxy) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if !p.running {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := p.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown proxy server: %v", err)
	}
	
	p.running = false
	return nil
}

// handleHTTP handles HTTP requests
func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply request handlers
	p.mu.RLock()
	for _, handler := range p.handlers {
		r = handler.HandleRequest(r)
	}
	p.mu.RUnlock()
	
	// Handle CONNECT method for HTTPS
	if r.Method == http.MethodConnect {
		p.handleConnect(w, r)
		return
	}
	
	// Handle regular HTTP requests
	p.handleRequest(w, r)
}

// handleConnect handles HTTPS CONNECT requests
func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	// Simplified CONNECT handling
	// In a real implementation, this would establish a tunnel
	w.WriteHeader(http.StatusOK)
}

// handleRequest handles regular HTTP requests
func (p *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Create a reverse proxy
	target, err := url.Parse(fmt.Sprintf("%s://%s", r.URL.Scheme, r.Host))
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	
	if target.Scheme == "" {
		target.Scheme = "http"
	}
	
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Modify transport to handle TLS
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
	}
	
	// Modify response
	proxy.ModifyResponse = func(resp *http.Response) error {
		p.mu.RLock()
		defer p.mu.RUnlock()
		
		for _, handler := range p.handlers {
			resp = handler.HandleResponse(resp)
		}
		return nil
	}
	
	proxy.ServeHTTP(w, r)
}

// IsRunning returns whether the proxy is currently running
func (p *Proxy) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// DefaultHandler provides a basic handler implementation
type DefaultHandler struct{}

func (h *DefaultHandler) HandleRequest(req *http.Request) *http.Request {
	// Default implementation - just return the request as is
	return req
}

func (h *DefaultHandler) HandleResponse(resp *http.Response) *http.Response {
	// Default implementation - just return the response as is
	return resp
}

// CursorHandler is a specialized handler for Cursor IDE traffic
type CursorHandler struct{}

func (h *CursorHandler) HandleRequest(req *http.Request) *http.Request {
	// Add special handling for Cursor IDE requests
	if req.Header.Get("User-Agent") != "" {
		// Modify headers to appear as VIP user
		req.Header.Set("X-Cursor-VIP", "true")
	}
	return req
}

func (h *CursorHandler) HandleResponse(resp *http.Response) *http.Response {
	// Modify responses for Cursor IDE
	if resp.Header.Get("Content-Type") == "application/json" {
		// Could modify JSON responses here
	}
	return resp
}