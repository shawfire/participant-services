package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWKSVerifier(t *testing.T) {
	key1, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	key2, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	const (
		issuer   = "https://issuer.example.test"
		audience = "participant-services"
	)

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

	keys := &JWKS{
		Keys: []JWK{
			toJWK("key-1", &key1.PublicKey),
			toJWK("key-2", &key2.PublicKey),
		},
	}

	verifier, err := NewJWKSVerifier(keys, issuer, audience)
	if err != nil {
		t.Fatal(err)
	}

	sign := func(
		kid string,
		privateKey *rsa.PrivateKey,
	) string {
		t.Helper()

		now := time.Now()

		token := jwt.NewWithClaims(
			jwt.SigningMethodRS256,
			jwt.RegisteredClaims{
				Issuer:    issuer,
				Subject:   "user-123",
				Audience:  jwt.ClaimStrings{audience},
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
		)

		if kid != "" {
			token.Header["kid"] = kid
		}

		signed, err := token.SignedString(privateKey)
		if err != nil {
			t.Fatal(err)
		}

		return signed
	}

	tests := []struct {
		name      string
		kid       string
		key       *rsa.PrivateKey
		wantError bool
	}{
		{"first key", "key-1", key1, false},
		{"second key", "key-2", key2, false},
		{"wrong signing key", "key-1", key2, true},
		{"unknown key ID", "unknown", key1, true},
		{"missing key ID", "", key1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signed := sign(tt.kid, tt.key)

			claims, err := verifier.Verify(signed)

			if tt.wantError {
				if err == nil {
					t.Fatal("expected JWT rejection")
				}
				return
			}

			if err != nil {
				t.Fatalf("valid JWT rejected: %v", err)
			}

			if claims.Subject != "user-123" {
				t.Fatalf(
					"unexpected subject: %q",
					claims.Subject,
				)
			}
		})
	}

	t.Run("duplicate key ID", func(t *testing.T) {
		duplicateKeys := &JWKS{
			Keys: []JWK{
				toJWK("duplicate", &key1.PublicKey),
				toJWK("duplicate", &key2.PublicKey),
			},
		}

		duplicateVerifier, err := NewJWKSVerifier(
			duplicateKeys,
			issuer,
			audience,
		)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := duplicateVerifier.Verify(
			sign("duplicate", key1),
		); err == nil {
			t.Fatal("duplicate key ID was accepted")
		}
	})
}
