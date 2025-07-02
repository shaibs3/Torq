package router

import (
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

// HTTPMetrics holds all HTTP-related metrics
type HTTPMetrics struct {
	RequestDuration metric.Float64Histogram
	RequestCount    metric.Int64Counter
	ErrorRequests   metric.Int64Counter
	ResponseStatus  metric.Int64Counter
	ActiveRequests  metric.Int64UpDownCounter
}

// NewHTTPMetrics creates and registers all HTTP metrics
func NewHTTPMetrics(meter metric.Meter, logger *zap.Logger) *HTTPMetrics {
	requestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		logger.Error("failed to create request duration metric", zap.Error(err))
	}

	requestCount, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		logger.Error("failed to create request count metric", zap.Error(err))
	}

	errorRequests, err := meter.Int64Counter(
		"http_error_requests_total",
		metric.WithDescription("Total number of HTTP error requests (4xx, 5xx)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		logger.Error("failed to create error requests metric", zap.Error(err))
	}

	responseStatus, err := meter.Int64Counter(
		"http_response_status_total",
		metric.WithDescription("Total number of HTTP responses by status code"),
		metric.WithUnit("1"),
	)
	if err != nil {
		logger.Error("failed to create response status metric", zap.Error(err))
	}

	activeRequests, err := meter.Int64UpDownCounter(
		"http_requests_in_flight",
		metric.WithDescription("Number of HTTP requests currently in flight"),
		metric.WithUnit("1"),
	)
	if err != nil {
		logger.Error("failed to create active requests metric", zap.Error(err))
	}

	return &HTTPMetrics{
		RequestDuration: requestDuration,
		RequestCount:    requestCount,
		ErrorRequests:   errorRequests,
		ResponseStatus:  responseStatus,
		ActiveRequests:  activeRequests,
	}
}

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
	r.router.HandleFunc("/health/live", service_health.LivenessHandler(r.logger)).Methods("GET", "HEAD")
	r.router.HandleFunc("/health/ready", service_health.ReadinessHandler(r.logger)).Methods("GET", "HEAD")

	// Metrics endpoint
	r.router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// API endpoints
	r.router.HandleFunc("/v1/find-country", countryFinder.FindIpHandler).Methods("GET")

	r.logger.Info("routes configured successfully")
}

// SetupMiddleware configures rate limiting and metrics middleware
func (r *Router) SetupMiddleware(rpsLimit int, metrics *HTTPMetrics) http.Handler {
	r.logger.Info("setting up middleware", zap.Int("rps_limit", rpsLimit))

	// Create rate limiter
	rl := limiter.NewRateLimiter(rpsLimit, r.logger.Named("rate_limiter"))

	// Apply middlewares in order: metrics -> rate limiting -> router
	metricsHandler := MetricsMiddleware(metrics, r.logger.Named("metrics"))(r.router)
	rateLimitedRouter := RateLimitMiddleware(rl, metricsHandler)

	r.logger.Info("middleware configured successfully")
	return rateLimitedRouter
}

// MetricsMiddleware creates middleware for comprehensive HTTP metrics
func MetricsMiddleware(metrics *HTTPMetrics, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Increment active requests
			if metrics.ActiveRequests != nil {
				metrics.ActiveRequests.Add(r.Context(), 1)
				defer metrics.ActiveRequests.Add(r.Context(), -1)
			}

			// Create response writer wrapper to capture status code
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

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

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	return size, err
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
