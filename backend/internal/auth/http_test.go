package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoginSetsSafeSessionCookie(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository)
	if _, err := service.Register(context.Background(), "user@example.com", "correct horse battery staple"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	response := httptest.NewRecorder()
	NewHandler(service).ServeHTTP(response, jsonRequest(t, http.MethodPost, "/v1/auth/login", map[string]string{
		"email":    "user@example.com",
		"password": "correct horse battery staple",
	}))

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unsafe session cookie: %#v", cookie)
	}
	if cookie.Value == "" {
		t.Fatal("session cookie value is empty")
	}
	if strings.Contains(response.Body.String(), cookie.Value) {
		t.Fatal("login JSON body exposes the raw session token")
	}
}

func TestProtectedEndpointRejectsMissingAndInvalidCookie(t *testing.T) {
	handler := NewHandler(NewService(newMemoryRepository()))

	tests := []struct {
		name   string
		cookie *http.Cookie
	}{
		{name: "missing"},
		{name: "invalid", cookie: &http.Cookie{Name: SessionCookieName, Value: "tampered-token"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/v1/auth/session", nil)
			if test.cookie != nil {
				request.AddCookie(test.cookie)
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status code = %d, want %d", response.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestSessionMiddlewareAttachesAccountToContext(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository)
	registered, err := service.Register(context.Background(), "user@example.com", "correct horse battery staple")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	rawToken, _, err := service.Login(context.Background(), "user@example.com", "correct horse battery staple")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	nextCalled := false
	protected := RequireSession(service, http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		nextCalled = true
		current, ok := AccountFromContext(request.Context())
		if !ok {
			t.Error("AccountFromContext() ok = false, want true")
			return
		}
		if current.Email != registered.Email {
			t.Errorf("context account email = %q, want %q", current.Email, registered.Email)
		}
		if current.ID != registered.ID {
			t.Errorf("context account ID = %s, want %s", current.ID, registered.ID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: rawToken})
	response := httptest.NewRecorder()

	protected.ServeHTTP(response, request)

	if !nextCalled {
		t.Fatal("protected handler was not called")
	}
	if response.Code != http.StatusNoContent {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestLoginUsesSamePublicErrorForUnknownEmailAndWrongPassword(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository)
	if _, err := service.Register(context.Background(), "user@example.com", "correct horse battery staple"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	handler := NewHandler(service)

	wrongPassword := httptest.NewRecorder()
	handler.ServeHTTP(wrongPassword, jsonRequest(t, http.MethodPost, "/v1/auth/login", map[string]string{
		"email": "user@example.com", "password": "wrong password",
	}))
	unknownEmail := httptest.NewRecorder()
	handler.ServeHTTP(unknownEmail, jsonRequest(t, http.MethodPost, "/v1/auth/login", map[string]string{
		"email": "missing@example.com", "password": "wrong password",
	}))

	if wrongPassword.Code != http.StatusUnauthorized || unknownEmail.Code != http.StatusUnauthorized {
		t.Fatalf("status codes = %d and %d, want both %d", wrongPassword.Code, unknownEmail.Code, http.StatusUnauthorized)
	}
	if wrongPassword.Body.String() != unknownEmail.Body.String() {
		t.Fatalf("public login errors differ: %q vs %q", wrongPassword.Body, unknownEmail.Body)
	}
}

func TestLogoutClearsCookieAndInvalidatesSession(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository)
	if _, err := service.Register(context.Background(), "user@example.com", "correct horse battery staple"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	rawToken, _, err := service.Login(context.Background(), "user@example.com", "correct horse battery staple")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: rawToken})
	response := httptest.NewRecorder()
	NewHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusNoContent)
	}
	assertExpiredSessionCookie(t, response)
	if _, err := service.AccountForSession(context.Background(), rawToken); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("AccountForSession() after logout error = %v, want ErrUnauthenticated", err)
	}
}

func TestLogoutExpiresCookieWhenSessionDeletionFails(t *testing.T) {
	repository := newMemoryRepository()
	repository.deleteSessionErr = errors.New("database unavailable")
	request := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "opaque-session-token"})
	response := httptest.NewRecorder()

	NewHandler(NewService(repository)).ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	assertExpiredSessionCookie(t, response)
}

func TestSessionEndpointRejectsExpiredSession(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository)
	ctx := context.Background()

	if _, err := service.Register(ctx, "user@example.com", "correct horse battery staple"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	rawToken, _, err := service.Login(ctx, "user@example.com", "correct horse battery staple")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	repository.sessions[0].ExpiresAt = time.Now().Add(-time.Second)
	request := httptest.NewRequest(http.MethodGet, "/v1/auth/session", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: rawToken})
	response := httptest.NewRecorder()

	NewHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestRegisterEndpointReturnsNormalizedAccount(t *testing.T) {
	response := httptest.NewRecorder()
	NewHandler(NewService(newMemoryRepository())).ServeHTTP(response, jsonRequest(t, http.MethodPost, "/v1/auth/register", map[string]string{
		"email": "User@Example.COM", "password": "correct horse battery staple",
	}))

	if response.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d; body = %s", response.Code, http.StatusCreated, response.Body)
	}
	var body Account
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Email != "user@example.com" {
		t.Fatalf("response email = %q, want %q", body.Email, "user@example.com")
	}
}

func jsonRequest(t *testing.T, method, target string, body any) *http.Request {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	request := httptest.NewRequest(method, target, bytes.NewReader(encoded))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func assertExpiredSessionCookie(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != SessionCookieName || cookie.Value != "" || cookie.MaxAge >= 0 {
		t.Fatalf("logout cookie = %#v, want expired %q cookie", cookie, SessionCookieName)
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unsafe expired session cookie: %#v", cookie)
	}
}
