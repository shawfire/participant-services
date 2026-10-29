package participant

import (
	"context"
	"errors"
	"sync"

	"github.com/shawfire/participant-services/internal/identity"
)

var ErrNotFound = errors.New("participant not found")

// ErrForbidden denies access without revealing whether the requested record exists.
var ErrForbidden = errors.New("participant access forbidden")

type Participant struct {
	ID        string
	FirstName string
	LastName  string
}

type Service struct {
	mu           sync.RWMutex
	participants map[string]Participant
	repository   Repository
}

// NewService creates the original in-memory implementation.
func NewService() *Service {
	return &Service{
		participants: map[string]Participant{
			"P001": {
				ID:        "P001",
				FirstName: "Alex",
				LastName:  "Taylor",
			},
			"P002": {
				ID:        "P002",
				FirstName: "Jordan",
				LastName:  "Lee",
			},
		},
	}
}

// NewServiceWithRepository injects an external data source.
func NewServiceWithRepository(repository Repository) *Service {
	if repository == nil {
		panic("participant repository must not be nil")
	}
	return &Service{
		repository: repository,
	}
}

func (s *Service) Get(
	ctx context.Context,
	id string,
) (Participant, error) {
	if err := ctx.Err(); err != nil {
		return Participant{}, err
	}

	principal, ok := identity.FromContext(ctx)
	if !ok || principal.Subject == "" || principal.ParticipantID == "" || principal.ParticipantID != id {
		return Participant{}, ErrForbidden
	}

	if s.repository != nil {
		p, err := s.repository.GetParticipant(ctx, id)
		if err != nil {
			return Participant{}, err
		}
		// Do not expose a different record if an upstream violates its lookup contract.
		if p.ID != id {
			return Participant{}, ErrForbidden
		}
		return p, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.participants[id]
	if !ok {
		return Participant{}, ErrNotFound
	}

	return p, nil
}
