package main

import (
	"github.com/gofrs/flock"
	"github.com/kingparks/cursor-vip/auth"
	"github.com/kingparks/cursor-vip/tui"
	"github.com/kingparks/cursor-vip/tui/client"
	"github.com/kingparks/cursor-vip/tui/logger"
	"github.com/kingparks/cursor-vip/tui/params"
	"github.com/kingparks/cursor-vip/tui/shortcut"
	"github.com/kingparks/cursor-vip/tui/tool"
	"github.com/kingparks/cursor-vip/tui/ui"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var lock *flock.Flock
var pidFilePath string

func main() {
	// Initialize enhanced components
	if err := initializeApplication(); err != nil {
		logger.Fatal("Failed to initialize application: %v", err)
	}
	
	// Ensure single instance
	var err error
	lock, pidFilePath, err = tool.EnsureSingleInstance("cursor-vip")
	if err != nil {
		logger.Fatal("Failed to ensure single instance: %v", err)
	}
	
	// Set up graceful shutdown
	setupGracefulShutdown()
	
	// Mark application as running
	params.GlobalState.SetRunning(true)
	defer params.GlobalState.SetRunning(false)
	
	// Run TUI and get selections
	productSelected, modelIndexSelected := tui.Run()
	
	// Start server with enhanced error handling
	startServer(productSelected, modelIndexSelected)
}

// Initialize all application components
func initializeApplication() error {
	// Initialize logger first
	if err := logger.InitDefault(); err != nil {
		return err
	}
	logger.Info("Application starting up")
	
	// Initialize UI manager
	ui.InitUI()
	
	// Initialize enhanced HTTP client
	config, err := tool.GetEnhancedConfig()
	if err != nil {
		logger.Warn("Failed to load enhanced config, using defaults: %v", err)
	}
	
	// Initialize client with hosts from params
	client.InitEnhancedClient(params.Hosts)
	
	logger.Info("Application initialized successfully")
	return nil
}

// Set up graceful shutdown handling
func setupGracefulShutdown() {
	params.Sigs = make(chan os.Signal, 1)
	signal.Notify(params.Sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGKILL)
	
	go func() {
		sig := <-params.Sigs
		logger.Info("Received signal: %v, shutting down gracefully", sig)
		
		// Cleanup resources
		cleanup()
		
		logger.Info("Application shut down complete")
		logger.Close()
		os.Exit(0)
	}()
}

// Cleanup resources
func cleanup() {
	// Unlock file lock
	if lock != nil {
		_ = lock.Unlock()
	}
	
	// Remove PID file
	if pidFilePath != "" {
		_ = os.Remove(pidFilePath)
	}
	
	// Unset client if needed
	// This would need to be implemented based on the auth package
	// auth.UnSetClient(productSelected)
	
	// Mark application as not running
	params.GlobalState.SetRunning(false)
}

func startServer(productSelected string, modelIndexSelected int) {
	logger.InfoWithFields("Starting server", logger.Fields{
		"product": productSelected,
		"model":   modelIndexSelected,
	})
	
	// Set up signal handling for countdown
	params.SigCountDown = make(chan int, 1)
	
	// Start shortcut handler
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Shortcut handler panicked: %v", r)
			}
		}()
		shortcut.Do()
	}()
	
	// Start authentication service with error recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Auth service panicked: %v", r)
				params.GlobalState.SetLastError(fmt.Errorf("auth service crashed: %v", r))
			}
		}()
		
		// Run auth service
		auth.Run(productSelected, modelIndexSelected)
	}()
	
	// Wait for shutdown signal
	<-params.Sigs
}
