package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSPreflightAllowsEveryMethodTheAPIUses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(cors([]string{"http://localhost:4173"}))
	router.PUT("/v1/profile", func(c *gin.Context) { c.Status(http.StatusOK) })

	request := httptest.NewRequest(http.MethodOptions, "/v1/profile", nil)
	request.Header.Set("Origin", "http://localhost:4173")
	request.Header.Set("Access-Control-Request-Method", http.MethodPut)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	allowed := recorder.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(allowed, http.MethodPut) {
		t.Fatalf("Access-Control-Allow-Methods = %q, want it to include PUT (profile and notification-settings saves use PUT)", allowed)
	}
}
