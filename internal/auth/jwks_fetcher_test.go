package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJWKSFetcher(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		wantError bool
	}{
		{
			name:   "successful retrieval",
			status: http.StatusOK,
			body:   `{"keys":[{"kid":"key-1","kty":"RSA"}]}`,
		},
		{
			name:      "server error",
			status:    http.StatusInternalServerError,
			body:      `error`,
			wantError: true,
		},
		{
			name:      "malformed JSON",
			status:    http.StatusOK,
			body:      `{invalid`,
			wantError: true,
		},
		{
			name:      "empty key set",
			status:    http.StatusOK,
			body:      `{"keys":[]}`,
			wantError: true,
		},
		{
			name:      "oversized response",
			status:    http.StatusOK,
			body:      strings.Repeat("x", maxJWKSResponse+1),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.status)
					_, _ = fmt.Fprint(w, tt.body)
				}),
			)
			defer server.Close()

			fetcher, err := NewJWKSFetcher(server.URL, true)
			if err != nil {
				t.Fatal(err)
			}

			keys, err := fetcher.Fetch(context.Background())

			if tt.wantError {
				if err == nil {
					t.Fatal("expected JWKS retrieval to fail")
				}
				return
			}

			if err != nil {
				t.Fatalf("JWKS retrieval failed: %v", err)
			}

			if len(keys.Keys) != 1 ||
				keys.Keys[0].KeyID != "key-1" {
				t.Fatal("unexpected JWKS response")
			}
		})
	}
}

func TestJWKSFetcherRejectsRedirect(t *testing.T) {
	destinationCalled := false

	destination := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			destinationCalled = true
			w.WriteHeader(http.StatusOK)
		}),
	)
	defer destination.Close()

	source := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(
				w,
				r,
				destination.URL,
				http.StatusFound,
			)
		}),
	)
	defer source.Close()

	fetcher, err := NewJWKSFetcher(source.URL, true)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := fetcher.Fetch(context.Background()); err == nil {
		t.Fatal("expected redirect to be rejected")
	}

	if destinationCalled {
		t.Fatal("JWKS fetcher followed an HTTP redirect")
	}
}

func TestJWKSFetcherRequiresHTTPS(t *testing.T) {
	if _, err := NewJWKSFetcher(
		"http://example.com/keys",
		false,
	); err == nil {
		t.Fatal("expected HTTP endpoint to be rejected")
	}

	if _, err := NewJWKSFetcher(
		"https://example.com/keys",
		false,
	); err != nil {
		t.Fatalf("valid HTTPS endpoint rejected: %v", err)
	}
}

func TestJWKSFetcherHonoursCancellation(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}),
	)
	defer server.Close()

	fetcher, err := NewJWKSFetcher(server.URL, true)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := fetcher.Fetch(ctx); err == nil {
		t.Fatal("expected cancelled request to fail")
	}
}
