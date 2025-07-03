package app

import (
	"context"
	"errors"
	"github.com/shaibs3/Torq/internal/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shaibs3/Torq/internal/limiter"

	"github.com/shaibs3/Torq/internal/config"
	"github.com/shaibs3/Torq/internal/finder"
	"github.com/shaibs3/Torq/internal/lookup"
	"github.com/shaibs3/Torq/internal/telemetry"
	"go.uber.org/zap"
)

// App represents the main application
type App struct {
	config    *config.Config
	logger    *zap.Logger
	telemetry *telemetry.Telemetry
	server    *http.Server
}

func NewApp(cfg *config.Config, logger *zap.Logger) (*App, error) {
	// Initialize telemetry
	tel, err := telemetry.NewTelemetry(logger)
	if err != nil {
		return nil, err
	}

	// Initialize IP DB provider
	dbProviderFactory := lookup.NewDbProviderFactory(logger.Named("db_provider"))
	dbProvider, err := dbProviderFactory.CreateProvider(cfg.IPDBConfig)
	if err != nil {
		return nil, err
	}
	logger.Info("database provider initialized")

	// Initialize router
	rateLimiter := limiter.NewBurstRateLimiter(cfg.RPSLimit, cfg.RPSBurst, logger)
	ipFinder := finder.NewIpFinder(dbProvider)
	appRouter := router.NewRouter(rateLimiter, tel, logger)
	server := appRouter.CreateServer(":"+cfg.Port, ipFinder)

	return &App{
		config:    cfg,
		logger:    logger,
		telemetry: tel,
		server:    server,
	}, nil
}

// Start starts the application server
func (app *App) start() error {
	app.logger.Info("starting server", zap.String("port", app.config.Port))

	go func() {
		if err := app.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	return nil
}

// Stop gracefully shuts down the application
func (app *App) stop() error {
	app.logger.Info("shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.server.Shutdown(shutdownCtx); err != nil {
		app.logger.Error("server forced to shutdown", zap.Error(err))
		return err
	}

	app.logger.Info("server exited gracefully")
	return nil
}

// Run starts the application and waits for shutdown signals
func (app *App) Run() error {
	// Start the server
	if err := app.start(); err != nil {
		return err
	}

	// Wait for interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Wait for shutdown signal
	<-ctx.Done()

	// Stop the application
	return app.stop()
}
