package auth

import (
	"crypto/rsa"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents identity claims in a JWT.
type Claims struct {
	jwt.RegisteredClaims
	ParticipantID string `json:"participant_id,omitempty"`
}

// Verifier validates JWTs signed by a trusted RSA private key.
type Verifier struct {
	publicKey *rsa.PublicKey
	issuer    string
	audience  string
}

func NewVerifier(
	publicKey *rsa.PublicKey,
	issuer string,
	audience string,
) (*Verifier, error) {
	if publicKey == nil || issuer == "" || audience == "" {
		return nil, errors.New("invalid JWT verifier configuration")
	}

	return &Verifier{
		publicKey: publicKey,
		issuer:    issuer,
		audience:  audience,
	}, nil
}

func (v *Verifier) Verify(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return v.publicKey, nil
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
