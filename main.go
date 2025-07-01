package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
	country_finder "torq/CountryFinder"
	"torq/lookup"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	_ = godotenv.Load()

	// Initialize zap logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

// livenessHandler checks if the service is alive
func livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := HealthResponse{
		Status:    "alive",
		Timestamp: time.Now(),
		Service:   "torq",
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		logger.Error("failed to encode liveness response", zap.Error(err))
		return
	}

	logger.Info("liveness check completed",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr))
}

// readinessHandler checks if the service is ready to serve requests
func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Check if the provider is properly initialized
	backend := os.Getenv("IP_DB_PROVIDER")
	status := "ready"
	if backend == "" {
		status = "not ready"
		logger.Warn("service not ready - missing IP_DB_PROVIDER configuration")
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Service:   "torq",
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		logger.Error("failed to encode readiness response", zap.Error(err))
		return
	}

	logger.Info("readiness check completed",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("status", status),
		zap.String("remote_addr", r.RemoteAddr))
}

func main() {
	backend := os.Getenv("IP_DB_PROVIDER")
	logger.Info("initializing service", zap.String("backend", backend))

	// Create a named logger for the lookup provider
	lookupLogger := logger.Named("lookup")
	provider, err := lookup.NewProvider(backend, lookupLogger)
	if err != nil {
		logger.Fatal("failed to init provider", zap.Error(err), zap.String("backend", backend))
	}

	logger.Info("provider initialized successfully", zap.String("backend", backend))

	CountryFinder := country_finder.NewCountryFinder(provider)
	router := mux.NewRouter()

	// Health check endpoints
	router.HandleFunc("/health/live", livenessHandler).Methods("GET")
	router.HandleFunc("/health/ready", readinessHandler).Methods("GET")

	// API endpoints
	router.HandleFunc("/v1/find-country", CountryFinder.FindCountryHandler).Methods("GET")

	port := ":8080"
	logger.Info("starting server", zap.String("port", port))

	srv := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Info("server configuration",
		zap.String("addr", srv.Addr),
		zap.Duration("read_timeout", srv.ReadTimeout),
		zap.Duration("write_timeout", srv.WriteTimeout),
		zap.Duration("idle_timeout", srv.IdleTimeout))

	logger.Info("server is running", zap.String("port", port))
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal("server failed to start", zap.Error(err))
	}
}
