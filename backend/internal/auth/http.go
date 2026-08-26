package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

const SessionCookieName = "session"

const maxRequestBodyBytes = 1 << 20

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ServiceAPI interface {
	Register(ctx context.Context, email, password string) (Account, error)
	Login(ctx context.Context, email, password string) (rawToken string, current Account, err error)
	Logout(ctx context.Context, rawToken string) error
	AccountForSession(ctx context.Context, rawToken string) (Account, error)
}

func NewHandler(service ServiceAPI) http.Handler {
	mux := http.NewServeMux()
	RegisterRoutes(mux, service)
	return mux
}

func RegisterRoutes(mux *http.ServeMux, service ServiceAPI) {
	mux.HandleFunc("POST /v1/auth/register", registerHandler(service))
	mux.HandleFunc("POST /v1/auth/login", loginHandler(service))
	mux.HandleFunc("POST /v1/auth/logout", logoutHandler(service))
	mux.Handle("GET /v1/auth/session", RequireSession(service, http.HandlerFunc(sessionHandler)))
}

func registerHandler(service ServiceAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		credentials, ok := decodeCredentials(w, request)
		if !ok {
			return
		}

		registered, err := service.Register(request.Context(), credentials.Email, credentials.Password)
		switch {
		case errors.Is(err, ErrInvalidRegistration):
			writeError(w, http.StatusBadRequest, "invalid registration")
		case errors.Is(err, ErrEmailTaken):
			writeError(w, http.StatusConflict, ErrEmailTaken.Error())
		case err != nil:
			writeError(w, http.StatusInternalServerError, "internal server error")
		default:
			writeJSON(w, http.StatusCreated, registered)
		}
	}
}

func loginHandler(service ServiceAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		credentials, ok := decodeCredentials(w, request)
		if !ok {
			return
		}

		rawToken, current, err := service.Login(request.Context(), credentials.Email, credentials.Password)
		if errors.Is(err, ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, ErrInvalidCredentials.Error())
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		http.SetCookie(w, sessionCookie(rawToken, time.Now().Add(sessionLifetime)))
		writeJSON(w, http.StatusOK, current)
	}
}

func logoutHandler(service ServiceAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		http.SetCookie(w, expiredSessionCookie())

		cookie, err := request.Cookie(SessionCookieName)
		if err == nil {
			if err := service.Logout(request.Context(), cookie.Value); err != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func sessionHandler(w http.ResponseWriter, request *http.Request) {
	current, ok := AccountFromContext(request.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrUnauthenticated.Error())
		return
	}
	writeJSON(w, http.StatusOK, current)
}

func decodeCredentials(w http.ResponseWriter, request *http.Request) (credentialsRequest, bool) {
	request.Body = http.MaxBytesReader(w, request.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	var credentials credentialsRequest
	if err := decoder.Decode(&credentials); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return credentialsRequest{}, false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid request")
		return credentialsRequest{}, false
	}
	return credentials, true
}

func sessionCookie(value string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(sessionLifetime.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func expiredSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
