package mulesoft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shawfire/participant-services/internal/participant"
)

var ErrUpstream = errors.New("mulesoft upstream error")

const (
	maxAttempts       = 3
	initialRetryDelay = 100 * time.Millisecond
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

var _ participant.Repository = (*Client)(nil)

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) GetParticipant(
	ctx context.Context,
	id string,
) (participant.Participant, error) {
	endpoint := c.baseURL + "/api/participants/" + url.PathEscape(id)

	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return participant.Participant{}, err
		}

		p, retry, err := c.getOnce(ctx, endpoint)
		if err == nil {
			return p, nil
		}

		lastErr = err

		if !retry || attempt == maxAttempts {
			return participant.Participant{}, err
		}

		delay := initialRetryDelay * time.Duration(1<<(attempt-1))
		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():
			timer.Stop()
			return participant.Participant{}, ctx.Err()

		case <-timer.C:
			// Continue to the next attempt.
		}
	}

	return participant.Participant{}, lastErr
}

func (c *Client) getOnce(
	ctx context.Context,
	endpoint string,
) (participant.Participant, bool, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return participant.Participant{}, false,
			fmt.Errorf("create participant request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return participant.Participant{}, false, ctx.Err()
		}

		var timeoutErr interface{ Timeout() bool }
		if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
			return participant.Participant{}, false,
				fmt.Errorf("mulesoft request timeout: %w", err)
		}

		return participant.Participant{}, true,
			fmt.Errorf("mulesoft request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return participant.Participant{}, false,
			participant.ErrNotFound

	case http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return participant.Participant{}, true,
			fmt.Errorf("%w: HTTP %d", ErrUpstream, resp.StatusCode)
	}

	if resp.StatusCode >= 500 {
		return participant.Participant{}, false,
			fmt.Errorf("%w: HTTP %d", ErrUpstream, resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return participant.Participant{}, false,
			fmt.Errorf(
				"unexpected mulesoft status: HTTP %d",
				resp.StatusCode,
			)
	}

	var p participant.Participant

	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return participant.Participant{}, false,
			fmt.Errorf("decode participant response: %w", err)
	}

	return p, false, nil
}
