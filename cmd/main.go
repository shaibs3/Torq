package main

import (
	"github.com/shaibs3/Torq/internal/finder"
	"github.com/shaibs3/Torq/internal/lookup"
	"github.com/shaibs3/Torq/internal/router"
	"os"

	"github.com/joho/godotenv"
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

	// Init IP DB provider
	providerType := os.Getenv("IP_DB_PROVIDER")
	logger.Info("initializing service", zap.String("provider_type", providerType))

	dbProvider, err := lookup.GetDbProvider(providerType, logger.Named("db_provider"))
	if err != nil {
		logger.Fatal("failed to initialize provider", zap.Error(err), zap.String("provider_type", providerType))
	}
	logger.Info("provider initialized", zap.String("provider_type", providerType))

	// Init country finder
	countryFinder := finder.NewIpFinder(dbProvider)

	// Init router
	appRouter := router.NewRouter(logger)
	appRouter.SetupRoutes(countryFinder)

	// Rate limit config
	rpsLimit := router.ParseRPSLimit(os.Getenv("RPS_LIMIT"), logger)

	// Middleware and server setup
	handler := appRouter.SetupMiddleware(rpsLimit)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port

	server := appRouter.CreateServer(port, handler)
	logger.Info("server is running", zap.String("port", port))

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("server failed to start", zap.Error(err))
	}
}
