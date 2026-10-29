package participant

import (
	"context"
	"errors"
	"testing"

	"github.com/shawfire/participant-services/internal/identity"
)

func TestSelfAccessPolicy(t *testing.T) {
	for _, tc := range []struct {
		name, requested string
		principal       identity.Principal
		present         bool
	}{
		{name: "missing identity", requested: "P001"},
		{name: "empty subject", requested: "P001", principal: identity.Principal{ParticipantID: "P001"}, present: true},
		{name: "unmapped subject", requested: "P001", principal: identity.Principal{Subject: "user"}, present: true},
		{name: "cross participant", requested: "P002", principal: identity.Principal{Subject: "user", ParticipantID: "P001"}, present: true},
		{name: "unknown other record", requested: "P999", principal: identity.Principal{Subject: "user", ParticipantID: "P001"}, present: true},
		{name: "empty record", principal: identity.Principal{Subject: "user", ParticipantID: "P001"}, present: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.present {
				ctx = identity.WithPrincipal(ctx, tc.principal)
			}
			repo := &stubRepository{}
			for _, service := range []*Service{NewService(), NewServiceWithRepository(repo)} {
				got, err := service.Get(ctx, tc.requested)
				if !errors.Is(err, ErrForbidden) || got != (Participant{}) {
					t.Fatalf("expected empty forbidden result, got %+v, %v", got, err)
				}
			}
			if repo.calledID != "" {
				t.Fatal("denied request called repository")
			}
		})
	}
}

func TestRepositoryCannotReturnAnotherParticipant(t *testing.T) {
	repo := &stubRepository{participant: Participant{ID: "P002", FirstName: "Other"}}
	ctx := identity.WithPrincipal(context.Background(), identity.Principal{Subject: "user", ParticipantID: "P001"})
	got, err := NewServiceWithRepository(repo).Get(ctx, "P001")
	if !errors.Is(err, ErrForbidden) || got != (Participant{}) {
		t.Fatal("mismatched upstream participant exposed")
	}
}
