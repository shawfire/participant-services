package mobile

import "github.com/shawfire/participant-services/internal/participant"

// Resolver provides dependencies to Mobile GraphQL resolvers.
type Resolver struct {
	Participants *participant.Service
}
