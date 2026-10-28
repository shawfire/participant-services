package auth

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

const maxJWTLength = 8 * 1024

// CachedJWKSVerifier verifies JWTs using a refreshable JWKS cache.
type CachedJWKSVerifier struct {
	cache    *JWKSCache
	issuer   string
	audience string
}

func NewCachedJWKSVerifier(
	cache *JWKSCache,
	issuer string,
	audience string,
) (*CachedJWKSVerifier, error) {
	if cache == nil || issuer == "" || audience == "" {
		return nil, errors.New("invalid cached JWKS verifier configuration")
	}

	return &CachedJWKSVerifier{
		cache:    cache,
		issuer:   issuer,
		audience: audience,
	}, nil
}

// Verify validates a JWT using trusted, cached JWKS keys.
func (v *CachedJWKSVerifier) Verify(
	ctx context.Context,
	tokenString string,
) (*Claims, error) {
	if len(tokenString) == 0 || len(tokenString) > maxJWTLength {
		return nil, errors.New("invalid JWT length")
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	keys, err := v.cache.Get(ctx)
	if err != nil {
		return nil, err
	}

	// Inspect the unverified header only to select a trusted key.
	// No claims or signature are trusted at this stage.
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
	)

	unverified, _, err := parser.ParseUnverified(
		tokenString,
		&Claims{},
	)
	if err != nil {
		return nil, err
	}

	if unverified.Method.Alg() != jwt.SigningMethodRS256.Alg() {
		return nil, errors.New("unsupported JWT signing algorithm")
	}

	kid, ok := unverified.Header["kid"].(string)
	if !ok || kid == "" {
		return nil, errors.New("JWT key ID is required")
	}

	// Only a genuinely unknown key ID can trigger early refresh.
	// An existing key with an invalid signature cannot.
	if _, err := keys.FindPublicKey(kid); err != nil {
		if !errors.Is(err, ErrUnknownKeyID) {
			return nil, err
		}

		keys, err = v.cache.RefreshForUnknownKey(ctx, kid)
		if err != nil {
			return nil, err
		}
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return keys.FindPublicKey(kid)
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
