package portal_test

import (
	"github.com/shawfire/participant-services/internal/identity"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	"github.com/shawfire/participant-services/graph/portal"
	"github.com/shawfire/participant-services/graph/portal/generated"
	"github.com/shawfire/participant-services/internal/participant"
)

func TestPortalParticipantQuery(t *testing.T) {
	api := handler.New(
		generated.NewExecutableSchema(generated.Config{
			Resolvers: &portal.Resolver{
				Participants: participant.NewService(),
			},
		}),
	)

	api.AddTransport(transport.POST{})

	body := `{"query":"{ participant(id: \"P001\") { id firstName lastName } }"}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/graphql/portal",
		strings.NewReader(body),
	)
	request = request.WithContext(identity.WithPrincipal(request.Context(), identity.Principal{Subject: "test-subject", ParticipantID: "P001"}))
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	api.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", response.Code)
	}

	result := response.Body.String()

	if !strings.Contains(result, `"firstName":"Alex"`) {
		t.Fatalf("unexpected GraphQL response: %s", result)
	}

	if strings.Contains(result, `"errors"`) {
		t.Fatalf("GraphQL returned errors: %s", result)
	}
}
