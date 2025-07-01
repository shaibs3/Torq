package main

import (
	"log"
	"net/http"
	"os"
	country_finder "torq/CountryFinder"
	"torq/lookup"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}
func main() {
	backend := os.Getenv("IP_DB_PROVIDER")
	provider, err := lookup.NewProvider(backend)
	if err != nil {
		log.Fatalf("failed to init provider: %v", err)
	}

	CountryFinder := country_finder.NewCountryFinder(provider)
	router := mux.NewRouter()
	router.HandleFunc("/v1/find-country", CountryFinder.FindCountryHandler).Methods("GET")

	port := ":8080"
	log.Printf("Server is running on port %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}
