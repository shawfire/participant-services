// Package identity carries identity established by trusted authentication code.
package identity

import "context"

// Principal contains verified caller identity and its issuer-managed mapping.
// ParticipantID must come from trusted claims, never a caller-supplied header.
type Principal struct {
	Subject       string
	ParticipantID string
}

type contextKey struct{}

// WithPrincipal is for trusted adapters after authentication, and test fixtures.
func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, contextKey{}, principal)
}

// FromContext returns a value copy so downstream callers cannot mutate identity.
func FromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(contextKey{}).(Principal)
	return principal, ok
}
