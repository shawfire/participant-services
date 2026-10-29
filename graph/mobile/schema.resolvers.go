package mobile

import (
	"context"
	"errors"

	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/shawfire/participant-services/graph/mobile/generated"
	"github.com/shawfire/participant-services/graph/mobile/model"
	"github.com/shawfire/participant-services/internal/participant"
)

func (r *queryResolver) Profile(
	ctx context.Context,
	participantID string,
) (*model.ParticipantProfile, error) {
	p, err := r.Participants.Get(ctx, participantID)

	if errors.Is(err, participant.ErrForbidden) {
		return nil, &gqlerror.Error{
			Message:    "forbidden",
			Extensions: map[string]any{"code": "FORBIDDEN"},
		}
	}

	if errors.Is(err, participant.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &model.ParticipantProfile{
		ParticipantID: p.ID,
		DisplayName:   p.FirstName + " " + p.LastName,
	}, nil
}

func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct {
	*Resolver
}
