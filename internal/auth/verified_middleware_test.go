package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

type stubTokenVerifier struct {
	claims *Claims
	err    error
	calls  int
	token  string
}

func (v *stubTokenVerifier) Verify(
	_ context.Context,
	token string,
) (*Claims, error) {
	v.calls++
	v.token = token
	return v.claims, v.err
}

func TestRequireJWT(t *testing.T) {
	tests := []struct {
		name        string
		headers     []string
		claims      *Claims
		verifyError error
		wantStatus  int
		wantCalls   int
	}{
		{
			name:       "valid JWT",
			headers:    []string{"Bearer valid-token"},
			claims:     &Claims{RegisteredClaims: jwt.RegisteredClaims{Subject: "participant-123"}},
			wantStatus: http.StatusOK,
			wantCalls:  1,
		},
		{
			name:       "missing Authorization header",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "incorrect authentication scheme",
			headers:    []string{"Basic credentials"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "empty Bearer token",
			headers:    []string{"Bearer"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "multiple Authorization headers",
			headers:    []string{"Bearer token-1", "Bearer token-2"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:        "invalid JWT signature",
			headers:     []string{"Bearer invalid-token"},
			verifyError: errors.New("invalid signature"),
			wantStatus:  http.StatusUnauthorized,
			wantCalls:   1,
		},
		{
			name:       "verifier returns no claims",
			headers:    []string{"Bearer token"},
			wantStatus: http.StatusUnauthorized,
			wantCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := &stubTokenVerifier{
				claims: tt.claims,
				err:    tt.verifyError,
			}

			nextCalled := false

			next := http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				nextCalled = true

				claims, ok := ClaimsFromContext(r.Context())
				if !ok || claims.Subject != "participant-123" {
					t.Error("verified claims missing from request context")
				}

				w.WriteHeader(http.StatusOK)
			})

			request := httptest.NewRequest(
				http.MethodPost,
				"/graphql/portal",
				nil,
			)

			for _, header := range tt.headers {
				request.Header.Add("Authorization", header)
			}

			response := httptest.NewRecorder()

			RequireJWT(verifier, next).ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf(
					"status = %d, want %d",
					response.Code,
					tt.wantStatus,
				)
			}

			if verifier.calls != tt.wantCalls {
				t.Errorf(
					"verification calls = %d, want %d",
					verifier.calls,
					tt.wantCalls,
				)
			}

			if nextCalled != (tt.wantStatus == http.StatusOK) {
				t.Error("unexpected downstream handler execution")
			}

			if tt.wantStatus == http.StatusUnauthorized {
				if got := response.Header().Get("WWW-Authenticate"); got != "Bearer" {
					t.Errorf("WWW-Authenticate = %q, want Bearer", got)
				}
			}

			if tt.wantStatus == http.StatusOK && verifier.token != "valid-token" {
				t.Errorf("verified token = %q", verifier.token)
			}
		})
	}
}
