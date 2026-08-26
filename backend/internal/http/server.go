package http

import (
	"database/sql"
	"net/http"

	"llm-homework/backend/internal/account"
	"llm-homework/backend/internal/auth"
)

func NewServer(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
	})
	auth.RegisterRoutes(mux, auth.NewService(account.NewRepository(db)))
	return mux
}
