package participant

import (
	"context"
	"errors"
	"testing"
)

func TestGetParticipant(t *testing.T) {
	service := NewService()

	p, err := service.Get(context.Background(), "P001")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if p.ID != "P001" || p.FirstName != "Alex" {
		t.Fatalf("unexpected participant: %+v", p)
	}
}

func TestGetParticipantNotFound(t *testing.T) {
	service := NewService()

	_, err := service.Get(context.Background(), "UNKNOWN")

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetParticipantCancelled(t *testing.T) {
	service := NewService()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.Get(ctx, "P001")

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestConcurrentParticipantReads(t *testing.T) {
	service := NewService()

	const workers = 100
	results := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func() {
			p, err := service.Get(context.Background(), "P001")
			if err == nil && p.ID != "P001" {
				err = errors.New("unexpected participant ID")
			}
			results <- err
		}()
	}

	for i := 0; i < workers; i++ {
		if err := <-results; err != nil {
			t.Fatalf("concurrent read failed: %v", err)
		}
	}
}
