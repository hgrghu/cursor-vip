package client

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"time"
	
	"github.com/kingparks/cursor-vip/tui/params"
	"github.com/kingparks/cursor-vip/tui/tool"
)

// Enhanced client with improved error handling and HTTP management
type EnhancedClient struct {
	httpManager *HTTPManager
	config      *ClientConfig
}

// Client configuration
type ClientConfig struct {
	Hosts          []string
	DefaultTimeout time.Duration
	MaxRetries     int
	EnableMetrics  bool
}

// Metrics for monitoring
type ClientMetrics struct {
	RequestCount    int64
	ErrorCount      int64
	SuccessCount    int64
	AverageLatency  time.Duration
	LastRequestTime time.Time
}

// Global enhanced client instance
var EnhancedCli *EnhancedClient
var clientMetrics *ClientMetrics

// Initialize enhanced client
func InitEnhancedClient(hosts []string) {
	config := &ClientConfig{
		Hosts:          hosts,
		DefaultTimeout: 30 * time.Second,
		MaxRetries:     3,
		EnableMetrics:  true,
	}
	
	httpConfig := &HTTPConfig{
		MaxRetries:     config.MaxRetries,
		RetryDelay:     1 * time.Second,
		Timeout:        config.DefaultTimeout,
		ConnectTimeout: 10 * time.Second,
		CircuitBreaker: true,
		EnableGzip:     true,
	}
	
	EnhancedCli = &EnhancedClient{
		httpManager: NewHTTPManager(hosts, httpConfig),
		config:      config,
	}
	
	clientMetrics = &ClientMetrics{}
}

// Set proxy for enhanced client
func (ec *EnhancedClient) SetProxy(lang string) {
	ec.httpManager.SetProxy(lang)
}

// Enhanced error handling wrapper
func (ec *EnhancedClient) withMetrics(operation string, fn func() (*APIResponse, error)) (*APIResponse, error) {
	if ec.config.EnableMetrics {
		clientMetrics.RequestCount++
		clientMetrics.LastRequestTime = time.Now()
	}
	
	startTime := time.Now()
	response, err := fn()
	duration := time.Since(startTime)
	
	if ec.config.EnableMetrics {
		if err != nil {
			clientMetrics.ErrorCount++
		} else {
			clientMetrics.SuccessCount++
		}
		
		// Update average latency
		if clientMetrics.AverageLatency == 0 {
			clientMetrics.AverageLatency = duration
		} else {
			clientMetrics.AverageLatency = (clientMetrics.AverageLatency + duration) / 2
		}
	}
	
	return response, err
}

// Enhanced API methods with better error handling

func (ec *EnhancedClient) GetAD() string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	response, err := ec.withMetrics("GetAD", func() (*APIResponse, error) {
		return ec.httpManager.Get(ctx, "/ad", "", nil)
	})
	
	if err != nil {
		params.GlobalState.SetLastError(fmt.Errorf("failed to get ad: %w", err))
		return ""
	}
	
	return response.Body
}

func (ec *EnhancedClient) GetPayUrl() (payUrl, orderID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	response, err := ec.withMetrics("GetPayUrl", func() (*APIResponse, error) {
		return ec.httpManager.Get(ctx, "/payUrl", "", nil)
	})
	
	if err != nil {
		params.GlobalState.SetLastError(fmt.Errorf("failed to get pay URL: %w", err))
		return "", ""
	}
	
	return response.GetString("payUrl"), response.GetString("orderID")
}

func (ec *EnhancedClient) GetExclusivePayUrl() (payUrl, orderID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	response, err := ec.withMetrics("GetExclusivePayUrl", func() (*APIResponse, error) {
		return ec.httpManager.Get(ctx, "/exclusivePayUrl", "", nil)
	})
	
	if err != nil {
		params.GlobalState.SetLastError(fmt.Errorf("failed to get exclusive pay URL: %w", err))
		return "", ""
	}
	
	return response.GetString("payUrl"), response.GetString("orderID")
}

// Generic payment URL method to reduce duplication
func (ec *EnhancedClient) getPaymentURL(endpoint string) (payUrl, orderID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	response, err := ec.withMetrics("GetPaymentURL", func() (*APIResponse, error) {
		return ec.httpManager.Get(ctx, endpoint, "", nil)
	})
	
	if err != nil {
		params.GlobalState.SetLastError(fmt.Errorf("failed to get payment URL from %s: %w", endpoint, err))
		return "", ""
	}
	
	return response.GetString("payUrl"), response.GetString("orderID")
}

// Use generic method for all payment URLs
func (ec *EnhancedClient) GetM3PayUrl() (payUrl, orderID string) {
	return ec.getPaymentURL("/m3PayUrl")
}

func (ec *EnhancedClient) GetM3tPayUrl() (payUrl, orderID string) {
	return ec.getPaymentURL("/m3tPayUrl")
}

func (ec *EnhancedClient) GetM3hPayUrl() (payUrl, orderID string) {
	return ec.getPaymentURL("/m3hPayUrl")
}

// Generic payment check method
func (ec *EnhancedClient) checkPayment(endpoint, orderID, deviceID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	path := fmt.Sprintf("%s?orderID=%s&deviceID=%s", endpoint, orderID, deviceID)
	response, err := ec.withMetrics("CheckPayment", func() (*APIResponse, error) {
		return ec.httpManager.Get(ctx, path, deviceID, nil)
	})
	
	if err != nil {
		params.GlobalState.SetLastError(fmt.Errorf("failed to check payment: %w", err))
		return false
	}
	
	return response.GetBool("isPay")
}

func (ec *EnhancedClient) PayCheck(orderID, deviceID string) bool {
	return ec.checkPayment("/payCheck", orderID, deviceID)
}

func (ec *EnhancedClient) ExclusivePayCheck(orderID, deviceID string) bool {
	return ec.checkPayment("/exclusivePayCheck", orderID, deviceID)
}

func (ec *EnhancedClient) M3PayCheck(orderID, deviceID string) bool {
	return ec.checkPayment("/m3PayCheck", orderID, deviceID)
}

func (ec *EnhancedClient) M3tPayCheck(orderID, deviceID string) bool {
	return ec.checkPayment("/m3tPayCheck", orderID, deviceID)
}

func (ec *EnhancedClient) M3hPayCheck(orderID, deviceID string) bool {
	return ec.checkPayment("/m3hPayCheck", orderID, deviceID)
}

func (ec *EnhancedClient) DelFToken(deviceID, category string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	path := fmt.Sprintf("/delFToken?category=%s", category)
	_, err := ec.withMetrics("DelFToken", func() (*APIResponse, error) {
		request := &APIRequest{
			Method:   "DELETE",
			Path:     path,
			DeviceID: deviceID,
		}
		return ec.httpManager.MakeRequest(ctx, request)
	})
	
	if err != nil {
		params.GlobalState.SetLastError(fmt.Errorf("failed to delete token: %w", err))
		return err
	}
	
	return nil
}

func (ec *EnhancedClient) CheckFToken(deviceID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	response, err := ec.withMetrics("CheckFToken", func() (*APIResponse, error) {
		return ec.httpManager.Get(ctx, "/checkFToken", deviceID, nil)
	})
	
	if err != nil {
		params.GlobalState.SetLastError(fmt.Errorf("failed to check token: %w", err))
		return false
	}
	
	return response.GetBool("has")
}

func (ec *EnhancedClient) UpExclusiveStatus(exclusiveUsed, exclusiveTotal int64, exclusiveErr, exclusiveToken, deviceID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	body := map[string]interface{}{
		"exclusiveUsed":  exclusiveUsed,
		"exclusiveTotal": exclusiveTotal,
		"exclusiveErr":   exclusiveErr,
		"exclusiveToken": exclusiveToken,
	}
	
	_, err := ec.withMetrics("UpExclusiveStatus", func() (*APIResponse, error) {
		return ec.httpManager.Post(ctx, "/upExclusiveStatus", body, deviceID, nil)
	})
	
	if err != nil {
		params.GlobalState.SetLastError(fmt.Errorf("failed to update exclusive status: %w", err))
	}
}

func (ec *EnhancedClient) UpChecksumPrefix(p, deviceID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	body := map[string]interface{}{"p": p}
	
	_, err := ec.withMetrics("UpChecksumPrefix", func() (*APIResponse, error) {
		return ec.httpManager.Post(ctx, "/upChecksumPrefix", body, deviceID, nil)
	})
	
	if err != nil {
		params.GlobalState.SetLastError(fmt.Errorf("failed to update checksum prefix: %w", err))
	}
}

func (ec *EnhancedClient) GetMyInfo(deviceID string) (sCount, sPayCount, isPay, ticket, exp, exclusiveAt, token, m3c, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	
	dUser, _ := user.Current()
	deviceName := ""
	if dUser != nil {
		deviceName = dUser.Name
		if deviceName == "" {
			deviceName = dUser.Username
		}
	}
	
	body := map[string]string{
		"device":    deviceID,
		"deviceMac": tool.GetMac_241018(),
		"sDevice":   params.Promotion,
	}
	
	headers := map[string]string{
		"deviceName": deviceName,
	}
	
	response, err := ec.withMetrics("GetMyInfo", func() (*APIResponse, error) {
		return ec.httpManager.Post(ctx, "/my", body, deviceID, headers)
	})
	
	if err != nil {
		errorMsg := fmt.Sprintf("Error, please contact cursor-vip@jeter.eu.org:\n%v", err)
		_, _ = fmt.Fprintf(params.ColorOut, params.Red, errorMsg)
		_, _ = fmt.Scanln()
		panic(fmt.Sprintf("\u001B[31m%s\u001B[0m", err))
	}
	
	if errorField := response.GetString("error"); errorField != "" {
		_, _ = fmt.Fprintf(params.ColorOut, params.Red, "Error, please contact cursor-vip@jeter.eu.org:\n"+errorField)
		_, _ = fmt.Scanln()
		panic(fmt.Sprintf("\u001B[31m%s\u001B[0m", errorField))
	}
	
	return response.GetString("sCount"),
		response.GetString("sPayCount"),
		response.GetString("isPay"),
		response.GetString("ticket"),
		response.GetString("exp"),
		response.GetString("exclusiveAt"),
		response.GetString("token"),
		response.GetString("m3c"),
		response.GetString("msg")
}

func (ec *EnhancedClient) CheckVersion(version string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	path := fmt.Sprintf("/version?version=%s&plat=%s_%s", version, runtime.GOOS, runtime.GOARCH)
	response, err := ec.withMetrics("CheckVersion", func() (*APIResponse, error) {
		return ec.httpManager.Get(ctx, path, "", nil)
	})
	
	if err != nil {
		params.GlobalState.SetLastError(fmt.Errorf("failed to check version: %w", err))
		return ""
	}
	
	return response.GetString("url")
}

func (ec *EnhancedClient) GetLic() (isOk bool, result string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	path := fmt.Sprintf("/getLic?mode=%d", params.Mode)
	response, err := ec.withMetrics("GetLic", func() (*APIResponse, error) {
		return ec.httpManager.Get(ctx, path, params.DeviceID, nil)
	})
	
	if err != nil {
		params.GlobalState.SetLastError(fmt.Errorf("failed to get license: %w", err))
		return false, err.Error()
	}
	
	code := response.GetInt("code")
	result = response.GetString("lic")
	isOk = code == 0
	
	return isOk, result
}

// Get client metrics
func (ec *EnhancedClient) GetMetrics() *ClientMetrics {
	return clientMetrics
}

// Health check
func (ec *EnhancedClient) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return ec.httpManager.HealthCheck(ctx)
}