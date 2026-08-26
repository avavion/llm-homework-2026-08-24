package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const ginAccountKey = "auth.account"

// GinRequireSession is the Gin-native counterpart of RequireSession, used by
// modules whose HTTP layer is built on gin.Engine instead of net/http.
func GinRequireSession(service ServiceAPI) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(SessionCookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthenticated.Error()})
			return
		}

		current, err := service.AccountForSession(c.Request.Context(), cookie)
		if errors.Is(err, ErrUnauthenticated) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthenticated.Error()})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.Set(ginAccountKey, current)
		c.Next()
	}
}

// GinAccount reads the account stored by GinRequireSession.
func GinAccount(c *gin.Context) (Account, bool) {
	value, ok := c.Get(ginAccountKey)
	if !ok {
		return Account{}, false
	}
	current, ok := value.(Account)
	return current, ok
}
