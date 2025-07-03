@echo off
chcp 65001 > nul
title Cursor VIP 开源免费版

echo.
echo ================================================
echo     🎉 Cursor VIP 开源免费版启动器
echo ================================================
echo.
echo 📋 正在启动Cursor VIP服务...
echo 💡 提示：请保持此窗口开启以维持VIP功能
echo.

REM 检查是否存在64位版本
if exist "cursor-vip.exe" (
    echo ✅ 启动64位版本...
    cursor-vip.exe
) else if exist "cursor-vip-x86.exe" (
    echo ✅ 启动32位版本...
    cursor-vip-x86.exe
) else (
    echo ❌ 错误：未找到可执行文件！
    echo 请确保 cursor-vip.exe 或 cursor-vip-x86.exe 存在
    pause
    exit /b 1
)

echo.
echo 程序已退出，按任意键关闭窗口...
pause > nul