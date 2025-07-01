package country_finder

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	country string
	city    string
	err     error
}

func (m *mockProvider) Lookup(ip string) (string, string, error) {
	if m.err != nil {
		return "", "", m.err
	}
	return m.country, m.city, nil
}

func TestFindCountryHandler_Success(t *testing.T) {
	provider := &mockProvider{country: "USA", city: "New York"}
	cf := NewCountryFinder(provider)

	req := httptest.NewRequest("GET", "/v1/find-country?ip=1.2.3.4", nil)
	w := httptest.NewRecorder()

	cf.FindCountryHandler(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := w.Body.String()
	require.Contains(t, body, `"country":"USA"`)
	require.Contains(t, body, `"city":"New York"`)
}

func TestFindCountryHandler_MissingIP(t *testing.T) {
	provider := &mockProvider{}
	cf := NewCountryFinder(provider)

	req := httptest.NewRequest("GET", "/v1/find-country", nil)
	w := httptest.NewRecorder()

	cf.FindCountryHandler(w, req)

	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	require.Contains(t, w.Body.String(), "ip parameter is required")
}

func TestFindCountryHandler_NotFound(t *testing.T) {
	provider := &mockProvider{err: errors.New("not found")}
	cf := NewCountryFinder(provider)

	req := httptest.NewRequest("GET", "/v1/find-country?ip=8.8.8.8", nil)
	w := httptest.NewRecorder()

	cf.FindCountryHandler(w, req)

	require.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	require.Contains(t, w.Body.String(), "not found")
}
