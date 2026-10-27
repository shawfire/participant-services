package participant

import (
	"context"
	"errors"
	"testing"
)

type stubRepository struct {
	participant Participant
	err         error
	calledID    string
}

func (s *stubRepository) GetParticipant(
	ctx context.Context,
	id string,
) (Participant, error) {
	s.calledID = id
	return s.participant, s.err
}

func TestServiceWithRepository(t *testing.T) {
	repo := &stubRepository{
		participant: Participant{
			ID:        "P003",
			FirstName: "Sam",
			LastName:  "Morgan",
		},
	}

	service := NewServiceWithRepository(repo)

	p, err := service.Get(context.Background(), "P003")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.calledID != "P003" {
		t.Fatalf("expected repository call for P003, got %q", repo.calledID)
	}

	if p.ID != "P003" || p.FirstName != "Sam" {
		t.Fatalf("unexpected participant: %+v", p)
	}
}

func TestServiceWithRepositoryNotFound(t *testing.T) {
	repo := &stubRepository{
		err: ErrNotFound,
	}

	service := NewServiceWithRepository(repo)

	_, err := service.Get(context.Background(), "P999")

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceWithRepositoryCancelledContext(t *testing.T) {
	repo := &stubRepository{}

	service := NewServiceWithRepository(repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.Get(ctx, "P001")

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if repo.calledID != "" {
		t.Fatal("repository should not be called for cancelled context")
	}
}
