package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shaibs3/Torq/internal/finder"
	"github.com/shaibs3/Torq/internal/lookup"
	"github.com/shaibs3/Torq/internal/router"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/zap"
)

func main() {
	// Load .env if present
	err := godotenv.Load()
	if err != nil {
		panic("failed to load .env file" + err.Error())
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Initialize OpenTelemetry with Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		logger.Fatal("failed to initialize prometheus exporter", zap.Error(err))
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
	)
	otel.SetMeterProvider(provider)

	logger.Info("OpenTelemetry metrics initialized with Prometheus exporter")

	// Initialize metrics
	meter := otel.GetMeterProvider().Meter("torq")
	httpMetrics := router.NewHTTPMetrics(meter, logger.Named("metrics"))

	// Init IP DB provider
	providerType := os.Getenv("IP_DB_PROVIDER")
	logger.Info("initializing service", zap.String("provider_type", providerType))

	dbProvider, err := lookup.GetDbProvider(providerType, logger.Named("db_provider"))
	if err != nil {
		logger.Fatal("failed to initialize provider", zap.Error(err), zap.String("provider_type", providerType))
	}
	logger.Info("provider initialized", zap.String("provider_type", providerType))

	// Init country finder
	ipFinder := finder.NewIpFinder(dbProvider)

	// Init router
	appRouter := router.NewRouter(logger)
	appRouter.SetupRoutes(ipFinder)

	// Rate limit config
	rpsLimit := router.ParseRPSLimit(os.Getenv("RPS_LIMIT"), logger)

	// Setup middleware with metrics
	handler := appRouter.SetupMiddleware(rpsLimit, httpMetrics)

	// Create server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port

	server := appRouter.CreateServer(port, handler)
	logger.Info("server is running", zap.String("port", port))

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	logger.Info("shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited")
}
