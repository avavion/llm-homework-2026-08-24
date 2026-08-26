package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzReturnsOK(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	NewServer(nil).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}

	if body := response.Body.String(); body != "{\"status\":\"ok\"}\n" {
		t.Fatalf("response body = %q, want %q", body, "{\"status\":\"ok\"}\n")
	}
}
