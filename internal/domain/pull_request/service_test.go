package pull_request

import (
	"InternshipTask/internal/domain/team"
	"InternshipTask/internal/domain/user"
	"context"
	"testing"
)

type stubPRRepo struct {
	prsByID        map[string]*PR
	created        *PR
	updated        *PR
	reviewerStats  map[string]int64
	getByReviewerR []PullRequestShort
}

func (r *stubPRRepo) Create(_ context.Context, pr *PR) error {
	r.created = pr
	if r.prsByID == nil {
		r.prsByID = make(map[string]*PR)
	}
	r.prsByID[pr.PullRequestId] = pr
	return nil
}

func (r *stubPRRepo) GetByID(_ context.Context, id string) (*PR, error) {
	if pr, ok := r.prsByID[id]; ok {
		return pr, nil
	}
	return nil, ErrNotFound
}

func (r *stubPRRepo) Update(_ context.Context, pr *PR) error {
	r.updated = pr
	if r.prsByID == nil {
		r.prsByID = make(map[string]*PR)
	}
	r.prsByID[pr.PullRequestId] = pr
	return nil
}

func (r *stubPRRepo) GetByReviewerID(_ context.Context, _ string) ([]PullRequestShort, error) {
	return r.getByReviewerR, nil
}

func (r *stubPRRepo) GetReviewerStats(_ context.Context) (map[string]int64, error) {
	return r.reviewerStats, nil
}

type stubUserReader struct {
	users map[string]*user.User
}

func (s *stubUserReader) GetByID(_ context.Context, id string) (*user.User, error) {
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return nil, user.ErrUserNotFound
}

type stubTeamReader struct {
	teams map[string]team.Team
}

func (s *stubTeamReader) GetByTeamName(_ context.Context, name string) (team.Team, error) {
	if t, ok := s.teams[name]; ok {
		return t, nil
	}
	return team.Team{}, team.ErrTeamNotFound
}

func TestService_CreateAssignsUpToTwoReviewers(t *testing.T) {
	repo := &stubPRRepo{}
	userR := &stubUserReader{
		users: map[string]*user.User{
			"u1": {UserId: "u1", UserName: "Author", TeamName: "backend", IsActive: true},
			"u2": {UserId: "u2", UserName: "Alice", TeamName: "backend", IsActive: true},
			"u3": {UserId: "u3", UserName: "Bob", TeamName: "backend", IsActive: true},
		},
	}
	teamR := &stubTeamReader{
		teams: map[string]team.Team{
			"backend": {
				TeamName: "backend",
				Members: map[uint]*user.User{
					0: userR.users["u1"],
					1: userR.users["u2"],
					2: userR.users["u3"],
				},
			},
		},
	}

	svc := NewService(repo, userR, teamR)

	pr, err := svc.Create(context.Background(), "pr-1", "Test PR", "u1")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if pr.Status != OPEN {
		t.Fatalf("expected status OPEN, got %s", pr.Status)
	}
	if len(pr.AssignedReviewers) == 0 || len(pr.AssignedReviewers) > 2 {
		t.Fatalf("expected 1 or 2 reviewers, got %d", len(pr.AssignedReviewers))
	}
	for _, rID := range pr.AssignedReviewers {
		if rID == "u1" {
			t.Fatalf("author must not be assigned as reviewer")
		}
	}
}

func TestService_MergeIsIdempotent(t *testing.T) {
	repo := &stubPRRepo{
		prsByID: map[string]*PR{
			"pr-1": NewPR("pr-1", "Test", "u1", OPEN),
		},
	}
	userR := &stubUserReader{}
	teamR := &stubTeamReader{}

	svc := NewService(repo, userR, teamR)

	pr, err := svc.Merge(context.Background(), "pr-1")
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}
	if pr.Status != MERGED || pr.MergedAt == nil {
		t.Fatalf("expected PR to be MERGED with non-nil MergedAt")
	}
	firstMergedAt := *pr.MergedAt

	pr2, err := svc.Merge(context.Background(), "pr-1")
	if err != nil {
		t.Fatalf("second Merge() error = %v", err)
	}
	if pr2.Status != MERGED || pr2.MergedAt == nil {
		t.Fatalf("expected PR to stay MERGED")
	}
	if !firstMergedAt.Equal(*pr2.MergedAt) {
		t.Fatalf("expected mergedAt to stay the same between merges")
	}
}

func TestService_ReviewerStats(t *testing.T) {
	repo := &stubPRRepo{
		reviewerStats: map[string]int64{
			"u2": 3,
			"u3": 1,
		},
	}
	userR := &stubUserReader{}
	teamR := &stubTeamReader{}

	svc := NewService(repo, userR, teamR)

	stats, err := svc.ReviewerStats(context.Background())
	if err != nil {
		t.Fatalf("ReviewerStats() error = %v", err)
	}
	if stats["u2"] != 3 || stats["u3"] != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}
