package main

import (
	"database/sql"
	"log"
	"net/http"

	"llm-homework/backend/internal/config"
	httpserver "llm-homework/backend/internal/http"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	configuration, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", configuration.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:    configuration.APIAddress,
		Handler: httpserver.NewServer(db),
	}
	log.Fatal(server.ListenAndServe())
}
