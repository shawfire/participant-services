package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// JWKSVerifier validates JWTs using public keys from a trusted JWKS.
type JWKSVerifier struct {
	keys     *JWKS
	issuer   string
	audience string
}

func NewJWKSVerifier(
	keys *JWKS,
	issuer string,
	audience string,
) (*JWKSVerifier, error) {
	if keys == nil || len(keys.Keys) == 0 ||
		issuer == "" || audience == "" {
		return nil, errors.New("invalid JWKS verifier configuration")
	}

	return &JWKSVerifier{
		keys:     keys,
		issuer:   issuer,
		audience: audience,
	}, nil
}

func (v *JWKSVerifier) Verify(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			kid, ok := token.Header["kid"].(string)
			if !ok || kid == "" {
				return nil, errors.New("JWT key ID is required")
			}

			return v.keys.FindPublicKey(kid)
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid JWT")
	}

	if claims.Subject == "" {
		return nil, errors.New("JWT subject is required")
	}

	return claims, nil
}
