package service_health

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

// LivenessHandler checks if the service is alive
func LivenessHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}

// ReadinessHandler checks if the service is ready to serve requests
func ReadinessHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}
