package auth

import (
	"context"
	"errors"
	"sync"
	"time"
)

// JWKSCache stores public keys retrieved from a trusted endpoint.
type JWKSCache struct {
	fetcher *JWKSFetcher
	ttl     time.Duration

	mu                 sync.Mutex
	keys               *JWKS
	expiresAt          time.Time
	lastRefreshAttempt time.Time
	minRefreshInterval time.Duration
}

func NewJWKSCache(
	fetcher *JWKSFetcher,
	ttl time.Duration,
) (*JWKSCache, error) {
	if fetcher == nil || ttl <= 0 {
		return nil, errors.New("invalid JWKS cache configuration")
	}

	return &JWKSCache{
		fetcher:            fetcher,
		ttl:                ttl,
		minRefreshInterval: 30 * time.Second,
	}, nil
}

// Get returns cached keys or retrieves fresh keys after expiry.
func (c *JWKSCache) Get(ctx context.Context) (*JWKS, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if c.keys != nil && time.Now().Before(c.expiresAt) {
		return cloneJWKS(c.keys), nil
	}

	keys, err := c.fetcher.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	if err := ValidateJWKS(keys); err != nil {
		return nil, err
	}

	c.keys = cloneJWKS(keys)
	c.expiresAt = time.Now().Add(c.ttl)

	return cloneJWKS(c.keys), nil
}

// RefreshForUnknownKey attempts an early JWKS refresh when a JWT
// references a key ID that is not present in the current cache.
//
// Refresh attempts are rate-limited to protect the upstream
// identity provider from repeated requests.
func (c *JWKSCache) RefreshForUnknownKey(
	ctx context.Context,
	kid string,
) (*JWKS, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if kid == "" {
		return nil, errors.New("JWT key ID is required")
	}

	// Another request may already have refreshed the cache.
	if c.keys != nil && time.Now().Before(c.expiresAt) {
		if _, err := c.keys.FindPublicKey(kid); err == nil {
			return cloneJWKS(c.keys), nil
		}
	}

	now := time.Now()

	if !c.lastRefreshAttempt.IsZero() &&
		now.Sub(c.lastRefreshAttempt) < c.minRefreshInterval {
		return nil, errors.New("JWKS early refresh rate limited")
	}

	// Record the attempt before making the network request.
	c.lastRefreshAttempt = now

	keys, err := c.fetcher.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	if err := ValidateJWKS(keys); err != nil {
		return nil, err
	}

	c.keys = cloneJWKS(keys)
	c.expiresAt = time.Now().Add(c.ttl)

	return cloneJWKS(c.keys), nil
}
