package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type requestIDKey struct{}

const RequestIDHeader = "X-Request-ID"

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bytes [16]byte

		if _, err := rand.Read(bytes[:]); err != nil {
			http.Error(w, "request ID generation failed",
				http.StatusInternalServerError)
			return
		}

		id := hex.EncodeToString(bytes[:])
		w.Header().Set(RequestIDHeader, id)

		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
