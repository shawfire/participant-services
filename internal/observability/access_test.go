package observability

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccessLogMiddleware(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"success", http.StatusOK},
		{"not found", http.StatusNotFound},
		{"server error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer

			logger := slog.New(
				slog.NewJSONHandler(&output, nil),
			)

			handler := RequestIDMiddleware(
				AccessLogMiddleware(
					logger,
					http.HandlerFunc(func(
						w http.ResponseWriter,
						r *http.Request,
					) {
						w.WriteHeader(tt.status)
					}),
				),
			)

			request := httptest.NewRequest(
				http.MethodGet,
				"/graphql/portal",
				nil,
			)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != tt.status {
				t.Fatalf(
					"expected HTTP %d, got %d",
					tt.status,
					recorder.Code,
				)
			}

			var entry map[string]any

			if err := json.Unmarshal(
				output.Bytes(),
				&entry,
			); err != nil {
				t.Fatalf("invalid JSON log: %v", err)
			}

			if entry["status"] != float64(tt.status) {
				t.Errorf("incorrect logged status: %v", entry["status"])
			}

			if entry["method"] != "GET" {
				t.Errorf("incorrect method: %v", entry["method"])
			}

			if entry["path"] != "/graphql/portal" {
				t.Errorf("incorrect path: %v", entry["path"])
			}

			if entry["request_id"] != recorder.Header().Get(
				RequestIDHeader,
			) {
				t.Error("logged request ID does not match response")
			}

			if _, ok := entry["duration_ms"]; !ok {
				t.Error("duration_ms missing from access log")
			}
		})
	}
}
