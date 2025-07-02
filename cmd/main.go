package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

func InitOpenTelemetryMetrics() func(context.Context) error {
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatalf("failed to initialize prometheus exporter: %v", err)
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
	)
	otel.SetMeterProvider(provider)

	// Expose metrics endpoint
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Prometheus metrics exposed at /metrics")
		if err := http.ListenAndServe(":9464", nil); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	return provider.Shutdown
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdown := InitOpenTelemetryMetrics()
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Printf("Error shutting down OpenTelemetry: %v", err)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Torq"))
	})

	port := "8080"
	log.Printf("Server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
