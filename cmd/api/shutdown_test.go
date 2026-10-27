package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestGracefulShutdownWaitsForActiveRequest(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	requestFinished := make(chan struct{})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /slow", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		close(started)
		<-release

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("completed"))
		close(requestFinished)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := &http.Server{
		Handler: mux,
	}

	var serveWG sync.WaitGroup
	serveWG.Add(1)

	go func() {
		defer serveWG.Done()
		_ = server.Serve(listener)
	}()

	defer func() {
		_ = server.Close()
		serveWG.Wait()
	}()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	responseResult := make(chan error, 1)

	go func() {
		resp, err := client.Get(
			"http://" + listener.Addr().String() + "/slow",
		)
		if err != nil {
			responseResult <- err
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			responseResult <- err
			return
		}

		if resp.StatusCode != http.StatusOK {
			responseResult <- fmt.Errorf(
				"unexpected status: %d",
				resp.StatusCode,
			)
			return
		}

		if string(body) != "completed" {
			responseResult <- fmt.Errorf(
				"unexpected body: %q",
				body,
			)
			return
		}

		responseResult <- nil
	}()

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not start")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	shutdownResult := make(chan error, 1)

	go func() {
		shutdownResult <- server.Shutdown(shutdownCtx)
	}()

	select {
	case err := <-shutdownResult:
		t.Fatalf("shutdown returned before request completed: %v", err)
	case <-time.After(100 * time.Millisecond):
		// Expected: shutdown is waiting for the active request.
	}

	close(release)

	select {
	case <-requestFinished:
	case <-time.After(3 * time.Second):
		t.Fatal("active request did not finish")
	}

	if err := <-responseResult; err != nil {
		t.Fatal(err)
	}

	if err := <-shutdownResult; err != nil {
		t.Fatalf("graceful shutdown failed: %v", err)
	}
}
