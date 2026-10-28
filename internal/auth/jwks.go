package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
)

// JWK represents the RSA public-key fields of a JSON Web Key.
type JWK struct {
	KeyID     string `json:"kid"`
	KeyType   string `json:"kty"`
	Algorithm string `json:"alg"`
	Use       string `json:"use"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

var ErrUnknownKeyID = errors.New("unknown JWKS key ID")

// PublicKey converts an RSA JWK into a Go RSA public key.
func (key JWK) PublicKey() (*rsa.PublicKey, error) {
	if key.KeyType != "RSA" {
		return nil, errors.New("unsupported JWK key type")
	}

	if key.Algorithm != "" && key.Algorithm != "RS256" {
		return nil, errors.New("unsupported JWK algorithm")
	}

	if key.Use != "" && key.Use != "sig" {
		return nil, errors.New("JWK is not a signing key")
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(key.Modulus)
	if err != nil || len(nBytes) == 0 {
		return nil, errors.New("invalid RSA modulus")
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.Exponent)
	if err != nil || len(eBytes) == 0 || len(eBytes) > 4 {
		return nil, errors.New("invalid RSA exponent")
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes).Uint64()
	if n.BitLen() < 2048 {
		return nil, errors.New("RSA modulus must be at least 2048 bits")
	}

	if n.Sign() <= 0 || e < 3 || e > 2147483647 || e%2 == 0 {
		return nil, errors.New("invalid RSA public key parameters")
	}

	return &rsa.PublicKey{
		N: n,
		E: int(e),
	}, nil
}

// ParseJWKS decodes a JSON Web Key Set.
func ParseJWKS(data []byte) (*JWKS, error) {
	var keys JWKS

	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, err
	}

	if len(keys.Keys) == 0 {
		return nil, errors.New("JWKS contains no keys")
	}

	return &keys, nil
}

// FindPublicKey selects a signing key by its exact key ID.
// Missing, unknown and duplicate key IDs are rejected.
func (keys *JWKS) FindPublicKey(kid string) (*rsa.PublicKey, error) {
	if keys == nil {
		return nil, errors.New("nil JWKS")
	}

	if kid == "" {
		return nil, errors.New("missing JWK key ID")
	}

	var selected *JWK

	for i := range keys.Keys {
		if keys.Keys[i].KeyID != kid {
			continue
		}

		if selected != nil {
			return nil, errors.New("duplicate JWK key ID")
		}

		selected = &keys.Keys[i]
	}

	if selected == nil {
		return nil, fmt.Errorf("%w: %q", ErrUnknownKeyID, kid)
	}

	return selected.PublicKey()
}
