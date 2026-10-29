package mulesoft_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	"github.com/shawfire/participant-services/graph/mobile"
	mobilegenerated "github.com/shawfire/participant-services/graph/mobile/generated"

	"github.com/shawfire/participant-services/graph/portal"
	portalgenerated "github.com/shawfire/participant-services/graph/portal/generated"

	"github.com/shawfire/participant-services/internal/identity"
	"github.com/shawfire/participant-services/internal/integration/mulesoft"
	"github.com/shawfire/participant-services/internal/mockmulesoft"
	"github.com/shawfire/participant-services/internal/participant"
)

func TestGraphQLEndToEnd(t *testing.T) {
	// Start a mock MuleSoft HTTP server.
	upstream := httptest.NewServer(mockmulesoft.NewHandler())
	defer upstream.Close()

	// Inject the HTTP client into the shared Participant Service.
	client := mulesoft.NewClient(upstream.URL, time.Second)
	service := participant.NewServiceWithRepository(client)

	portalAPI := handler.New(
		portalgenerated.NewExecutableSchema(
			portalgenerated.Config{
				Resolvers: &portal.Resolver{
					Participants: service,
				},
			},
		),
	)
	portalAPI.AddTransport(transport.POST{})

	mobileAPI := handler.New(
		mobilegenerated.NewExecutableSchema(
			mobilegenerated.Config{
				Resolvers: &mobile.Resolver{
					Participants: service,
				},
			},
		),
	)
	mobileAPI.AddTransport(transport.POST{})

	mux := http.NewServeMux()
	withIdentity := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(identity.WithPrincipal(r.Context(), identity.Principal{Subject: "fixture", ParticipantID: "P001"})))
		})
	}
	mux.Handle("POST /graphql/portal", withIdentity(portalAPI))
	mux.Handle("POST /graphql/mobile", withIdentity(mobileAPI))

	api := httptest.NewServer(mux)
	defer api.Close()

	tests := []struct {
		name     string
		path     string
		query    string
		expected map[string]any
	}{
		{
			name:  "Portal participant",
			path:  "/graphql/portal",
			query: `{ participant(id: "P001") { id firstName lastName } }`,
			expected: map[string]any{
				"participant": map[string]any{
					"id":        "P001",
					"firstName": "Alex",
					"lastName":  "Taylor",
				},
			},
		},
		{
			name:  "Mobile profile",
			path:  "/graphql/mobile",
			query: `{ profile(participantId: "P001") { participantId displayName } }`,
			expected: map[string]any{
				"profile": map[string]any{
					"participantId": "P001",
					"displayName":   "Alex Taylor",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{
				"query": tt.query,
			})
			if err != nil {
				t.Fatal(err)
			}

			resp, err := api.Client().Post(
				api.URL+tt.path,
				"application/json",
				bytes.NewReader(body),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("unexpected HTTP status: %d", resp.StatusCode)
			}

			var result struct {
				Data   map[string]any `json:"data"`
				Errors []any          `json:"errors"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatal(err)
			}

			if len(result.Errors) != 0 {
				t.Fatalf("GraphQL errors: %v", result.Errors)
			}

			expectedJSON, _ := json.Marshal(tt.expected)
			actualJSON, _ := json.Marshal(result.Data)

			if !bytes.Equal(actualJSON, expectedJSON) {
				t.Fatalf(
					"unexpected response:\nexpected: %s\nactual: %s",
					expectedJSON,
					actualJSON,
				)
			}
		})
	}
}
