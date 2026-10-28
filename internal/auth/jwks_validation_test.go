package auth

import (
	"encoding/json"
	"testing"
)

func TestValidateJWKS(t *testing.T) {
	var valid JWKS

	if err := json.Unmarshal(
		[]byte(testJWKSJSON(t, "key-1")),
		&valid,
	); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		keys    *JWKS
		wantErr bool
	}{
		{
			name: "valid RSA key",
			keys: &valid,
		},
		{
			name:    "nil JWKS",
			wantErr: true,
		},
		{
			name:    "empty JWKS",
			keys:    &JWKS{},
			wantErr: true,
		},
		{
			name: "missing key ID",
			keys: &JWKS{
				Keys: []JWK{{KeyType: "RSA"}},
			},
			wantErr: true,
		},
		{
			name: "duplicate key ID",
			keys: &JWKS{
				Keys: []JWK{
					valid.Keys[0],
					valid.Keys[0],
				},
			},
			wantErr: true,
		},
		{
			name: "invalid RSA modulus",
			keys: &JWKS{
				Keys: []JWK{{
					KeyID:   "invalid",
					KeyType: "RSA",
				}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJWKS(tt.keys)

			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("valid JWKS rejected: %v", err)
			}
		})
	}
}

func TestCloneJWKS(t *testing.T) {
	original := &JWKS{
		Keys: []JWK{
			{KeyID: "original"},
		},
	}

	cloned := cloneJWKS(original)

	cloned.Keys[0].KeyID = "modified"
	cloned.Keys = append(cloned.Keys, JWK{
		KeyID: "additional",
	})

	if original.Keys[0].KeyID != "original" {
		t.Fatal("clone modified original key")
	}

	if len(original.Keys) != 1 {
		t.Fatal("clone modified original key collection")
	}

	if cloneJWKS(nil) != nil {
		t.Fatal("expected nil clone")
	}
}
