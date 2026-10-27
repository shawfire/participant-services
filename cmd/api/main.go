package main

import (
	"log"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/shawfire/participant-services/graph/mobile"
	mobilegenerated "github.com/shawfire/participant-services/graph/mobile/generated"
	"github.com/shawfire/participant-services/graph/portal"
	portalgenerated "github.com/shawfire/participant-services/graph/portal/generated"
	"github.com/shawfire/participant-services/internal/participant"
)

func main() {
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	// One shared Participant Service instance.
	participantService := participant.NewService()

	// Portal GraphQL API
	portalAPI := handler.New(
		portalgenerated.NewExecutableSchema(portalgenerated.Config{
			Resolvers: &portal.Resolver{
				Participants: participantService,
			},
		}),
	)

	portalAPI.AddTransport(transport.Options{})
	portalAPI.AddTransport(transport.POST{})

	mux.Handle("POST /graphql/portal", portalAPI)

	mux.Handle(
		"GET /playground/portal",
		playground.Handler("Portal GraphQL", "/graphql/portal"),
	)

	// Mobile GraphQL API
	mobileAPI := handler.New(
		mobilegenerated.NewExecutableSchema(mobilegenerated.Config{
			Resolvers: &mobile.Resolver{
				Participants: participantService,
			},
		}),
	)

	mobileAPI.AddTransport(transport.Options{})
	mobileAPI.AddTransport(transport.POST{})

	mux.Handle("POST /graphql/mobile", mobileAPI)

	mux.Handle(
		"GET /playground/mobile",
		playground.Handler("Mobile GraphQL", "/graphql/mobile"),
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
