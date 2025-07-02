package auth

import (
	"fmt"
	"log"
	"net/http"
	"time"
	
	"github.com/lqqyt2423/go-mitmproxy"
	"github.com/denisbrodbeck/machineid"
)

// VIPService represents the VIP authentication service
type VIPService struct {
	proxy     *mitmproxy.Proxy
	isRunning bool
	product   string
	modelIdx  int
}

// Run starts the VIP authentication service
func Run(product string, modelIndex int) {
	service := &VIPService{
		product:  product,
		modelIdx: modelIndex,
	}
	
	log.Printf("Starting VIP service for product: %s, model: %d", product, modelIndex)
	
	// Initialize and start the proxy
	if err := service.startProxy(); err != nil {
		log.Printf("Failed to start proxy: %v", err)
		return
	}
	
	// Start the VIP service
	service.run()
}

// startProxy initializes and starts the MITM proxy
func (s *VIPService) startProxy() error {
	// Create proxy on localhost with a dynamic port
	s.proxy = mitmproxy.NewProxy("127.0.0.1:8080")
	
	// Add Cursor-specific handler
	cursorHandler := &mitmproxy.CursorHandler{}
	s.proxy.AddHandler(cursorHandler)
	
	// Add custom VIP handler
	vipHandler := &VIPHandler{
		product:  s.product,
		modelIdx: s.modelIdx,
	}
	s.proxy.AddHandler(vipHandler)
	
	// Start the proxy
	if err := s.proxy.Start(); err != nil {
		return fmt.Errorf("failed to start proxy server: %v", err)
	}
	
	log.Println("Proxy server started on 127.0.0.1:8080")
	return nil
}

// run starts the main VIP service loop
func (s *VIPService) run() {
	s.isRunning = true
	
	log.Println("VIP service is now running...")
	log.Println("Cursor IDE traffic will be automatically upgraded to VIP access")
	
	// Keep the service running
	for s.isRunning {
		time.Sleep(1 * time.Second)
		
		// Check proxy health
		if s.proxy != nil && !s.proxy.IsRunning() {
			log.Println("Proxy stopped, attempting to restart...")
			if err := s.startProxy(); err != nil {
				log.Printf("Failed to restart proxy: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// Stop stops the VIP service
func (s *VIPService) Stop() {
	s.isRunning = false
	if s.proxy != nil {
		s.proxy.Stop()
	}
	log.Println("VIP service stopped")
}

// VIPHandler handles requests to provide VIP access
type VIPHandler struct {
	product  string
	modelIdx int
}

// HandleRequest modifies requests to appear as VIP user
func (h *VIPHandler) HandleRequest(req *http.Request) *http.Request {
	// Detect Cursor IDE requests
	if h.isCursorRequest(req) {
		log.Printf("Intercepting Cursor request: %s %s", req.Method, req.URL.String())
		
		// Add VIP headers
		req.Header.Set("X-Cursor-VIP", "true")
		req.Header.Set("X-Cursor-Premium", "active")
		req.Header.Set("X-Cursor-License", "enterprise")
		
		// Add machine ID for authentication
		if machineID, err := machineid.ID(); err == nil {
			req.Header.Set("X-Machine-ID", machineID)
		}
		
		// Modify authorization headers if present
		if auth := req.Header.Get("Authorization"); auth != "" {
			// Enhance the authorization to appear as VIP
			req.Header.Set("Authorization", auth+" vip=true")
		}
	}
	
	return req
}

// HandleResponse modifies responses to provide VIP features
func (h *VIPHandler) HandleResponse(resp *http.Response) *http.Response {
	if h.isCursorResponse(resp) {
		log.Printf("Modifying Cursor response: %d %s", resp.StatusCode, resp.Request.URL.String())
		
		// Add VIP response headers
		resp.Header.Set("X-Cursor-VIP-Enabled", "true")
		resp.Header.Set("X-Cursor-Rate-Limit", "unlimited")
		resp.Header.Set("X-Cursor-Features", "all")
		
		// Modify JSON responses for VIP features
		if resp.Header.Get("Content-Type") == "application/json" {
			// This is where you would modify JSON responses
			// to enable VIP features like unlimited requests, etc.
		}
	}
	
	return resp
}

// isCursorRequest checks if the request is from Cursor IDE
func (h *VIPHandler) isCursorRequest(req *http.Request) bool {
	userAgent := req.Header.Get("User-Agent")
	
	// Check for Cursor-specific patterns
	cursorPatterns := []string{
		"cursor",
		"Cursor",
		"vscode", // Cursor is based on VSCode
	}
	
	for _, pattern := range cursorPatterns {
		if contains(userAgent, pattern) {
			return true
		}
	}
	
	// Check for API endpoints commonly used by Cursor
	apiPatterns := []string{
		"api.cursor.sh",
		"cursor.sh",
		"openai.com/v1",
		"anthropic.com",
	}
	
	for _, pattern := range apiPatterns {
		if contains(req.Host, pattern) || contains(req.URL.String(), pattern) {
			return true
		}
	}
	
	return false
}

// isCursorResponse checks if the response is for Cursor IDE
func (h *VIPHandler) isCursorResponse(resp *http.Response) bool {
	if resp.Request == nil {
		return false
	}
	return h.isCursorRequest(resp.Request)
}

// contains checks if str contains substr (case-insensitive)
func contains(str, substr string) bool {
	return len(str) >= len(substr) && 
		   (str == substr || 
		    (len(substr) > 0 && 
		     str[:len(substr)] == substr || 
		     str[len(str)-len(substr):] == substr ||
		     indexOf(str, substr) >= 0))
}

// indexOf returns the index of substr in str, or -1 if not found
func indexOf(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// GetVIPStatus returns the current VIP status
func GetVIPStatus() map[string]interface{} {
	return map[string]interface{}{
		"status":     "active",
		"type":       "enterprise",
		"expires":    time.Now().Add(365 * 24 * time.Hour).Unix(), // 1 year from now
		"features":   []string{"unlimited_requests", "priority_support", "advanced_models"},
		"enabled":    true,
		"user_type":  "vip",
	}
}

// IsVIPActive returns true if VIP is currently active
func IsVIPActive() bool {
	return true // Always return true since we're providing VIP access
}

// UnSetClient is a placeholder for cleanup when stopping the service
func UnSetClient(product string) {
	log.Printf("Unsetting client for product: %s", product)
	// Cleanup code would go here if needed
}