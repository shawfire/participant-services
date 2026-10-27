package main

import (
	"log"
	"net/http"
	"time"

	"github.com/shawfire/participant-services/internal/mockmulesoft"
)

func main() {
	server := &http.Server{
		Addr:              ":8081",
		Handler:           mockmulesoft.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Mock MuleSoft API listening on :8081")

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
