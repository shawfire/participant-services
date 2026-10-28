package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	jwksRequestTimeout = 3 * time.Second
	maxJWKSResponse    = 1 << 20 // 1 MiB
)

// JWKSFetcher retrieves public keys from a configured endpoint.
type JWKSFetcher struct {
	endpoint string
	client   *http.Client
}

// NewJWKSFetcher configures a trusted JWKS endpoint.
// HTTPS is required except for local test servers.
func NewJWKSFetcher(
	endpoint string,
	allowHTTPForTests bool,
) (*JWKSFetcher, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" ||
		parsed.User != nil || parsed.Fragment != "" {
		return nil, errors.New("invalid JWKS endpoint")
	}

	if parsed.Scheme != "https" &&
		!(allowHTTPForTests && parsed.Scheme == "http") {
		return nil, errors.New("JWKS endpoint must use HTTPS")
	}

	return &JWKSFetcher{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: jwksRequestTimeout,
			CheckRedirect: func(
				req *http.Request,
				via []*http.Request,
			) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

// Fetch retrieves and parses a JWKS document.
func (f *JWKSFetcher) Fetch(ctx context.Context) (*JWKS, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		f.endpoint,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("JWKS request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"JWKS endpoint returned HTTP %d",
			resp.StatusCode,
		)
	}

	body, err := io.ReadAll(
		io.LimitReader(resp.Body, maxJWKSResponse+1),
	)
	if err != nil {
		return nil, err
	}

	if len(body) > maxJWKSResponse {
		return nil, errors.New("JWKS response exceeds size limit")
	}

	keys, err := ParseJWKS(body)
	if err != nil {
		return nil, fmt.Errorf("invalid JWKS response: %w", err)
	}

	return keys, nil
}
