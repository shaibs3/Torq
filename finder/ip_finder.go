package finder

import (
	"encoding/json"
	"github.com/shaibs3/Torq/lookup"
	"net/http"
)

type IpFinder struct {
	provider lookup.DbProvider
}

func NewIpFinder(provider lookup.DbProvider) *IpFinder {
	return &IpFinder{provider: provider}
}

func (cf *IpFinder) FindIpHandler(w http.ResponseWriter, r *http.Request) {
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
