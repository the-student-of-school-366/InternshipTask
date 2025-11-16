package pull_request

import (
	"InternshipTask/internal/domain/team"
	"InternshipTask/internal/domain/user"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

var (
	ErrPRExists    = errors.New("pr already exists")
	ErrPRMerged    = errors.New("pr already merged")
	ErrNotAssigned = errors.New("reviewer not assigned")
	ErrNoCandidate = errors.New("no active replacement user")
	ErrNotFound    = errors.New("pr not found")
)

type Repository interface {
	Create(ctx context.Context, pr *PR) error
	GetByID(ctx context.Context, id string) (*PR, error)
	Update(ctx context.Context, pr *PR) error
	GetByReviewerID(ctx context.Context, reviewerID string) ([]PullRequestShort, error)
	// Для статистики по назначенным ревьюверам.
	GetReviewerStats(ctx context.Context) (map[string]int64, error)
}

type UserReader interface {
	GetByID(ctx context.Context, id string) (*user.User, error)
}

type TeamReader interface {
	GetByTeamName(ctx context.Context, teamName string) (team.Team, error)
}

type Service struct {
	repo       Repository
	userReader UserReader
	teamReader TeamReader
	rand       *rand.Rand
}

func NewService(repo Repository, ur UserReader, tr TeamReader) *Service {
	return &Service{
		repo:       repo,
		userReader: ur,
		teamReader: tr,
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Service) Create(ctx context.Context, id, name, authorID string) (*PR, error) {
	if _, err := s.repo.GetByID(ctx, id); err == nil {
		return nil, ErrPRExists
	} else if !errors.Is(err, ErrNotFound) && err != nil {
		return nil, fmt.Errorf("get pr by id: %w", err)
	}

	author, err := s.userReader.GetByID(ctx, authorID)
	if err != nil {
		// Соответствует 404 Автор/команда не найдены.
		return nil, fmt.Errorf("get author: %w", err)
	}

	t, err := s.teamReader.GetByTeamName(ctx, author.TeamName)
	if err != nil {
		return nil, fmt.Errorf("get team: %w", err)
	}

	reviewers := s.pickReviewersFromTeam(&t, authorID)

	pr := NewPR(id, name, authorID, OPEN)
	pr.AssignedReviewers = reviewers

	if err := s.repo.Create(ctx, pr); err != nil {
		return nil, fmt.Errorf("create pr: %w", err)
	}

	return pr, nil
}

func (s *Service) Merge(ctx context.Context, id string) (*PR, error) {
	pr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if pr.Status == MERGED {
		// Идемпотентность — просто возвращаем текущее состояние.
		return pr, nil
	}

	now := time.Now().UTC()
	pr.Status = MERGED
	pr.MergedAt = &now

	if err := s.repo.Update(ctx, pr); err != nil {
		return nil, fmt.Errorf("update pr on merge: %w", err)
	}

	return pr, nil
}

func (s *Service) Reassign(ctx context.Context, prID, oldUserID string) (*PR, string, error) {
	pr, err := s.repo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	if pr.Status == MERGED {
		return nil, "", ErrPRMerged
	}

	idx := -1
	for i, rid := range pr.AssignedReviewers {
		if rid == oldUserID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, "", ErrNotAssigned
	}

	oldUser, err := s.userReader.GetByID(ctx, oldUserID)
	if err != nil {
		return nil, "", fmt.Errorf("get old reviewer: %w", err)
	}

	t, err := s.teamReader.GetByTeamName(ctx, oldUser.TeamName)
	if err != nil {
		return nil, "", fmt.Errorf("get team for old reviewer: %w", err)
	}

	candidate, ok := s.pickReplacementFromTeam(&t, oldUserID, pr.AssignedReviewers)
	if !ok {
		return nil, "", ErrNoCandidate
	}

	pr.AssignedReviewers[idx] = candidate

	if err := s.repo.Update(ctx, pr); err != nil {
		return nil, "", fmt.Errorf("update pr on reassign: %w", err)
	}

	return pr, candidate, nil
}

func (s *Service) GetByReviewerID(ctx context.Context, userID string) ([]PullRequestShort, error) {
	return s.repo.GetByReviewerID(ctx, userID)
}

// ReviewerStats возвращает статистику "user_id -> количество назначений ревьювером".
func (s *Service) ReviewerStats(ctx context.Context) (map[string]int64, error) {
	return s.repo.GetReviewerStats(ctx)
}

func (s *Service) pickReviewersFromTeam(t *team.Team, authorID string) []string {
	candidates := make([]string, 0, len(t.Members))

	for _, u := range t.Members {
		if u == nil {
			continue
		}
		if !u.IsActive {
			continue
		}
		if u.UserId == authorID {
			continue
		}
		candidates = append(candidates, u.UserId)
	}

	s.rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	if len(candidates) > 2 {
		candidates = candidates[:2]
	}

	return candidates
}

func (s *Service) pickReplacementFromTeam(t *team.Team, oldUserID string, currentReviewers []string) (string, bool) {
	currentSet := make(map[string]struct{}, len(currentReviewers))
	for _, id := range currentReviewers {
		currentSet[id] = struct{}{}
	}

	candidates := make([]string, 0, len(t.Members))
	for _, u := range t.Members {
		if u == nil {
			continue
		}
		if !u.IsActive {
			continue
		}
		if u.UserId == oldUserID {
			continue
		}
		if _, used := currentSet[u.UserId]; used {
			continue
		}
		candidates = append(candidates, u.UserId)
	}

	if len(candidates) == 0 {
		return "", false
	}

	idx := s.rand.Intn(len(candidates))
	return candidates[idx], true
}
