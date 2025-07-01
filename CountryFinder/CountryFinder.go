package country_finder

import (
	"encoding/json"
	"net/http"
	"torq/lookup"
)

type CountryFinder struct {
	provider lookup.LookupProvider
}

func NewCountryFinder(provider lookup.LookupProvider) *CountryFinder {
	return &CountryFinder{provider: provider}
}

func (cf *CountryFinder) FindCountryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		http.Error(w, `{"error":"ip parameter is required"}`, http.StatusBadRequest)
		return
	}

	country, city, err := cf.provider.Lookup(ip)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	resp := map[string]string{
		"country": country,
		"city":    city,
	}
	jsonResp, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonResp)
}
