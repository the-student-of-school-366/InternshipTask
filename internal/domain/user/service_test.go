package user

import (
	"context"
	"testing"
)

type stubUserStorage struct {
	setIsActiveCalled bool
	lastID            string
	lastActive        bool
	user              *User
}

func (s *stubUserStorage) GetByID(_ context.Context, id string) (*User, error) {
	return &User{UserId: id}, nil
}

func (s *stubUserStorage) SetIsActive(_ context.Context, id string, isActive bool) (*User, error) {
	s.setIsActiveCalled = true
	s.lastID = id
	s.lastActive = isActive
	if s.user != nil {
		return s.user, nil
	}
	return &User{UserId: id, IsActive: isActive}, nil
}

func TestService_SetIsActiveDelegatesToStorage(t *testing.T) {
	storage := &stubUserStorage{}
	svc := NewService(storage)

	user, err := svc.SetIsActive(context.Background(), "u1", true)
	if err != nil {
		t.Fatalf("SetIsActive() error = %v", err)
	}

	if !storage.setIsActiveCalled {
		t.Fatalf("expected storage.SetIsActive to be called")
	}
	if storage.lastID != "u1" || !storage.lastActive {
		t.Fatalf("unexpected args: id=%s active=%v", storage.lastID, storage.lastActive)
	}
	if user.UserId != "u1" || !user.IsActive {
		t.Fatalf("unexpected returned user: %+v", user)
	}
}


