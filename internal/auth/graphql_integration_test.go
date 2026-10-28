package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/golang-jwt/jwt/v5"

	"github.com/shawfire/participant-services/graph/mobile"
	mobilegenerated "github.com/shawfire/participant-services/graph/mobile/generated"
	"github.com/shawfire/participant-services/graph/portal"
	portalgenerated "github.com/shawfire/participant-services/graph/portal/generated"
	"github.com/shawfire/participant-services/internal/health"
	"github.com/shawfire/participant-services/internal/integration/mulesoft"
	"github.com/shawfire/participant-services/internal/participant"
)

func TestGraphQLJWTIntegration(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	const (
		issuer   = "https://test-issuer.example"
		audience = "participant-services"
		keyID    = "integration-test-key"
	)

	jwksBody, err := json.Marshal(JWKS{
		Keys: []JWK{
			{
				KeyID:     keyID,
				KeyType:   "RSA",
				Algorithm: "RS256",
				Use:       "sig",
				Modulus: base64.RawURLEncoding.EncodeToString(
					privateKey.PublicKey.N.Bytes(),
				),
				Exponent: base64.RawURLEncoding.EncodeToString(
					big.NewInt(int64(privateKey.PublicKey.E)).Bytes(),
				),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	identityProvider := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/.well-known/jwks.json" {
				http.NotFound(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(jwksBody)
		}),
	)
	defer identityProvider.Close()

	fetcher, err := NewJWKSFetcher(
		identityProvider.URL+"/.well-known/jwks.json",
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Trust the test server's certificate, without disabling TLS verification.
	fetcher.client = identityProvider.Client()

	cache, err := NewJWKSCache(fetcher, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	verifier, err := NewCachedJWKSVerifier(cache, issuer, audience)
	if err != nil {
		t.Fatal(err)
	}

	// __typename does not invoke participant resolvers or MuleSoft.
	// Supply a non-nil repository to satisfy the service constructor.
	mulesoftClient := mulesoft.NewClient(
		"http://127.0.0.1:1",
		2*time.Second,
	)

	participantService := participant.NewServiceWithRepository(
		mulesoftClient,
	)

	portalAPI := handler.New(
		portalgenerated.NewExecutableSchema(
			portalgenerated.Config{
				Resolvers: &portal.Resolver{
					Participants: participantService,
				},
			},
		),
	)
	portalAPI.AddTransport(transport.POST{})

	mobileAPI := handler.New(
		mobilegenerated.NewExecutableSchema(
			mobilegenerated.Config{
				Resolvers: &mobile.Resolver{
					Participants: participantService,
				},
			},
		),
	)
	mobileAPI.AddTransport(transport.POST{})

	mux := http.NewServeMux()
	mux.Handle("POST /graphql/portal", RequireJWT(verifier, portalAPI))
	mux.Handle("POST /graphql/mobile", RequireJWT(verifier, mobileAPI))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("GET /readyz", health.NewReadiness())

	api := httptest.NewServer(mux)
	defer api.Close()

	signToken := func(claims jwt.RegisteredClaims) string {
		t.Helper()

		token := jwt.NewWithClaims(
			jwt.SigningMethodRS256,
			claims,
		)
		token.Header["kid"] = keyID

		signed, err := token.SignedString(privateKey)
		if err != nil {
			t.Fatal(err)
		}
		return signed
	}

	now := time.Now()

	validToken := signToken(jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   "participant-123",
		Audience:  jwt.ClaimStrings{audience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
	})

	expiredToken := signToken(jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   "participant-123",
		Audience:  jwt.ClaimStrings{audience},
		IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
	})

	wrongAudienceToken := signToken(jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   "participant-123",
		Audience:  jwt.ClaimStrings{"different-service"},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
	})

	for _, path := range []string{
		"/graphql/portal",
		"/graphql/mobile",
	} {
		t.Run(path, func(t *testing.T) {
			tests := []struct {
				name       string
				token      string
				wantStatus int
			}{
				{"valid JWT", validToken, http.StatusOK},
				{"missing JWT", "", http.StatusUnauthorized},
				{"expired JWT", expiredToken, http.StatusUnauthorized},
				{"wrong audience", wrongAudienceToken, http.StatusUnauthorized},
				{"malformed JWT", "not-a-jwt", http.StatusUnauthorized},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					request, err := http.NewRequestWithContext(
						context.Background(),
						http.MethodPost,
						api.URL+path,
						bytes.NewBufferString(`{"query":"{ __typename }"}`),
					)
					if err != nil {
						t.Fatal(err)
					}

					request.Header.Set("Content-Type", "application/json")
					if tt.token != "" {
						request.Header.Set("Authorization", "Bearer "+tt.token)
					}

					response, err := api.Client().Do(request)
					if err != nil {
						t.Fatal(err)
					}
					defer response.Body.Close()

					if response.StatusCode != tt.wantStatus {
						t.Fatalf(
							"HTTP status = %d, want %d",
							response.StatusCode,
							tt.wantStatus,
						)
					}
				})
			}
		})
	}

	for _, path := range []string{"/healthz", "/readyz"} {
		t.Run(path+" public", func(t *testing.T) {
			response, err := api.Client().Get(api.URL + path)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				t.Fatalf(
					"HTTP status = %d, want 200",
					response.StatusCode,
				)
			}
		})
	}
}
