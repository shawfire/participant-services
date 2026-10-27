package mulesoft

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shawfire/participant-services/internal/participant"
)

func TestRetryRecoversAfterTransientFailures(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			attempt := attempts.Add(1)

			if attempt < maxAttempts {
				http.Error(
					w,
					"temporary upstream failure",
					http.StatusServiceUnavailable,
				)
				return
			}

			w.Header().Set("Content-Type", "application/json")

			_ = json.NewEncoder(w).Encode(participant.Participant{
				ID:        "P001",
				FirstName: "Alex",
				LastName:  "Taylor",
			})
		},
	))
	defer server.Close()

	client := NewClient(server.URL, time.Second)

	got, err := client.GetParticipant(context.Background(), "P001")
	if err != nil {
		t.Fatalf("expected successful recovery, got: %v", err)
	}

	if got.ID != "P001" {
		t.Errorf("expected participant P001, got %q", got.ID)
	}

	if actual := attempts.Load(); actual != maxAttempts {
		t.Errorf(
			"expected %d attempts, got %d",
			maxAttempts,
			actual,
		)
	}
}

func TestRetryStopsAfterMaximumAttempts(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			attempts.Add(1)
			http.Error(
				w,
				"upstream unavailable",
				http.StatusServiceUnavailable,
			)
		},
	))
	defer server.Close()

	client := NewClient(server.URL, time.Second)

	_, err := client.GetParticipant(context.Background(), "P001")
	if err == nil {
		t.Fatal("expected an upstream error")
	}

	if actual := attempts.Load(); actual != maxAttempts {
		t.Errorf(
			"expected %d attempts, got %d",
			maxAttempts,
			actual,
		)
	}
}

func TestNotFoundIsNotRetried(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			attempts.Add(1)
			http.NotFound(w, r)
		},
	))
	defer server.Close()

	client := NewClient(server.URL, time.Second)

	_, err := client.GetParticipant(context.Background(), "UNKNOWN")
	if err != participant.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if actual := attempts.Load(); actual != 1 {
		t.Errorf("expected 1 attempt, got %d", actual)
	}
}

func TestCancellationDuringRetryBackoff(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			attempts.Add(1)
			http.Error(
				w,
				"upstream unavailable",
				http.StatusServiceUnavailable,
			)
		},
	))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	client := NewClient(server.URL, time.Second)

	done := make(chan error, 1)

	go func() {
		_, err := client.GetParticipant(ctx, "P001")
		done <- err
	}()

	// Wait until the initial request has reached the server.
	deadline := time.After(time.Second)

	for attempts.Load() == 0 {
		select {
		case <-deadline:
			cancel()
			t.Fatal("initial request did not reach the server")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, got %v", err)
		}

	case <-time.After(time.Second):
		t.Fatal("retry did not stop after cancellation")
	}

	if actual := attempts.Load(); actual != 1 {
		t.Errorf("expected 1 attempt, got %d", actual)
	}
}
