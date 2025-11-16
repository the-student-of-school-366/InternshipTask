package team

import (
	"context"
	"testing"
)

type stubStorage struct {
	createCalled bool
	lastTeam     Team
}

func (s *stubStorage) Create(_ context.Context, t Team) error {
	s.createCalled = true
	s.lastTeam = t
	return nil
}

func (s *stubStorage) GetByTeamName(_ context.Context, _ string) (Team, error) {
	return Team{}, nil
}

func TestService_CreateDelegatesToStorage(t *testing.T) {
	storage := &stubStorage{}
	svc := NewService(storage)

	team := Team{TeamName: "backend"}
	if err := svc.Create(context.Background(), team); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if !storage.createCalled {
		t.Fatalf("expected storage.Create to be called")
	}
	if storage.lastTeam.TeamName != "backend" {
		t.Fatalf("expected team name backend, got %s", storage.lastTeam.TeamName)
	}
}


