package main

import (
	"log"

	"github.com/shaibs3/Torq/internal/app"
	"github.com/shaibs3/Torq/internal/config"
	"github.com/shaibs3/Torq/internal/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger first (for configuration loading)
	initialLogger, err := logger.New("production", "info")
	if err != nil {
		log.Fatal("failed to initialize logger:", err)
	}
	defer func() {
		_ = initialLogger.Sync()
	}()

	// Load configuration
	cfg := config.Load(initialLogger)

	// Create application logger with proper configuration
	appLogger, err := logger.New(cfg.Environment, cfg.LogLevel)
	if err != nil {
		initialLogger.Fatal("failed to create application logger", zap.Error(err))
	}
	defer func() {
		_ = appLogger.Sync()
	}()

	// Create and run application
	application, err := app.New(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("failed to create application", zap.Error(err))
	}

	// Run the application
	if err := application.Run(); err != nil {
		appLogger.Fatal("application failed", zap.Error(err))
	}
}
