package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
	
	"github.com/mattn/go-colorable"
)

// UI components and utilities
type UIManager struct {
	output    io.Writer
	width     int
	height    int
	isLoading bool
	loadingMu sync.Mutex
}

// Progress bar configuration
type ProgressConfig struct {
	Width      int
	Character  string
	Background string
	ShowPercent bool
	ShowTime    bool
}

// Spinner configuration
type SpinnerConfig struct {
	Frames []string
	Delay  time.Duration
}

// Common spinners
var (
	DotSpinner = SpinnerConfig{
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		Delay:  100 * time.Millisecond,
	}
	
	ArrowSpinner = SpinnerConfig{
		Frames: []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
		Delay:  150 * time.Millisecond,
	}
	
	LoadingSpinner = SpinnerConfig{
		Frames: []string{"◐", "◓", "◑", "◒"},
		Delay:  200 * time.Millisecond,
	}
)

// Global UI manager
var GlobalUI *UIManager

// Initialize UI manager
func InitUI() {
	GlobalUI = &UIManager{
		output: colorable.NewColorableStdout(),
		width:  80, // Default width
		height: 24, // Default height
	}
}

// Enhanced print functions with better formatting
func (ui *UIManager) PrintTitle(title string) {
	border := strings.Repeat("═", len(title)+4)
	_, _ = fmt.Fprintf(ui.output, "\033[1;36m╔%s╗\033[0m\n", border)
	_, _ = fmt.Fprintf(ui.output, "\033[1;36m║  %s  ║\033[0m\n", title)
	_, _ = fmt.Fprintf(ui.output, "\033[1;36m╚%s╝\033[0m\n", border)
}

func (ui *UIManager) PrintSection(section string) {
	_, _ = fmt.Fprintf(ui.output, "\033[1;33m▶ %s\033[0m\n", section)
}

func (ui *UIManager) PrintSuccess(message string) {
	_, _ = fmt.Fprintf(ui.output, "\033[1;32m✓ %s\033[0m\n", message)
}

func (ui *UIManager) PrintWarning(message string) {
	_, _ = fmt.Fprintf(ui.output, "\033[1;33m⚠ %s\033[0m\n", message)
}

func (ui *UIManager) PrintError(message string) {
	_, _ = fmt.Fprintf(ui.output, "\033[1;31m✗ %s\033[0m\n", message)
}

func (ui *UIManager) PrintInfo(message string) {
	_, _ = fmt.Fprintf(ui.output, "\033[1;34mℹ %s\033[0m\n", message)
}

// Progress bar implementation
func (ui *UIManager) ShowProgress(current, total int, config *ProgressConfig) {
	if config == nil {
		config = &ProgressConfig{
			Width:       50,
			Character:   "█",
			Background:  "░",
			ShowPercent: true,
			ShowTime:    false,
		}
	}
	
	percentage := float64(current) / float64(total)
	filled := int(percentage * float64(config.Width))
	
	bar := strings.Repeat(config.Character, filled) + 
		  strings.Repeat(config.Background, config.Width-filled)
	
	output := fmt.Sprintf("\r[%s]", bar)
	
	if config.ShowPercent {
		output += fmt.Sprintf(" %.1f%%", percentage*100)
	}
	
	_, _ = fmt.Fprint(ui.output, output)
}

// Spinner with loading message
func (ui *UIManager) ShowSpinner(message string, config SpinnerConfig, done <-chan bool) {
	ui.loadingMu.Lock()
	ui.isLoading = true
	ui.loadingMu.Unlock()
	
	ticker := time.NewTicker(config.Delay)
	defer ticker.Stop()
	
	frame := 0
	for {
		select {
		case <-done:
			ui.loadingMu.Lock()
			ui.isLoading = false
			ui.loadingMu.Unlock()
			
			// Clear the spinner line
			fmt.Fprint(ui.output, "\r\033[K")
			return
		case <-ticker.C:
			fmt.Fprintf(ui.output, "\r%s %s", config.Frames[frame], message)
			frame = (frame + 1) % len(config.Frames)
		}
	}
}

// Enhanced countdown with better formatting
func (ui *UIManager) ShowCountdown(seconds int, message string) {
	for countdown := seconds; countdown >= 0; countdown-- {
		days := countdown / (24 * 3600)
		hours := (countdown % (24 * 3600)) / 3600
		minutes := (countdown % 3600) / 60
		secs := countdown % 60
		
		timeStr := ""
		if days > 0 {
			timeStr = fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, secs)
		} else if hours > 0 {
			timeStr = fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
		} else if minutes > 0 {
			timeStr = fmt.Sprintf("%dm %ds", minutes, secs)
		} else {
			timeStr = fmt.Sprintf("%ds", secs)
		}
		
		fmt.Fprintf(ui.output, "\r%s %s", message, timeStr)
		time.Sleep(1 * time.Second)
	}
	fmt.Fprint(ui.output, "\n")
}

// Menu selection with better formatting
func (ui *UIManager) ShowMenu(title string, options []string, defaultOption int) int {
	ui.PrintSection(title)
	
	for i, option := range options {
		marker := " "
		if i+1 == defaultOption {
			marker = "▶"
		}
		_, _ = fmt.Fprintf(ui.output, "\033[32m%s %d. %s\033[0m\n", marker, i+1, option)
	}
	
	fmt.Print("\n选择选项 (默认: ", defaultOption, "): ")
	var choice int
	_, _ = fmt.Scanln(&choice)
	
	if choice < 1 || choice > len(options) {
		choice = defaultOption
	}
	
	return choice
}

// Enhanced status display
func (ui *UIManager) ShowStatus(items map[string]string) {
	ui.PrintSection("系统状态")
	
	maxKeyLen := 0
	for key := range items {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}
	
	for key, value := range items {
		padding := strings.Repeat(" ", maxKeyLen-len(key))
		_, _ = fmt.Fprintf(ui.output, "\033[36m%s%s:\033[0m %s\n", key, padding, value)
	}
}

// Box drawing for important messages
func (ui *UIManager) ShowBox(title, content string) {
	lines := strings.Split(content, "\n")
	maxLen := len(title)
	
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	
	width := maxLen + 4
	
	// Top border
	_, _ = fmt.Fprintf(ui.output, "\033[1;36m╔%s╗\033[0m\n", strings.Repeat("═", width-2))
	
	// Title
	titlePadding := (width - len(title) - 2) / 2
	_, _ = fmt.Fprintf(ui.output, "\033[1;36m║%s%s%s║\033[0m\n", 
		strings.Repeat(" ", titlePadding), 
		title, 
		strings.Repeat(" ", width-2-titlePadding-len(title)))
	
	// Separator
	_, _ = fmt.Fprintf(ui.output, "\033[1;36m╠%s╣\033[0m\n", strings.Repeat("═", width-2))
	
	// Content
	for _, line := range lines {
		padding := width - len(line) - 3
		_, _ = fmt.Fprintf(ui.output, "\033[1;36m║\033[0m %s%s\033[1;36m║\033[0m\n", 
			line, strings.Repeat(" ", padding))
	}
	
	// Bottom border
	_, _ = fmt.Fprintf(ui.output, "\033[1;36m╚%s╝\033[0m\n", strings.Repeat("═", width-2))
}

// Enhanced input validation
func (ui *UIManager) GetValidatedInput(prompt string, validator func(string) error) string {
	for {
		fmt.Print(prompt)
		var input string
		_, _ = fmt.Scanln(&input)
		
		if err := validator(input); err != nil {
			ui.PrintError(err.Error())
			continue
		}
		
		return input
	}
}

// Confirmation dialog
func (ui *UIManager) Confirm(message string, defaultYes bool) bool {
	defaultText := "y/N"
	if defaultYes {
		defaultText = "Y/n"
	}
	
	fmt.Printf("%s (%s): ", message, defaultText)
	var response string
	_, _ = fmt.Scanln(&response)
	
	if response == "" {
		return defaultYes
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// Loading with timeout
func (ui *UIManager) ShowLoadingWithTimeout(message string, timeout time.Duration) bool {
	done := make(chan bool, 1)
	
	// Start spinner
	go ui.ShowSpinner(message, DotSpinner, done)
	
	// Wait for timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	
	select {
	case <-timer.C:
		done <- true
		ui.PrintWarning("操作超时")
		return false
	default:
		time.Sleep(100 * time.Millisecond) // Simulate work
		done <- true
		ui.PrintSuccess("操作完成")
		return true
	}
}

// Clear screen
func (ui *UIManager) Clear() {
	fmt.Print("\033[2J\033[H")
}

// Move cursor
func (ui *UIManager) MoveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

// Hide/Show cursor
func (ui *UIManager) HideCursor() {
	fmt.Print("\033[?25l")
}

func (ui *UIManager) ShowCursor() {
	fmt.Print("\033[?25h")
}

// Enhanced error display with suggestions
func (ui *UIManager) ShowErrorWithSuggestions(err error, suggestions []string) {
	ui.PrintError(err.Error())
	
	if len(suggestions) > 0 {
		fmt.Println("\n建议:")
		for i, suggestion := range suggestions {
			_, _ = fmt.Fprintf(ui.output, "\033[33m  %d. %s\033[0m\n", i+1, suggestion)
		}
	}
}

// Global UI functions for backward compatibility
func PrintTitle(title string) {
	if GlobalUI != nil {
		GlobalUI.PrintTitle(title)
	}
}

func PrintSuccess(message string) {
	if GlobalUI != nil {
		GlobalUI.PrintSuccess(message)
	}
}

func PrintError(message string) {
	if GlobalUI != nil {
		GlobalUI.PrintError(message)
	}
}

func PrintWarning(message string) {
	if GlobalUI != nil {
		GlobalUI.PrintWarning(message)
	}
}

func PrintInfo(message string) {
	if GlobalUI != nil {
		GlobalUI.PrintInfo(message)
	}
}

func ShowProgress(current, total int) {
	if GlobalUI != nil {
		GlobalUI.ShowProgress(current, total, nil)
	}
}

func Confirm(message string) bool {
	if GlobalUI != nil {
		return GlobalUI.Confirm(message, false)
	}
	return false
}