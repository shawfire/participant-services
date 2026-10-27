package mobile_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	"github.com/shawfire/participant-services/graph/mobile"
	"github.com/shawfire/participant-services/graph/mobile/generated"
	"github.com/shawfire/participant-services/internal/participant"
)

func TestMobileProfileQuery(t *testing.T) {
	api := handler.New(
		generated.NewExecutableSchema(generated.Config{
			Resolvers: &mobile.Resolver{
				Participants: participant.NewService(),
			},
		}),
	)
	api.AddTransport(transport.POST{})

	body := `{"query":"{ profile(participantId: \"P001\") { participantId displayName } }"}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/graphql/mobile",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", rec.Code)
	}

	var result struct {
		Data struct {
			Profile struct {
				ParticipantID string `json:"participantId"`
				DisplayName   string `json:"displayName"`
			} `json:"profile"`
		} `json:"data"`
		Errors []json.RawMessage `json:"errors"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	if len(result.Errors) != 0 {
		t.Fatalf("GraphQL errors: %s", rec.Body.String())
	}

	if result.Data.Profile.ParticipantID != "P001" ||
		result.Data.Profile.DisplayName != "Alex Taylor" {
		t.Fatalf("unexpected profile: %+v", result.Data.Profile)
	}
}
