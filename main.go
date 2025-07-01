package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func findCountryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Find country endpoint"}`))
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/v1/find-country", findCountryHandler).Methods("GET")

	port := ":8080"
	log.Printf("Server is running on port %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}
