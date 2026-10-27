package participant

import (
	"context"
	"errors"
	"sync"
)

var ErrNotFound = errors.New("participant not found")

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

	if s.repository != nil {
		return s.repository.GetParticipant(ctx, id)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.participants[id]
	if !ok {
		return Participant{}, ErrNotFound
	}

	return p, nil
}
