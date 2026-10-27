package mockmulesoft

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shawfire/participant-services/internal/participant"
)

func TestGetParticipant(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/participants/P001",
		nil,
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var p participant.Participant

	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if p.ID != "P001" || p.FirstName != "Alex" {
		t.Fatalf("unexpected participant: %+v", p)
	}
}

func TestParticipantNotFound(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/participants/P999",
		nil,
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
