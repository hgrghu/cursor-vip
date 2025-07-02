package tui

import (
	"embed"
	"fmt"
	"github.com/kingparks/cursor-vip/tui/client"
	"github.com/kingparks/cursor-vip/tui/params"
	"github.com/kingparks/cursor-vip/tui/tool"
	"github.com/mattn/go-colorable"
	"os/signal"
	"syscall"

	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/unknwon/i18n"
)

//go:embed all:locales
var localeFS embed.FS

// Run 启动
func Run() (productSelected string, modelIndexSelected int) {
	params.ColorOut = colorable.NewColorableStdout()
	params.Lang, params.Promotion, params.Mode = tool.GetConfig()
	params.DeviceID = tool.GetMachineID()
	params.MachineID = tool.GetMachineID()
	client.Cli = client.Client{Hosts: params.Hosts}

	localeFileEn, _ := localeFS.ReadFile("locales/en.ini")
	_ = i18n.SetMessage("en", localeFileEn)
	localeFileNl, _ := localeFS.ReadFile("locales/nl.ini")
	_ = i18n.SetMessage("nl", localeFileNl)
	localeFileRu, _ := localeFS.ReadFile("locales/ru.ini")
	_ = i18n.SetMessage("ru", localeFileRu)
	localeFileHu, _ := localeFS.ReadFile("locales/hu.ini")
	_ = i18n.SetMessage("hu", localeFileHu)
	localeFileTr, _ := localeFS.ReadFile("locales/tr.ini")
	_ = i18n.SetMessage("tr", localeFileTr)
	localeFileEs, _ := localeFS.ReadFile("locales/es.ini")
	_ = i18n.SetMessage("es", localeFileEs)
	switch params.Lang {
	case "zh":
		params.Trr = &params.Tr{Locale: i18n.Locale{Lang: "zh"}}
		params.GithubPath = strings.ReplaceAll(params.GithubPath, "https://github.com", "https://gitee.com")
		params.GithubInstall = "ic.sh"
	case "nl":
		params.Trr = &params.Tr{Locale: i18n.Locale{Lang: "nl"}}
	case "ru":
		params.Trr = &params.Tr{Locale: i18n.Locale{Lang: "ru"}}
	case "hu":
		params.Trr = &params.Tr{Locale: i18n.Locale{Lang: "hu"}}
	case "tr":
		params.Trr = &params.Tr{Locale: i18n.Locale{Lang: "tr"}}
	case "es":
		params.Trr = &params.Tr{Locale: i18n.Locale{Lang: "es"}}
	default:
		params.Trr = &params.Tr{Locale: i18n.Locale{Lang: "en"}}
	}

	_, _ = fmt.Fprintf(params.ColorOut, params.Green, params.Trr.Tr("CURSOR VIP")+` v`+strings.Join(strings.Split(fmt.Sprint(params.Version), ""), "."))
	
	// 检查是否在容器环境
	if content, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		if strings.Contains(string(content), "/docker/") {
			_, _ = fmt.Fprintf(params.ColorOut, params.Red, params.Trr.Tr("不支持容器环境"))
			_, _ = fmt.Scanln()
			// 发送退出信号
			params.Sigs <- syscall.SIGTERM
			panic(params.Trr.Tr("不支持容器环境"))
		}
	}
	
	client.Cli.SetProxy(params.Lang)
	_, _ = fmt.Fprintf(params.ColorOut, params.Green, params.Trr.Tr("设备码")+":"+params.DeviceID)
	_, _ = fmt.Fprintf(params.ColorOut, params.Green, params.Trr.Tr("当前模式")+": "+fmt.Sprint(params.Mode))
	
	// 显示开源版本信息
	_, _ = fmt.Fprintf(params.ColorOut, params.HGreen, "🎉 "+params.Trr.Tr("开源免费版本，无需付费！"))
	_, _ = fmt.Fprintf(params.ColorOut, params.Green, "📧 "+params.Trr.Tr("项目地址")+"：https://github.com/kingparks/cursor-vip")
	_, _ = fmt.Fprintf(params.ColorOut, params.Green, "⭐ "+params.Trr.Tr("如果觉得有用，请给项目点个星！"))
	fmt.Println()

	printAD()
	fmt.Println()
	checkUpdate(params.Version)

	// 快捷键
	_, _ = fmt.Fprintf(params.ColorOut, params.Green, params.Trr.Tr("Switch to English：Press 'sen' on keyboard in turn"))
	modelIndexSelected = int(params.Mode)

	fmt.Println()

	// 产品选择
	if len(params.Product) > 1 {
		_, _ = fmt.Fprintf(params.ColorOut, params.DefaultColor, params.Trr.Tr("选择要授权的产品："))
		for i, v := range params.Product {
			_, _ = fmt.Fprintf(params.ColorOut, params.HGreen, fmt.Sprintf("%d. %s\t", i+1, v))
		}
		fmt.Println()
		fmt.Print(params.Trr.Tr("请输入产品编号（直接回车默认为1，可以同时输入多个例如 145）："))
		productIndex := 1
		_, _ = fmt.Scanln(&productIndex)
		if productIndex < 1 {
			fmt.Println(params.Trr.Tr("输入有误"))
			return
		}
		for _, v := range strings.Split(fmt.Sprint(productIndex), "") {
			vi, _ := strconv.Atoi(v)
			productSelected += params.Product[vi-1] + ","
		}
		if len(productSelected) > 1 {
			productSelected = productSelected[:len(productSelected)-1]
		}
		fmt.Println(params.Trr.Tr("选择的产品为：") + productSelected)
		fmt.Println()
	} else {
		productSelected = params.Product[0]
	}

	// 直接授权成功，无需付费验证
	_, _ = fmt.Fprintf(params.ColorOut, params.Green, "✅ "+params.Trr.Tr("授权成功！开源版本永久免费使用"))
	_, _ = fmt.Fprintf(params.ColorOut, params.Green, "🚀 "+params.Trr.Tr("正在启动服务，请保持此窗口开启..."))
	fmt.Println()

	// 启动倒计时（设置为一年，实际上是永久）
	go func() {
		params.SigCountDown = make(chan int, 1)
		<-params.SigCountDown
		_, _ = fmt.Fprintf(params.ColorOut, params.Green, params.Trr.Tr("服务运行中，使用过程请不要关闭此窗口"))
		// 设置一个很长的时间，表示永久授权
		tool.CountDown(365 * 24 * 3600) // 一年时间
	}()
	
	return
}

func printAD() {
	// 简化广告显示，只显示开源项目信息
	_, _ = fmt.Fprintf(params.ColorOut, params.Yellow, "📢 "+params.Trr.Tr("感谢使用 Cursor VIP 开源版本！"))
}

func checkUpdate(version int) {
	// 保留版本检查功能
	upUrl := client.Cli.CheckVersion(fmt.Sprint(version))
	if upUrl == "" {
		return
	}
	
	installCmd := `bash -c "$(curl -fsSLk ` + params.GithubPath + params.GithubDownLoadPath + params.GithubInstall + `)"`
	
	switch runtime.GOOS {
	case "windows":
		_, _ = fmt.Fprintf(params.ColorOut, params.Red, params.Trr.Tr("有新版本，请关闭本窗口，将下面命令粘贴到GitBash窗口执行")+`：`)
	default:
		_, _ = fmt.Fprintf(params.ColorOut, params.Red, params.Trr.Tr("有新版本，请关闭本窗口，将下面命令粘贴到新终端窗口执行")+`：`)
	}
	_, _ = fmt.Fprintf(params.ColorOut, params.HGreen, installCmd)
	fmt.Println()

	// 捕获 Ctrl+C 信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		params.Sigs <- syscall.SIGTERM
	}()

	_, _ = fmt.Scanln()
	params.Sigs <- syscall.SIGTERM
	return
}
