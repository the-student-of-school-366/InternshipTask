package team

import "context"

type Storager interface {
	Create(ctx context.Context, team Team) error
	GetByTeamName(ctx context.Context, teamName string) (Team, error)
}
type Service struct {
	storage Storager
	Teams   []Team
}

func NewService(storage Storager) *Service {
	return &Service{
		storage: storage,
		Teams:   make([]Team, 0, 2),
	}
}

func (s *Service) Create(ctx context.Context, team Team) error {
	return s.storage.Create(ctx, team)
}

func (s *Service) GetByTeamName(ctx context.Context, teamName string) (Team, error) {
	return s.storage.GetByTeamName(ctx, teamName)
}
