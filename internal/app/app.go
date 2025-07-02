package app

import (
	"context"
	"github.com/shaibs3/Torq/internal/limiter"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shaibs3/Torq/internal/config"
	"github.com/shaibs3/Torq/internal/finder"
	"github.com/shaibs3/Torq/internal/lookup"
	"github.com/shaibs3/Torq/internal/router"
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

// New creates a new application instance
func New(cfg *config.Config, logger *zap.Logger) (*App, error) {
	// Initialize telemetry
	tel, err := telemetry.New(logger)
	if err != nil {
		return nil, err
	}

	// Initialize IP DB provider
	dbProvider, err := lookup.GetDbProvider(cfg.IPDBConfig, logger.Named("db_provider"))
	if err != nil {
		return nil, err
	}
	logger.Info("database provider initialized")

	// Initialize country finder
	ipFinder := finder.NewIpFinder(dbProvider)

	// Initialize router
	appRouter := router.NewRouter(limiter.NewRateLimiter(cfg.RPSLimit, logger.Named("rate_limiter")), logger)
	appRouter.SetupRoutes(ipFinder)

	// Setup middleware with metrics
	handler := appRouter.SetupMiddleware(tel.GetHTTPMetrics())

	// Create server
	port := ":" + cfg.Port
	server := appRouter.CreateServer(port, handler)

	return &App{
		config:    cfg,
		logger:    logger,
		telemetry: tel,
		server:    server,
	}, nil
}

// Start starts the application server
func (a *App) Start() error {
	a.logger.Info("starting server", zap.String("port", a.config.Port))

	// Start server in a goroutine
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	return nil
}

// Stop gracefully shuts down the application
func (a *App) Stop() error {
	a.logger.Info("shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("server forced to shutdown", zap.Error(err))
		return err
	}

	a.logger.Info("server exited gracefully")
	return nil
}

// Run starts the application and waits for shutdown signals
func (a *App) Run() error {
	// Start the server
	if err := a.Start(); err != nil {
		return err
	}

	// Wait for interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Wait for shutdown signal
	<-ctx.Done()

	// Stop the application
	return a.Stop()
}
