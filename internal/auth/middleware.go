package auth

import (
	"net/http"
	"strings"
)

// RequireBearer checks for a Bearer token.
// This is a preliminary check, not JWT validation.
func RequireBearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		parts := strings.Fields(authorization)
		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			parts[1] == "" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(
				w,
				"unauthorized",
				http.StatusUnauthorized,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
