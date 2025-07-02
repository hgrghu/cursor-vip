package client

import (
	"context"
	"fmt"
	"time"
	
	"github.com/kingparks/cursor-vip/tui/params"
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
	// 返回开源版本信息
	return "🎉 Cursor VIP 开源版本 - 完全免费使用！"
}

// 简化的用户信息获取，去掉支付相关信息
func (ec *EnhancedClient) GetMyInfo(deviceID string) (sCount, sPayCount, isPay, ticket, exp, exclusiveAt, token, m3c, msg string) {
	// 返回虚拟的已授权信息，表示永久有效
	currentTime := time.Now()
	futureTime := currentTime.AddDate(10, 0, 0) // 添加10年，表示永久有效
	
	return "0",                                      // sCount
		"0",                                         // sPayCount  
		"true",                                      // isPay
		"open-source-ticket",                        // ticket
		futureTime.Format("2006-01-02 15:04:05"),   // exp (10年后过期)
		"",                                          // exclusiveAt
		"",                                          // token
		"∞",                                         // m3c (无限)
		"🎉 开源版本永久免费！感谢使用！"                     // msg
}

func (ec *EnhancedClient) CheckVersion(version string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	path := fmt.Sprintf("/version?version=%s&plat=%s_%s", version, "linux", "amd64")
	response, err := ec.withMetrics("CheckVersion", func() (*APIResponse, error) {
		return ec.httpManager.Get(ctx, path, "", nil)
	})
	
	if err != nil {
		params.GlobalState.SetLastError(fmt.Errorf("failed to check version: %w", err))
		return ""
	}
	
	return response.GetString("url")
}

// 简化的许可证获取，直接返回成功
func (ec *EnhancedClient) GetLic() (isOk bool, result string) {
	// 开源版本直接返回成功
	return true, "open-source-license-valid"
}

// 删除的支付相关方法（已移除）：
/*
func (ec *EnhancedClient) GetPayUrl() (payUrl, orderID string)
func (ec *EnhancedClient) GetExclusivePayUrl() (payUrl, orderID string)
func (ec *EnhancedClient) getPaymentURL(endpoint string) (payUrl, orderID string)
func (ec *EnhancedClient) GetM3PayUrl() (payUrl, orderID string)
func (ec *EnhancedClient) GetM3tPayUrl() (payUrl, orderID string)
func (ec *EnhancedClient) GetM3hPayUrl() (payUrl, orderID string)
func (ec *EnhancedClient) checkPayment(endpoint, orderID, deviceID string) bool
func (ec *EnhancedClient) PayCheck(orderID, deviceID string) bool
func (ec *EnhancedClient) ExclusivePayCheck(orderID, deviceID string) bool
func (ec *EnhancedClient) M3PayCheck(orderID, deviceID string) bool
func (ec *EnhancedClient) M3tPayCheck(orderID, deviceID string) bool
func (ec *EnhancedClient) M3hPayCheck(orderID, deviceID string) bool
*/

// 保留的功能性方法
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
	// 开源版本默认返回有效
	return true
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