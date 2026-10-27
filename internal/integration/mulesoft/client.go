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

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return participant.Participant{}, fmt.Errorf(
			"create participant request: %w", err,
		)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return participant.Participant{}, fmt.Errorf(
			"mulesoft request failed: %w", err,
		)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return participant.Participant{}, participant.ErrNotFound

	case resp.StatusCode >= 500:
		return participant.Participant{}, fmt.Errorf(
			"%w: HTTP %d", ErrUpstream, resp.StatusCode,
		)

	case resp.StatusCode != http.StatusOK:
		return participant.Participant{}, fmt.Errorf(
			"unexpected mulesoft status: HTTP %d", resp.StatusCode,
		)
	}

	var p participant.Participant

	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return participant.Participant{}, fmt.Errorf(
			"decode participant response: %w", err,
		)
	}

	return p, nil
}
