package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// cors allows a browser-based frontend running on a different origin (e.g. a
// Vite or Next.js dev server) to call the API with the session cookie
// attached. "*" cannot be combined with credentials per the Fetch spec, so a
// matching request origin is reflected back explicitly instead.
func cors(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
