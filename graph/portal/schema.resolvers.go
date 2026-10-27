package portal

import (
	"context"
	"errors"

	"github.com/shawfire/participant-services/graph/portal/generated"
	"github.com/shawfire/participant-services/graph/portal/model"
	"github.com/shawfire/participant-services/internal/participant"
)

func (r *queryResolver) Participant(
	ctx context.Context,
	id string,
) (*model.Participant, error) {
	p, err := r.Participants.Get(ctx, id)
	if errors.Is(err, participant.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &model.Participant{
		ID:        p.ID,
		FirstName: p.FirstName,
		LastName:  p.LastName,
	}, nil
}

func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct {
	*Resolver
}
