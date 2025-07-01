package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
	country_finder "torq/CountryFinder"
	"torq/lookup"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
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
		return
	}
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
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Service:   "torq",
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func main() {
	backend := os.Getenv("IP_DB_PROVIDER")
	provider, err := lookup.NewProvider(backend)
	if err != nil {
		log.Fatalf("failed to init provider: %v", err)
	}

	CountryFinder := country_finder.NewCountryFinder(provider)
	router := mux.NewRouter()

	// Health check endpoints
	router.HandleFunc("/health/live", livenessHandler).Methods("GET")
	router.HandleFunc("/health/ready", readinessHandler).Methods("GET")

	// API endpoints
	router.HandleFunc("/v1/find-country", CountryFinder.FindCountryHandler).Methods("GET")

	port := ":8080"

	srv := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Printf("Server is running on port %s", port)
	log.Fatal(srv.ListenAndServe())
}
