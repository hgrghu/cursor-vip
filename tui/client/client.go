package client

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/kingparks/cursor-vip/auth/sign"
	"github.com/kingparks/cursor-vip/tui/params"
	"github.com/kingparks/cursor-vip/tui/tool"
	"github.com/tidwall/gjson"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"runtime"
	"time"
)

var Cli Client

type Client struct {
	Hosts []string // 服务器地址s
	host  string   // 检查后的服务器地址
}

func (c *Client) SetProxy(lang string) {
	defer c.setHost()
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
	httplib.SetDefaultSetting(httplib.BeegoHTTPSettings{
		Proxy:            proxy,
		ReadWriteTimeout: 30 * time.Second,
		ConnectTimeout:   30 * time.Second,
		Gzip:             true,
		DumpBody:         true,
		UserAgent: fmt.Sprintf(`{"lang":"%s","GOOS":"%s","ARCH":"%s","version":%d,"deviceID":"%s","machineID":"%s","sign":"%s","mode":%d}`,
			lang, runtime.GOOS, runtime.GOARCH, params.Version, params.DeviceID, params.MachineID, sign.Sign(params.DeviceID), params.Mode),
	})
	if len(proxyText) > 0 {
		_, _ = fmt.Fprintf(params.ColorOut, params.Yellow, proxyText)
	}
}

func (c *Client) setHost() {
	c.host = c.Hosts[0]
	for _, v := range c.Hosts {
		_, err := httplib.Get(v).SetTimeout(4*time.Second, 4*time.Second).String()
		if err == nil {
			c.host = v
			return
		}
	}
	return
}

func (c *Client) GetAD() (ad string) {
	// 简化广告功能，返回开源信息
	return "🎉 Cursor VIP 开源版本 - 完全免费使用！"
}

// 简化的用户信息获取，去掉支付相关信息
func (c *Client) GetMyInfo(deviceID string) (sCount, sPayCount, isPay, ticket, exp, exclusiveAt, token, m3c, msg string) {
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

func (c *Client) CheckVersion(version string) (upUrl string) {
	res, err := httplib.Get(c.host + "/version?version=" + version + "&plat=" + runtime.GOOS + "_" + runtime.GOARCH).String()
	if err != nil {
		return ""
	}
	upUrl = gjson.Get(res, "url").String()
	return
}

// 简化的许可证获取，直接返回成功
func (c *Client) GetLic() (isOk bool, result string) {
	// 开源版本直接返回成功
	return true, "open-source-license-valid"
}

// 删除的支付相关方法（已注释，实际删除）：
/*
func (c *Client) GetPayUrl() (payUrl, orderID string)
func (c *Client) GetExclusivePayUrl() (payUrl, orderID string)  
func (c *Client) GetM3PayUrl() (payUrl, orderID string)
func (c *Client) GetM3tPayUrl() (payUrl, orderID string)
func (c *Client) GetM3hPayUrl() (payUrl, orderID string)
func (c *Client) PayCheck(orderID, deviceID string) (isPay bool)
func (c *Client) ExclusivePayCheck(orderID, deviceID string) (isPay bool)
func (c *Client) M3PayCheck(orderID, deviceID string) (isPay bool)
func (c *Client) M3tPayCheck(orderID, deviceID string) (isPay bool)
func (c *Client) M3hPayCheck(orderID, deviceID string) (isPay bool)
*/

// 保留的功能性方法
func (c *Client) DelFToken(deviceID, category string) (err error) {
	_, err = httplib.Delete(c.host+"/delFToken?category="+category).Header("sign", sign.Sign(deviceID)).String()
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (c *Client) CheckFToken(deviceID string) (has bool) {
	// 开源版本默认返回有效
	return true
}

func (c *Client) UpExclusiveStatus(exclusiveUsed, exclusiveTotal int64, exclusiveErr, exclusiveToken, deviceID string) {
	body, _ := json.Marshal(map[string]interface{}{
		"exclusiveUsed":  exclusiveUsed,
		"exclusiveTotal": exclusiveTotal,
		"exclusiveErr":   exclusiveErr,
		"exclusiveToken": exclusiveToken,
	})
	_, _ = httplib.Post(c.host+"/upExclusiveStatus").
		Header("sign", sign.Sign(deviceID)).
		Body(body).
		String()
	return
}

func (c *Client) UpChecksumPrefix(p, deviceID string) {
	body, _ := json.Marshal(map[string]interface{}{"p": p})
	_, _ = httplib.Post(c.host+"/upChecksumPrefix").
		Header("sign", sign.Sign(deviceID)).
		Body(body).
		String()
	return
}
