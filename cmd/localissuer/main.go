package main

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shawfire/participant-services/internal/auth"
)

const (
	issuer   = "https://local-issuer:8443"
	audience = "participant-services"
	keyID    = "local-dev-key"
)

func main() {
	certificate, err := tls.LoadX509KeyPair(
		"/certs/local-issuer.crt",
		"/certs/local-issuer.key",
	)
	if err != nil {
		log.Fatal(err)
	}

	privateKey, ok := certificate.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		log.Fatal("issuer requires an RSA private key")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		publicKey := &privateKey.PublicKey

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]string{
				{
					"kty": "RSA",
					"kid": keyID,
					"alg": "RS256",
					"use": "sig",
					"n": base64.RawURLEncoding.EncodeToString(
						publicKey.N.Bytes(),
					),
					"e": base64.RawURLEncoding.EncodeToString(
						big.NewInt(int64(publicKey.E)).Bytes(),
					),
				},
			},
		})
	})

	// Local testing only. Never expose this endpoint publicly.
	mux.HandleFunc("POST /dev/token", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		token := jwt.NewWithClaims(
			jwt.SigningMethodRS256,
			auth.Claims{
				ParticipantID: "P001",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    issuer,
					Subject:   "participant-123",
					Audience:  jwt.ClaimStrings{audience},
					IssuedAt:  jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
				},
			},
		)
		token.Header["kid"] = keyID

		signed, err := token.SignedString(privateKey)
		if err != nil {
			http.Error(w, "token signing failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"access_token": signed,
			"token_type":   "Bearer",
		})
	})

	server := &http.Server{
		Addr:              ":8443",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Local HTTPS identity provider listening on :8443")
	log.Fatal(server.ListenAndServeTLS(
		"/certs/local-issuer.crt",
		"/certs/local-issuer.key",
	))
}
