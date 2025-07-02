package telemetry

import (
	"github.com/shaibs3/Torq/internal/router"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/zap"
)

// Telemetry handles OpenTelemetry initialization and metrics
type Telemetry struct {
	HTTPMetrics *router.HTTPMetrics
	logger      *zap.Logger
}

// New initializes OpenTelemetry with Prometheus exporter
func New(logger *zap.Logger) (*Telemetry, error) {
	logger = logger.Named("telemetry")

	// Initialize Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	// Create meter provider
	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
	)
	otel.SetMeterProvider(provider)

	logger.Info("OpenTelemetry metrics initialized with Prometheus exporter")

	// Initialize HTTP metrics
	meter := otel.GetMeterProvider().Meter("torq")
	httpMetrics := router.NewHTTPMetrics(meter, logger.Named("metrics"))

	return &Telemetry{
		HTTPMetrics: httpMetrics,
		logger:      logger,
	}, nil
}

// GetHTTPMetrics returns the HTTP metrics instance
func (t *Telemetry) GetHTTPMetrics() *router.HTTPMetrics {
	return t.HTTPMetrics
}
