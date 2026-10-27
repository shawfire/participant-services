package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/shawfire/participant-services/graph/mobile"
	mobilegenerated "github.com/shawfire/participant-services/graph/mobile/generated"

	"github.com/shawfire/participant-services/graph/portal"
	portalgenerated "github.com/shawfire/participant-services/graph/portal/generated"

	"github.com/shawfire/participant-services/internal/health"
	"github.com/shawfire/participant-services/internal/integration/mulesoft"
	"github.com/shawfire/participant-services/internal/observability"
	"github.com/shawfire/participant-services/internal/participant"
)

const (
	serverAddress      = ":8080"
	defaultMuleSoftURL = "http://localhost:8081"
	muleSoftTimeout    = 2 * time.Second
	readHeaderTimeout  = 5 * time.Second
	readTimeout        = 10 * time.Second
	writeTimeout       = 15 * time.Second
	idleTimeout        = 60 * time.Second
	shutdownTimeout    = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger := observability.NewLogger()
	slog.SetDefault(logger)

	readiness := health.NewReadiness()
	mux := http.NewServeMux()

	// Liveness: the application process is running.
	mux.HandleFunc("GET /healthz", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	// Readiness: the application can accept traffic.
	mux.Handle("GET /readyz", readiness)

	mulesoftURL := os.Getenv("MULESOFT_URL")
	if mulesoftURL == "" {
		mulesoftURL = defaultMuleSoftURL
	}

	mulesoftClient := mulesoft.NewClient(
		mulesoftURL,
		muleSoftTimeout,
	)

	participantService := participant.NewServiceWithRepository(
		mulesoftClient,
	)

	portalAPI := handler.New(
		portalgenerated.NewExecutableSchema(
			portalgenerated.Config{
				Resolvers: &portal.Resolver{
					Participants: participantService,
				},
			},
		),
	)

	portalAPI.AddTransport(transport.Options{})
	portalAPI.AddTransport(transport.POST{})

	mux.Handle("POST /graphql/portal", portalAPI)
	mux.Handle(
		"GET /playground/portal",
		playground.Handler("Portal GraphQL", "/graphql/portal"),
	)

	mobileAPI := handler.New(
		mobilegenerated.NewExecutableSchema(
			mobilegenerated.Config{
				Resolvers: &mobile.Resolver{
					Participants: participantService,
				},
			},
		),
	)

	mobileAPI.AddTransport(transport.Options{})
	mobileAPI.AddTransport(transport.POST{})

	mux.Handle("POST /graphql/mobile", mobileAPI)
	mux.Handle(
		"GET /playground/mobile",
		playground.Handler("Mobile GraphQL", "/graphql/mobile"),
	)

	server := &http.Server{
		Addr: serverAddress,
		Handler: observability.RequestIDMiddleware(
			observability.AccessLogMiddleware(logger, mux),
		),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info(
			"Participant Services API starting",
			"address", serverAddress,
			"mulesoft_url", mulesoftURL,
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case <-signalCtx.Done():
		logger.Info("Shutdown signal received")
		readiness.SetReady(false)

	case err := <-serverErrors:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancel()

	logger.Info(
		"Starting graceful shutdown",
		"timeout", shutdownTimeout.String(),
	)

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	logger.Info("Graceful shutdown completed")
	return nil
}
