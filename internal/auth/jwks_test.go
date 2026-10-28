package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"testing"
)

func TestJWKPublicKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	publicKey := &privateKey.PublicKey

	jwk := JWK{
		KeyID:     "test-key-1",
		KeyType:   "RSA",
		Algorithm: "RS256",
		Use:       "sig",
		Modulus: base64.RawURLEncoding.EncodeToString(
			publicKey.N.Bytes(),
		),
		Exponent: base64.RawURLEncoding.EncodeToString(
			big.NewInt(int64(publicKey.E)).Bytes(),
		),
	}

	reconstructed, err := jwk.PublicKey()
	if err != nil {
		t.Fatalf("failed to parse valid JWK: %v", err)
	}

	if reconstructed.N.Cmp(publicKey.N) != 0 {
		t.Fatal("RSA modulus does not match")
	}

	if reconstructed.E != publicKey.E {
		t.Fatal("RSA exponent does not match")
	}
}

func TestJWKPublicKeyRejectsInvalidKeys(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	valid := JWK{
		KeyID:     "test-key",
		KeyType:   "RSA",
		Algorithm: "RS256",
		Use:       "sig",
		Modulus: base64.RawURLEncoding.EncodeToString(
			privateKey.PublicKey.N.Bytes(),
		),
		Exponent: base64.RawURLEncoding.EncodeToString(
			big.NewInt(int64(privateKey.PublicKey.E)).Bytes(),
		),
	}

	tests := []struct {
		name   string
		modify func(*JWK)
	}{
		{
			name: "RSA modulus too small",
			modify: func(k *JWK) {
				k.Modulus = base64.RawURLEncoding.EncodeToString(
					[]byte{1, 2, 3},
				)
			},
		},
		{"unsupported key type", func(k *JWK) { k.KeyType = "EC" }},
		{"unsupported algorithm", func(k *JWK) { k.Algorithm = "RS512" }},
		{"wrong key use", func(k *JWK) { k.Use = "enc" }},
		{"invalid modulus", func(k *JWK) { k.Modulus = "!!!" }},
		{"missing modulus", func(k *JWK) { k.Modulus = "" }},
		{"invalid exponent", func(k *JWK) { k.Exponent = "!!!" }},
		{"missing exponent", func(k *JWK) { k.Exponent = "" }},
		{"even exponent", func(k *JWK) { k.Exponent = "BA" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := valid
			tt.modify(&key)

			if _, err := key.PublicKey(); err == nil {
				t.Fatal("expected invalid JWK to be rejected")
			}
		})
	}
}

func TestParseJWKS(t *testing.T) {
	data, err := json.Marshal(JWKS{
		Keys: []JWK{
			{KeyID: "key-1", KeyType: "RSA"},
			{KeyID: "key-2", KeyType: "RSA"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	keys, err := ParseJWKS(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(keys.Keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys.Keys))
	}

	if keys.Keys[0].KeyID != "key-1" {
		t.Fatal("unexpected first key ID")
	}
}

func TestParseJWKSRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"invalid JSON", `{invalid`},
		{"empty key set", `{"keys":[]}`},
		{"missing keys", `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseJWKS([]byte(tt.data)); err == nil {
				t.Fatal("expected JWKS parsing to fail")
			}
		})
	}
}

func TestJWKSFindPublicKey(t *testing.T) {
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

	keys := &JWKS{
		Keys: []JWK{
			toJWK("key-1", &key1.PublicKey),
			toJWK("key-2", &key2.PublicKey),
		},
	}

	t.Run("select first key", func(t *testing.T) {
		selected, err := keys.FindPublicKey("key-1")
		if err != nil {
			t.Fatal(err)
		}

		if selected.N.Cmp(key1.PublicKey.N) != 0 ||
			selected.E != key1.PublicKey.E {
			t.Fatal("incorrect public key selected")
		}
	})

	t.Run("select second key", func(t *testing.T) {
		selected, err := keys.FindPublicKey("key-2")
		if err != nil {
			t.Fatal(err)
		}

		if selected.N.Cmp(key2.PublicKey.N) != 0 ||
			selected.E != key2.PublicKey.E {
			t.Fatal("incorrect public key selected")
		}
	})

	t.Run("unknown key ID", func(t *testing.T) {
		if _, err := keys.FindPublicKey("unknown"); err == nil {
			t.Fatal("expected unknown key ID to be rejected")
		}
	})

	t.Run("missing key ID", func(t *testing.T) {
		if _, err := keys.FindPublicKey(""); err == nil {
			t.Fatal("expected missing key ID to be rejected")
		}
	})

	t.Run("duplicate key ID", func(t *testing.T) {
		duplicates := &JWKS{
			Keys: []JWK{
				toJWK("duplicate", &key1.PublicKey),
				toJWK("duplicate", &key2.PublicKey),
			},
		}

		if _, err := duplicates.FindPublicKey("duplicate"); err == nil {
			t.Fatal("expected duplicate key IDs to be rejected")
		}
	})

	t.Run("nil JWKS", func(t *testing.T) {
		var empty *JWKS

		if _, err := empty.FindPublicKey("key-1"); err == nil {
			t.Fatal("expected nil JWKS to be rejected")
		}
	})
}

func TestJWKSUnknownKeyErrorClassification(t *testing.T) {
	keys := &JWKS{
		Keys: []JWK{
			{KeyID: "known-key"},
		},
	}

	tests := []struct {
		name        string
		keys        *JWKS
		kid         string
		wantUnknown bool
	}{
		{
			name:        "unknown key ID",
			keys:        keys,
			kid:         "unknown-key",
			wantUnknown: true,
		},
		{
			name: "missing key ID",
			keys: keys,
			kid:  "",
		},
		{
			name: "nil JWKS",
			keys: nil,
			kid:  "unknown-key",
		},
		{
			name: "duplicate key ID",
			keys: &JWKS{
				Keys: []JWK{
					{KeyID: "duplicate"},
					{KeyID: "duplicate"},
				},
			},
			kid: "duplicate",
		},
		{
			name: "invalid existing key",
			keys: keys,
			kid:  "known-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.keys.FindPublicKey(tt.kid)
			if err == nil {
				t.Fatal("expected key lookup to fail")
			}

			gotUnknown := errors.Is(err, ErrUnknownKeyID)

			if gotUnknown != tt.wantUnknown {
				t.Fatalf(
					"errors.Is(err, ErrUnknownKeyID) = %v, want %v; error: %v",
					gotUnknown,
					tt.wantUnknown,
					err,
				)
			}
		})
	}
}
