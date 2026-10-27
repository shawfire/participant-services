package graph

import (
	"context"
	"testing"

	"github.com/shawfire/participant-services/graph/model"
)

func TestCreateTodo(t *testing.T) {
	r := &Resolver{}
	mutation := r.Mutation()

	todo, err := mutation.CreateTodo(context.Background(), model.NewTodo{
		Text:   "Prepare NDIA interview",
		UserID: "shawfire",
	})
	if err != nil {
		t.Fatalf("CreateTodo failed: %v", err)
	}

	if todo.ID != "1" {
		t.Errorf("expected ID 1, got %s", todo.ID)
	}
	if todo.Text != "Prepare NDIA interview" {
		t.Errorf("unexpected text: %s", todo.Text)
	}
	if todo.Done {
		t.Error("new todo should not be completed")
	}

	todos, err := r.Query().Todos(context.Background())
	if err != nil {
		t.Fatalf("Todos failed: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}
}
