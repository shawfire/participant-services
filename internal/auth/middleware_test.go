package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireBearer(t *testing.T) {
	tests := []struct {
		name          string
		authorization string
		expected      int
	}{
		{"missing token", "", http.StatusUnauthorized},
		{"basic authentication", "Basic abc", http.StatusUnauthorized},
		{"bearer without token", "Bearer", http.StatusUnauthorized},
		{"extra fields", "Bearer abc def", http.StatusUnauthorized},
		{"bearer token present", "Bearer test-token", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequireBearer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			)

			req := httptest.NewRequest(http.MethodPost, "/graphql/portal", nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expected {
				t.Fatalf(
					"expected HTTP %d, got %d",
					tt.expected,
					rec.Code,
				)
			}
		})
	}
}
