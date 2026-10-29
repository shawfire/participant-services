package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/shawfire/participant-services/internal/identity"
)

// TokenVerifier verifies a JWT and returns its validated claims.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*Claims, error)
}

type claimsContextKey struct{}

// ClaimsFromContext retrieves claims established by verified middleware.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey{}).(*Claims)
	return claims, ok
}

// RequireJWT validates a Bearer JWT before allowing access.
func RequireJWT(
	verifier TokenVerifier,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Values("Authorization")

		if len(authorization) != 1 {
			unauthorized(w)
			return
		}

		parts := strings.Fields(authorization[0])

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			parts[1] == "" {
			unauthorized(w)
			return
		}

		claims, err := verifier.Verify(r.Context(), parts[1])
		if err != nil || claims == nil {
			unauthorized(w)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			claimsContextKey{},
			claims,
		)

		ctx = identity.WithPrincipal(ctx, identity.Principal{
			Subject:       claims.Subject,
			ParticipantID: claims.ParticipantID,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}
