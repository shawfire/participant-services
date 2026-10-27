package portal

import "github.com/shawfire/participant-services/internal/participant"

// Resolver provides dependencies to Portal GraphQL resolvers.
type Resolver struct {
	Participants *participant.Service
}
