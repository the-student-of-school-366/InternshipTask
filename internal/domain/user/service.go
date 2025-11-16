package user

import (
	"context"
)

type Storager interface {
	GetByID(ctx context.Context, id string) (*User, error)
	SetIsActive(ctx context.Context, id string, isActive bool) (*User, error)
}

type Service struct {
	storage Storager
}

func NewService(storage Storager) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	return s.storage.GetByID(ctx, id)
}

func (s *Service) SetIsActive(ctx context.Context, id string, isActive bool) (*User, error) {
	return s.storage.SetIsActive(ctx, id, isActive)
}
