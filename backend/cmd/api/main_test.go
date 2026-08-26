package main

import (
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestRunServerShutsDownOnSignal(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	server := &http.Server{Handler: http.NewServeMux()}
	signals := make(chan os.Signal, 1)
	serveStarted := make(chan struct{})
	result := make(chan error, 1)

	go func() {
		result <- runServer(server, func() error {
			close(serveStarted)
			return server.Serve(listener)
		}, signals)
	}()

	<-serveStarted
	signals <- os.Interrupt

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("runServer() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("runServer() did not return after a shutdown signal")
	}
}
