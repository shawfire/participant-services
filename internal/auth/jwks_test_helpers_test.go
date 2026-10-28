package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
)

// testJWKSJSON generates a valid JWKS document for testing.
func testJWKSJSON(t *testing.T, keyIDs ...string) string {
	t.Helper()

	keys := make([]JWK, 0, len(keyIDs))

	for _, kid := range keyIDs {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}

		keys = append(keys, JWK{
			KeyID:     kid,
			KeyType:   "RSA",
			Algorithm: "RS256",
			Use:       "sig",
			Modulus: base64.RawURLEncoding.EncodeToString(
				privateKey.PublicKey.N.Bytes(),
			),
			Exponent: base64.RawURLEncoding.EncodeToString(
				big.NewInt(int64(privateKey.PublicKey.E)).Bytes(),
			),
		})
	}

	data, err := json.Marshal(JWKS{Keys: keys})
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}
