package main

import (
	"log"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/shawfire/participant-services/graph/portal"
	"github.com/shawfire/participant-services/graph/portal/generated"
	"github.com/shawfire/participant-services/internal/participant"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	participantService := participant.NewService()

	portalAPI := handler.New(
		generated.NewExecutableSchema(generated.Config{
			Resolvers: &portal.Resolver{
				Participants: participantService,
			},
		}),
	)

	portalAPI.AddTransport(transport.Options{})
	portalAPI.AddTransport(transport.POST{})

	mux.Handle("POST /graphql/portal", portalAPI)

	// Development-only GraphQL Playground.
	mux.Handle(
		"GET /playground/portal",
		playground.Handler("Portal GraphQL", "/graphql/portal"),
	)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Participant Services API listening on :8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
