package auth

import (
	"errors"
	"fmt"
)

// ValidateJWKS checks every key before it enters the cache.
func ValidateJWKS(keys *JWKS) error {
	if keys == nil || len(keys.Keys) == 0 {
		return errors.New("JWKS contains no keys")
	}

	seen := make(map[string]struct{}, len(keys.Keys))

	for i, key := range keys.Keys {
		if key.KeyID == "" {
			return fmt.Errorf("JWKS key %d has no key ID", i)
		}

		if _, exists := seen[key.KeyID]; exists {
			return fmt.Errorf(
				"duplicate JWKS key ID %q",
				key.KeyID,
			)
		}
		seen[key.KeyID] = struct{}{}

		if _, err := key.PublicKey(); err != nil {
			return fmt.Errorf(
				"invalid JWKS key %q: %w",
				key.KeyID,
				err,
			)
		}
	}

	return nil
}

// cloneJWKS returns an independent copy of the key collection.
func cloneJWKS(keys *JWKS) *JWKS {
	if keys == nil {
		return nil
	}

	cloned := &JWKS{
		Keys: make([]JWK, len(keys.Keys)),
	}

	copy(cloned.Keys, keys.Keys)
	return cloned
}
