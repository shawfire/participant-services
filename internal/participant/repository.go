package participant

import "context"

// Repository defines how participant data is retrieved.
// Implementations may use memory, MuleSoft, or another backend.
type Repository interface {
	GetParticipant(
		ctx context.Context,
		id string,
	) (Participant, error)
}
