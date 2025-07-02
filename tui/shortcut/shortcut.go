package shortcut

import (
	"fmt"
	"github.com/eiannone/keyboard"
	"github.com/kingparks/cursor-vip/authtool"
	"github.com/kingparks/cursor-vip/tui/params"
	"github.com/kingparks/cursor-vip/tui/tool"
	"os"
	"runtime"
	"strings"
	"syscall"
)

func Do() {
	if err := keyboard.Open(); err != nil {
		fmt.Println("Failed to initialize keyboard:", err)
		return
	}
	defer keyboard.Close()

	var keyBuffer []rune
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			//fmt.Println("Error reading keyboard:", err)
			return
		}

		//// 检查是否按下 Ctrl+C
		if key == keyboard.KeyCtrlC {
			// 发送退出信号
			params.Sigs <- syscall.SIGTERM
			keyboard.Close()
		}
		// 判断是否按下回车键
		if key == keyboard.KeyEnter {
			//break
		}

		// 将按键添加到缓冲区
		if char != 0 {
			keyBuffer = append(keyBuffer, char)
		}

		// 保持缓冲区最多3个字符
		if len(keyBuffer) > 3 {
			keyBuffer = keyBuffer[1:]
		}

		// 检查快捷键组合
		combination := string(keyBuffer)

		switch {
		case strings.HasSuffix(combination, "sen"):
			params.Lang = "en"
			tool.SetConfig(params.Lang, params.Mode)
			fmt.Println()
			_, _ = fmt.Fprintf(params.ColorOut, params.Red, params.Trr.Tr("Settings successful, will take effect after restart"))
			keyBuffer = nil
			tool.OpenNewTerminal()

		case strings.HasSuffix(combination, "szh"):
			params.Lang = "zh"
			tool.SetConfig(params.Lang, params.Mode)
			fmt.Println()
			_, _ = fmt.Fprintf(params.ColorOut, params.Red, params.Trr.Tr("Settings successful, will take effect after restart"))
			keyBuffer = nil
			tool.OpenNewTerminal()

		case strings.HasSuffix(combination, "sm1"):
			params.Mode = 1
			tool.SetConfig(params.Lang, params.Mode)
			fmt.Println()
			_, _ = fmt.Fprintf(params.ColorOut, params.Red, params.Trr.Tr("设置成功，将在重启 cursor-vip 后生效"))
			keyBuffer = nil
			tool.OpenNewTerminal()

		case strings.HasSuffix(combination, "sm2"):
			has := check45Version()
			if has {
				params.Mode = 2
				tool.SetConfig(params.Lang, params.Mode)
				fmt.Println()
				_, _ = fmt.Fprintf(params.ColorOut, params.Red, params.Trr.Tr("设置成功，将在重启 cursor-vip 后生效"))
				keyBuffer = nil
				tool.OpenNewTerminal()
			} else {
				keyBuffer = nil
			}

		case strings.HasSuffix(combination, "sm3"):
			params.Mode = 3
			tool.SetConfig(params.Lang, params.Mode)
			fmt.Println()
			_, _ = fmt.Fprintf(params.ColorOut, params.Red, params.Trr.Tr("设置成功，将在重启 cursor-vip 后生效"))
			keyBuffer = nil
			tool.OpenNewTerminal()

		case strings.HasSuffix(combination, "sm4"):
			params.Mode = 4
			tool.SetConfig(params.Lang, params.Mode)
			fmt.Println()
			_, _ = fmt.Fprintf(params.ColorOut, params.Red, params.Trr.Tr("设置成功，将在重启 cursor-vip 后生效"))
			keyBuffer = nil
			tool.OpenNewTerminal()

		case strings.HasSuffix(combination, "ver"):
			// 显示版本信息
			fmt.Println()
			_, _ = fmt.Fprintf(params.ColorOut, params.Green, "🎉 Cursor VIP 开源版本 v%s", strings.Join(strings.Split(fmt.Sprint(params.Version), ""), "."))
			_, _ = fmt.Fprintf(params.ColorOut, params.Green, "📧 项目地址：https://github.com/kingparks/cursor-vip")
			_, _ = fmt.Fprintf(params.ColorOut, params.Green, "⭐ 如果觉得有用，请给项目点个星！")
			keyBuffer = nil

		case strings.HasSuffix(combination, "hlp"):
			// 显示帮助信息
			fmt.Println()
			_, _ = fmt.Fprintf(params.ColorOut, params.Yellow, "🔧 快捷键帮助：")
			_, _ = fmt.Fprintf(params.ColorOut, params.Green, "sen - 切换到英文")
			_, _ = fmt.Fprintf(params.ColorOut, params.Green, "szh - 切换到中文")
			_, _ = fmt.Fprintf(params.ColorOut, params.Green, "sm1 - 切换到模式1")
			_, _ = fmt.Fprintf(params.ColorOut, params.Green, "sm2 - 切换到模式2")
			_, _ = fmt.Fprintf(params.ColorOut, params.Green, "sm3 - 切换到模式3")
			_, _ = fmt.Fprintf(params.ColorOut, params.Green, "sm4 - 切换到模式4")
			_, _ = fmt.Fprintf(params.ColorOut, params.Green, "ver - 显示版本信息")
			_, _ = fmt.Fprintf(params.ColorOut, params.Green, "hlp - 显示此帮助")
			keyBuffer = nil
		}
		
		// 移除的支付相关快捷键：
		// buy, u3d, u3t, u3h, ckp, c3p, c3t, c3h, q3d
	}
}

func check45Version() bool {
	if strings.HasPrefix(authtool.GetCursorVersion(), "0.45.") {
		return true
	}
	switch runtime.GOOS {
	case "darwin":
		_, err := os.Stat("/Applications/Cursor.45.app/Contents/Resources/app/bin/code")
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintf(params.ColorOut, params.Green, "0.45 cursor 下载地址："+"https://github.com/oslook/cursor-ai-downloads#:~:text=x64%0Alinux%2Darm64-,0.45.15,-2025%2D02%2D20")
			_, _ = fmt.Fprintf(params.ColorOut, params.Red, "请先覆盖安装0.45的cursor,然后执行 mv /Applications/Cursor.app /Applications/Cursor.45.app;")
			return false
		} else {
			return true
		}
	case "linux":
		//todo 0.45版本
	}
	return false
}
