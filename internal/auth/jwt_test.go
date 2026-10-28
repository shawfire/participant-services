package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTVerifier(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	const (
		issuer   = "https://issuer.example.test"
		audience = "participant-services"
	)

	verifier, err := NewVerifier(&privateKey.PublicKey, issuer, audience)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()

	validClaims := func() jwt.RegisteredClaims {
		return jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   "user-123",
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-time.Minute)),
		}
	}

	tests := []struct {
		name      string
		claims    jwt.RegisteredClaims
		key       *rsa.PrivateKey
		method    jwt.SigningMethod
		wantError bool
	}{
		{
			name:   "valid token",
			claims: validClaims(),
			key:    privateKey,
			method: jwt.SigningMethodRS256,
		},
		{
			name:      "incorrect signature",
			claims:    validClaims(),
			key:       otherKey,
			method:    jwt.SigningMethodRS256,
			wantError: true,
		},
		{
			name: "expired token",
			claims: func() jwt.RegisteredClaims {
				c := validClaims()
				c.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Hour))
				return c
			}(),
			key:       privateKey,
			method:    jwt.SigningMethodRS256,
			wantError: true,
		},
		{
			name: "incorrect issuer",
			claims: func() jwt.RegisteredClaims {
				c := validClaims()
				c.Issuer = "https://untrusted.example.test"
				return c
			}(),
			key:       privateKey,
			method:    jwt.SigningMethodRS256,
			wantError: true,
		},
		{
			name: "incorrect audience",
			claims: func() jwt.RegisteredClaims {
				c := validClaims()
				c.Audience = jwt.ClaimStrings{"other-service"}
				return c
			}(),
			key:       privateKey,
			method:    jwt.SigningMethodRS256,
			wantError: true,
		},
		{
			name: "missing expiration",
			claims: func() jwt.RegisteredClaims {
				c := validClaims()
				c.ExpiresAt = nil
				return c
			}(),
			key:       privateKey,
			method:    jwt.SigningMethodRS256,
			wantError: true,
		},
		{
			name: "future not-before",
			claims: func() jwt.RegisteredClaims {
				c := validClaims()
				c.NotBefore = jwt.NewNumericDate(now.Add(time.Hour))
				return c
			}(),
			key:       privateKey,
			method:    jwt.SigningMethodRS256,
			wantError: true,
		},
		{
			name: "missing subject",
			claims: func() jwt.RegisteredClaims {
				c := validClaims()
				c.Subject = ""
				return c
			}(),
			key:       privateKey,
			method:    jwt.SigningMethodRS256,
			wantError: true,
		},
		{
			name:      "incorrect signing algorithm",
			claims:    validClaims(),
			key:       privateKey,
			method:    jwt.SigningMethodRS512,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := jwt.NewWithClaims(tt.method, tt.claims)

			signed, err := token.SignedString(tt.key)
			if err != nil {
				t.Fatal(err)
			}

			claims, err := verifier.Verify(signed)

			if tt.wantError {
				if err == nil {
					t.Fatal("expected token rejection")
				}
				return
			}

			if err != nil {
				t.Fatalf("valid token rejected: %v", err)
			}

			if claims.Subject != "user-123" {
				t.Fatalf("unexpected subject: %q", claims.Subject)
			}
		})
	}
}

func TestNewVerifierRejectsInvalidConfiguration(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		key      *rsa.PublicKey
		issuer   string
		audience string
	}{
		{"missing key", nil, "issuer", "audience"},
		{"missing issuer", &key.PublicKey, "", "audience"},
		{"missing audience", &key.PublicKey, "issuer", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewVerifier(
				tt.key,
				tt.issuer,
				tt.audience,
			); err == nil {
				t.Fatal("expected invalid configuration to be rejected")
			}
		})
	}
}

func TestJWTVerifierRejectsTamperedSignature(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	const (
		issuer   = "https://issuer.example.test"
		audience = "participant-services"
	)

	verifier, err := NewVerifier(
		&privateKey.PublicKey,
		issuer,
		audience,
	)
	if err != nil {
		t.Fatal(err)
	}

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

	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	parts := strings.Split(signed, ".")
	if len(parts) != 3 {
		t.Fatal("expected three JWT segments")
	}

	// Change one character in the encoded signature.
	signature := []byte(parts[2])
	if signature[0] == 'A' {
		signature[0] = 'B'
	} else {
		signature[0] = 'A'
	}

	parts[2] = string(signature)
	tampered := strings.Join(parts, ".")

	if _, err := verifier.Verify(tampered); err == nil {
		t.Fatal("tampered JWT signature was accepted")
	}
}
