package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"
	country_finder "torq/CountryFinder"
	"torq/limiter"
	"torq/lookup"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	_ = godotenv.Load()

	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			panic("failed to sync logger: " + err.Error())
		}
	}(logger)
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

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

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

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

// RateLimitMiddleware wraps handlers with the rate limiter
func RateLimitMiddleware(l *limiter.RpsRateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	backend := os.Getenv("IP_DB_PROVIDER")
	logger.Info("initializing service", zap.String("backend", backend))

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

	// Create and apply rate limiter middleware (e.g., 10 requests per second)
	rpsLimitStr := os.Getenv("RPS_LIMIT")
	rpsLimit := 10 // default RPS limit
	if rpsLimitStr != "" {
		if val, err := strconv.Atoi(rpsLimitStr); err == nil && val > 0 {
			rpsLimit = val
		} else {
			logger.Warn("invalid RPS_LIMIT, using default", zap.String("RPS_LIMIT", rpsLimitStr))
		}
	}
	rl := limiter.NewRateLimiter(rpsLimit, logger)
	rateLimitedRouter := RateLimitMiddleware(rl, router)

	port := ":8080"
	logger.Info("starting server", zap.String("port", port))

	srv := &http.Server{
		Addr:         port,
		Handler:      rateLimitedRouter,
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
