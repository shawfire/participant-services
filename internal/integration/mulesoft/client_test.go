package mulesoft

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shawfire/participant-services/internal/mockmulesoft"
	"github.com/shawfire/participant-services/internal/participant"
)

func TestGetParticipant(t *testing.T) {
	server := httptest.NewServer(mockmulesoft.NewHandler())
	defer server.Close()

	client := NewClient(server.URL, time.Second)

	p, err := client.GetParticipant(context.Background(), "P001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.ID != "P001" ||
		p.FirstName != "Alex" ||
		p.LastName != "Taylor" {
		t.Fatalf("unexpected participant: %+v", p)
	}
}

func TestParticipantNotFound(t *testing.T) {
	server := httptest.NewServer(mockmulesoft.NewHandler())
	defer server.Close()

	client := NewClient(server.URL, time.Second)

	_, err := client.GetParticipant(context.Background(), "P999")

	if !errors.Is(err, participant.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpstreamServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
		},
	))
	defer server.Close()

	client := NewClient(server.URL, time.Second)

	_, err := client.GetParticipant(context.Background(), "P001")

	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("expected ErrUpstream, got %v", err)
	}
}

func TestRequestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
				return
			case <-time.After(200 * time.Millisecond):
				w.WriteHeader(http.StatusOK)
			}
		},
	))
	defer server.Close()

	client := NewClient(server.URL, 20*time.Millisecond)

	_, err := client.GetParticipant(context.Background(), "P001")

	if err == nil {
		t.Fatal("expected timeout error")
	}

	var netErr interface{ Timeout() bool }
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestCancelledContext(t *testing.T) {
	server := httptest.NewServer(mockmulesoft.NewHandler())
	defer server.Close()

	client := NewClient(server.URL, time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetParticipant(ctx, "P001")

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestNetworkFailure(t *testing.T) {
	server := httptest.NewServer(mockmulesoft.NewHandler())
	serverURL := server.URL
	server.Close()

	client := NewClient(serverURL, 100*time.Millisecond)

	_, err := client.GetParticipant(context.Background(), "P001")

	if err == nil {
		t.Fatal("expected network error")
	}
}
