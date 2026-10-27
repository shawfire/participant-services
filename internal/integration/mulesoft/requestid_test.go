package mulesoft

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shawfire/participant-services/internal/observability"
)

func TestRequestIDPropagation(t *testing.T) {
	const expectedID = "test-request-123"

	var receivedID string

	upstream := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			receivedID = r.Header.Get(observability.RequestIDHeader)

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(
				`{"id":"123","firstName":"Jane","lastName":"Smith"}`,
			))
		},
	))
	defer upstream.Close()

	client := NewClient(upstream.URL, 2*time.Second)

	// The middleware generates the request ID in the context.
	// Capture it from the incoming HTTP response instead of
	// relying on a hard-coded value.
	var downstreamID string

	handler := observability.RequestIDMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			downstreamID = observability.RequestID(r.Context())

			_, err := client.GetParticipant(r.Context(), "123")
			if err != nil {
				t.Errorf("GetParticipant failed: %v", err)
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}

			w.WriteHeader(http.StatusOK)
		}),
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", recorder.Code)
	}

	if downstreamID == "" {
		t.Fatal("request ID missing from incoming context")
	}

	if receivedID != downstreamID {
		t.Fatalf(
			"request ID mismatch: upstream=%q, incoming=%q",
			receivedID,
			downstreamID,
		)
	}

	if recorder.Header().Get(observability.RequestIDHeader) != receivedID {
		t.Fatal("response request ID differs from MuleSoft request ID")
	}

	_ = context.Background()
	_ = expectedID
}
