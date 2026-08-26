package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"llm-homework/backend/internal/config"
	httpserver "llm-homework/backend/internal/http"
	"llm-homework/backend/internal/notification"
	"llm-homework/backend/internal/regulation"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	shutdownTimeout           = 10 * time.Second
	notificationCheckInterval = time.Minute
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
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Printf("ping database: %v", err)
		return
	}

	server := &http.Server{
		Addr:    configuration.APIAddress,
		Handler: httpserver.NewServer(db),
	}

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	notificationService := notification.NewService(
		notification.NewRepository(db), regulation.NewRepository(),
		notification.NewRepository(db), notification.NewDevSender(),
	)
	go notification.NewWorker(notificationService, notificationCheckInterval).Run(workerCtx)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	if err := runServer(server, server.ListenAndServe, signals); err != nil {
		log.Printf("run API server: %v", err)
	}
}

func runServer(server *http.Server, serve func() error, signals <-chan os.Signal) error {
	serveResult := make(chan error, 1)
	go func() {
		serveResult <- serve()
	}()

	select {
	case err := <-serveResult:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-signals:
		shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownContext); err != nil {
			return err
		}

		err := <-serveResult
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
