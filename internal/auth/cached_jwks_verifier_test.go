package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestCachedJWKSVerifierKeyRotation(t *testing.T) {
	key1, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	key2, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	toJWK := func(kid string, key *rsa.PublicKey) JWK {
		return JWK{
			KeyID:     kid,
			KeyType:   "RSA",
			Algorithm: "RS256",
			Use:       "sig",
			Modulus: base64.RawURLEncoding.EncodeToString(
				key.N.Bytes(),
			),
			Exponent: base64.RawURLEncoding.EncodeToString(
				big.NewInt(int64(key.E)).Bytes(),
			),
		}
	}

	toJSON := func(key JWK) []byte {
		t.Helper()

		data, err := json.Marshal(JWKS{
			Keys: []JWK{key},
		})
		if err != nil {
			t.Fatal(err)
		}

		return data
	}

	jwks1 := toJSON(toJWK("key-1", &key1.PublicKey))
	jwks2 := toJSON(toJWK("key-2", &key2.PublicKey))

	var version atomic.Int32
	version.Store(1)

	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.Header().Set("Content-Type", "application/json")

			if version.Load() == 1 {
				_, _ = w.Write(jwks1)
			} else {
				_, _ = w.Write(jwks2)
			}
		},
	))
	defer server.Close()

	fetcher, err := NewJWKSFetcher(server.URL, true)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := NewJWKSCache(fetcher, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	const (
		issuer   = "https://issuer.example.test"
		audience = "participant-services"
	)

	verifier, err := NewCachedJWKSVerifier(
		cache,
		issuer,
		audience,
	)
	if err != nil {
		t.Fatal(err)
	}

	sign := func(kid string, key *rsa.PrivateKey) string {
		t.Helper()

		now := time.Now()

		token := jwt.NewWithClaims(
			jwt.SigningMethodRS256,
			jwt.RegisteredClaims{
				Issuer:    issuer,
				Subject:   "participant-123",
				Audience:  jwt.ClaimStrings{audience},
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
		)

		token.Header["kid"] = kid

		signed, err := token.SignedString(key)
		if err != nil {
			t.Fatal(err)
		}

		return signed
	}

	token1 := sign("key-1", key1)
	token2 := sign("key-2", key2)

	ctx := context.Background()

	// Initial identity-provider key.
	claims, err := verifier.Verify(ctx, token1)
	if err != nil {
		t.Fatalf("original key rejected: %v", err)
	}

	if claims.Subject != "participant-123" {
		t.Fatalf("unexpected subject: %q", claims.Subject)
	}

	if requests.Load() != 1 {
		t.Fatalf("expected one JWKS fetch, got %d", requests.Load())
	}

	// Identity provider rotates to key-2.
	version.Store(2)

	// The unknown kid triggers an immediate JWKS refresh,
	// even though the original cache entry has not expired.
	claims, err = verifier.Verify(ctx, token2)
	if err != nil {
		t.Fatalf("rotated key rejected: %v", err)
	}

	if claims.Subject != "participant-123" {
		t.Fatalf("unexpected subject after rotation: %q", claims.Subject)
	}

	if requests.Load() != 2 {
		t.Fatalf("expected two JWKS fetches, got %d", requests.Load())
	}

	// The retired key must no longer be trusted.
	if _, err := verifier.Verify(ctx, token1); err == nil {
		t.Fatal("retired signing key was accepted")
	}
}

func TestCachedJWKSVerifierInvalidConfiguration(t *testing.T) {
	if _, err := NewCachedJWKSVerifier(
		nil,
		"issuer",
		"audience",
	); err == nil {
		t.Fatal("expected nil cache to be rejected")
	}
}

func TestCachedJWKSVerifierInvalidSignatureDoesNotRefresh(t *testing.T) {
	trustedKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	attackerKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	jwk := JWK{
		KeyID:     "trusted-key",
		KeyType:   "RSA",
		Algorithm: "RS256",
		Use:       "sig",
		Modulus: base64.RawURLEncoding.EncodeToString(
			trustedKey.PublicKey.N.Bytes(),
		),
		Exponent: base64.RawURLEncoding.EncodeToString(
			big.NewInt(int64(trustedKey.PublicKey.E)).Bytes(),
		),
	}

	jwksJSON, err := json.Marshal(JWKS{
		Keys: []JWK{jwk},
	})
	if err != nil {
		t.Fatal(err)
	}

	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(jwksJSON)
		},
	))
	defer server.Close()

	fetcher, err := NewJWKSFetcher(server.URL, true)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := NewJWKSCache(fetcher, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	const (
		issuer   = "https://issuer.example.test"
		audience = "participant-services"
	)

	verifier, err := NewCachedJWKSVerifier(
		cache,
		issuer,
		audience,
	)
	if err != nil {
		t.Fatal(err)
	}

	sign := func(key *rsa.PrivateKey) string {
		t.Helper()

		now := time.Now()

		token := jwt.NewWithClaims(
			jwt.SigningMethodRS256,
			jwt.RegisteredClaims{
				Issuer:    issuer,
				Subject:   "participant-123",
				Audience:  jwt.ClaimStrings{audience},
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
		)

		// Both tokens claim to use the trusted key.
		token.Header["kid"] = "trusted-key"

		signed, err := token.SignedString(key)
		if err != nil {
			t.Fatal(err)
		}

		return signed
	}

	ctx := context.Background()

	// A legitimate token establishes the trusted JWKS cache.
	if _, err := verifier.Verify(ctx, sign(trustedKey)); err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}

	if requests.Load() != 1 {
		t.Fatalf(
			"expected one initial JWKS request, got %d",
			requests.Load(),
		)
	}

	// An attacker signs with a different private key but
	// supplies the same trusted kid in the JWT header.
	invalidToken := sign(attackerKey)

	for i := 0; i < 20; i++ {
		if _, err := verifier.Verify(ctx, invalidToken); err == nil {
			t.Fatal("invalid JWT signature was accepted")
		}
	}

	// Invalid signatures must never trigger early refresh.
	if requests.Load() != 1 {
		t.Fatalf(
			"invalid signatures triggered JWKS refresh: %d requests",
			requests.Load(),
		)
	}

	// Failed verification must not disrupt legitimate tokens.
	if _, err := verifier.Verify(ctx, sign(trustedKey)); err != nil {
		t.Fatalf("valid token rejected after attack: %v", err)
	}
}

func TestCachedJWKSVerifierRejectsInvalidTokenLength(t *testing.T) {
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			http.Error(w, "unexpected JWKS request", http.StatusInternalServerError)
		},
	))
	defer server.Close()

	fetcher, err := NewJWKSFetcher(server.URL, true)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := NewJWKSCache(fetcher, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	verifier, err := NewCachedJWKSVerifier(
		cache,
		"https://issuer.example.test",
		"participant-services",
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "oversized token",
			token: strings.Repeat("A", maxJWTLength+1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := verifier.Verify(
				context.Background(),
				tt.token,
			)

			if err == nil {
				t.Fatal("expected invalid-length token to be rejected")
			}

			if requests.Load() != 0 {
				t.Fatalf(
					"invalid token triggered %d JWKS requests",
					requests.Load(),
				)
			}
		})
	}
}
