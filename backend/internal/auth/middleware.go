package auth

import (
	"context"
	"errors"
	"net/http"
)

type accountContextKey struct{}

func RequireSession(service ServiceAPI, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		cookie, err := request.Cookie(SessionCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, ErrUnauthenticated.Error())
			return
		}

		current, err := service.AccountForSession(request.Context(), cookie.Value)
		if errors.Is(err, ErrUnauthenticated) {
			writeError(w, http.StatusUnauthorized, ErrUnauthenticated.Error())
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		ctx := context.WithValue(request.Context(), accountContextKey{}, current)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

func AccountFromContext(ctx context.Context) (Account, bool) {
	current, ok := ctx.Value(accountContextKey{}).(Account)
	return current, ok
}
