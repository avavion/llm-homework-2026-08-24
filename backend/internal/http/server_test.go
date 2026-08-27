package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzReturnsOK(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	NewServer(nil, nil, "").ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}

	if body := response.Body.String(); body != "{\"status\":\"ok\"}\n" {
		t.Fatalf("response body = %q, want %q", body, "{\"status\":\"ok\"}\n")
	}
}

func TestAllowedFrontendOriginGetsCORSHeaders(t *testing.T) {
	request := httptest.NewRequest(http.MethodOptions, "/v1/products", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", "POST")
	response := httptest.NewRecorder()

	NewServer(nil, []string{"http://localhost:5173"}, "").ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("Access-Control-Allow-Credentials = %q", got)
	}
}

func TestUnlistedOriginGetsNoCORSHeaders(t *testing.T) {
	request := httptest.NewRequest(http.MethodOptions, "/v1/products", nil)
	request.Header.Set("Origin", "https://not-allowed.example.com")
	response := httptest.NewRecorder()

	NewServer(nil, []string{"http://localhost:5173"}, "").ServeHTTP(response, request)

	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty for an unlisted origin", got)
	}
}
