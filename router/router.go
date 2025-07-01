package router

import (
	"github.com/shaibs3/Torq/finder"
	"github.com/shaibs3/Torq/limiter"
	"github.com/shaibs3/Torq/service_health"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Router handles all routing logic and middleware setup
type Router struct {
	router *mux.Router
	logger *zap.Logger
}

// NewRouter creates a new router instance with all routes and middleware configured
func NewRouter(logger *zap.Logger) *Router {
	r := &Router{
		router: mux.NewRouter(),
		logger: logger.Named("router"),
	}
	return r
}

// SetupRoutes configures all application routes
func (r *Router) SetupRoutes(countryFinder *finder.IpFinder) {
	r.logger.Info("setting up application routes")

	// Health check endpoints
	r.router.HandleFunc("/health/live", service_health.LivenessHandler(r.logger)).Methods("GET")
	r.router.HandleFunc("/health/ready", service_health.ReadinessHandler(r.logger)).Methods("GET")

	// API endpoints
	r.router.HandleFunc("/v1/find-country", countryFinder.FindCountryHandler).Methods("GET")

	r.logger.Info("routes configured successfully")
}

// SetupMiddleware configures rate limiting middleware
func (r *Router) SetupMiddleware(rpsLimit int) http.Handler {
	r.logger.Info("setting up middleware", zap.Int("rps_limit", rpsLimit))

	// Create rate limiter
	rl := limiter.NewRateLimiter(rpsLimit, r.logger.Named("rate_limiter"))

	// Apply rate limiting middleware
	rateLimitedRouter := RateLimitMiddleware(rl, r.router)

	r.logger.Info("middleware configured successfully")
	return rateLimitedRouter
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

// GetRouter returns the configured router
func (r *Router) GetRouter() *mux.Router {
	return r.router
}

// CreateServer creates and configures an HTTP server with the router
func (r *Router) CreateServer(port string, handler http.Handler) *http.Server {
	r.logger.Info("creating HTTP server", zap.String("port", port))

	srv := &http.Server{
		Addr:         port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	r.logger.Info("server configuration",
		zap.String("addr", srv.Addr),
		zap.Duration("read_timeout", srv.ReadTimeout),
		zap.Duration("write_timeout", srv.WriteTimeout),
		zap.Duration("idle_timeout", srv.IdleTimeout))

	return srv
}

// ParseRPSLimit parses the RPS limit from environment variable with fallback
func ParseRPSLimit(rpsLimitStr string, logger *zap.Logger) int {
	rpsLimit := 10 // default RPS limit
	if rpsLimitStr != "" {
		if val, err := strconv.Atoi(rpsLimitStr); err == nil && val > 0 {
			rpsLimit = val
		} else {
			logger.Warn("invalid RPS_LIMIT, using default", zap.String("RPS_LIMIT", rpsLimitStr))
		}
	}
	return rpsLimit
}
