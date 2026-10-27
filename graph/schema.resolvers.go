package graph

import (
	"context"
	"strconv"

	"github.com/shawfire/participant-services/graph/model"
)

// CreateTodo handles the GraphQL createTodo mutation.
func (r *mutationResolver) CreateTodo(
	ctx context.Context,
	input model.NewTodo,
) (*model.Todo, error) {
	todo := &model.Todo{
		ID:   strconv.Itoa(len(r.todos) + 1),
		Text: input.Text,
		Done: false,
		User: &model.User{
			ID:   input.UserID,
			Name: "Test Participant",
		},
	}

	r.todos = append(r.todos, todo)

	return todo, nil
}

// Todos handles the GraphQL todos query.
func (r *queryResolver) Todos(
	ctx context.Context,
) ([]*model.Todo, error) {
	return r.todos, nil
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type (
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
)
