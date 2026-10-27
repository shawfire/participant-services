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
}

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

func (s *Service) Get(
	ctx context.Context,
	id string,
) (Participant, error) {
	if err := ctx.Err(); err != nil {
		return Participant{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.participants[id]
	if !ok {
		return Participant{}, ErrNotFound
	}

	return p, nil
}
