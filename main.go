package main

import (
	"log"
	"net/http"
	"torq/lookup"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func findCountryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Find country endpoint"}`))
}

func init() {
	_ = godotenv.Load()
}
func main() {

	_, err := lookup.NewProvider()
	if err != nil {
		log.Fatalf("failed to init provider: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/v1/find-country", findCountryHandler).Methods("GET")

	port := ":8080"
	log.Printf("Server is running on port %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}
