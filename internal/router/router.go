package router

import (
	"github.com/shaibs3/Torq/internal/telemetry"
	"net/http"
	"strconv"
	"time"

	"github.com/shaibs3/Torq/internal/finder"
	"github.com/shaibs3/Torq/internal/limiter"
	"github.com/shaibs3/Torq/internal/service_health"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// Router handles all routing logic and middleware setup
type Router struct {
	router      *mux.Router
	rateLimiter limiter.RateLimiter
	logger      *zap.Logger
}

// NewRouter creates a new router instance with all routes and middleware configured
func NewRouter(rateLimiter limiter.RateLimiter, logger *zap.Logger) *Router {
	r := &Router{
		router:      mux.NewRouter(),
		rateLimiter: rateLimiter,
		logger:      logger.Named("router"),
	}
	return r
}

// SetupRoutes configures all application routes
func (r *Router) SetupRoutes(countryFinder *finder.IpFinder) {
	r.logger.Info("setting up application routes")

	// Health check endpoints
	r.router.HandleFunc("/health/live", service_health.LivenessHandler(r.logger)).Methods("GET", "HEAD")
	r.router.HandleFunc("/health/ready", service_health.ReadinessHandler(r.logger)).Methods("GET", "HEAD")

	// Metrics endpoint
	r.router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// API endpoints
	r.router.HandleFunc("/v1/find-country", countryFinder.FindIpHandler).Methods("GET")

	r.logger.Info("routes configured successfully")
}

// SetupMiddleware configures rate limiting and metrics middleware
func (r *Router) SetupMiddleware(metrics *telemetry.HTTPMetrics) http.Handler {
	r.logger.Info("setting up middleware")

	// Apply middlewares in order: metrics -> rate limiting -> router
	metricsHandler := metricsMiddleware(metrics, r.logger.Named("metrics"))(r.router)
	rateLimitedRouter := r.rateLimitMiddleware(metricsHandler)

	r.logger.Info("middleware configured successfully")
	return rateLimitedRouter
}

// MetricsMiddleware creates middleware for comprehensive HTTP metrics
func metricsMiddleware(metrics *telemetry.HTTPMetrics, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Increment active requests
			if metrics.ActiveRequests != nil {
				metrics.ActiveRequests.Add(r.Context(), 1)
				defer metrics.ActiveRequests.Add(r.Context(), -1)
			}

			// Create response writer wrapper to capture status code
			wrappedWriter := &ResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call next handler
			next.ServeHTTP(wrappedWriter, r)

			// Record metrics
			duration := time.Since(start)

			attrs := []attribute.KeyValue{
				attribute.String("method", r.Method),
				attribute.String("path", r.URL.Path),
				attribute.Int("status_code", wrappedWriter.statusCode),
			}

			// Record request duration
			if metrics.RequestDuration != nil {
				metrics.RequestDuration.Record(r.Context(), duration.Seconds(), metric.WithAttributes(attrs...))
			}

			// Record request count
			if metrics.RequestCount != nil {
				metrics.RequestCount.Add(r.Context(), 1, metric.WithAttributes(attrs...))
			}

			// Record error requests (4xx, 5xx status codes)
			if metrics.ErrorRequests != nil && (wrappedWriter.statusCode >= 400) {
				errorAttrs := []attribute.KeyValue{
					attribute.String("method", r.Method),
					attribute.String("path", r.URL.Path),
					attribute.String("status_code", strconv.Itoa(wrappedWriter.statusCode)),
				}
				metrics.ErrorRequests.Add(r.Context(), 1, metric.WithAttributes(errorAttrs...))
			}

			// Record response status
			if metrics.ResponseStatus != nil {
				statusAttrs := []attribute.KeyValue{
					attribute.String("status_code", strconv.Itoa(wrappedWriter.statusCode)),
				}
				metrics.ResponseStatus.Add(r.Context(), 1, metric.WithAttributes(statusAttrs...))
			}

			logger.Info("request completed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status_code", wrappedWriter.statusCode),
				zap.Duration("duration", duration),
				zap.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

func (router *Router) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !router.rateLimiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
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
