package http

import (
	"InternshipTask/internal/app/dto"
	"InternshipTask/internal/domain/pull_request"
	"InternshipTask/internal/domain/team"
	"InternshipTask/internal/domain/user"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubTeamStorage struct {
	createdTeam *team.Team
	teamByName  map[string]team.Team
}

func (s *stubTeamStorage) Create(_ context.Context, t team.Team) error {
	s.createdTeam = &t
	if s.teamByName == nil {
		s.teamByName = make(map[string]team.Team)
	}
	s.teamByName[t.TeamName] = t
	return nil
}

func (s *stubTeamStorage) GetByTeamName(_ context.Context, name string) (team.Team, error) {
	if s.teamByName == nil {
		return team.Team{}, team.ErrTeamNotFound
	}
	t, ok := s.teamByName[name]
	if !ok {
		return team.Team{}, team.ErrTeamNotFound
	}
	return t, nil
}

type stubUserStorage struct {
	users map[string]*user.User
}

func (s *stubUserStorage) GetByID(_ context.Context, id string) (*user.User, error) {
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return nil, user.ErrUserNotFound
}

func (s *stubUserStorage) SetIsActive(_ context.Context, id string, isActive bool) (*user.User, error) {
	if s.users == nil {
		s.users = make(map[string]*user.User)
	}
	u, ok := s.users[id]
	if !ok {
		u = &user.User{UserId: id}
		s.users[id] = u
	}
	u.IsActive = isActive
	return u, nil
}

type stubPRRepo struct {
	prByID        map[string]*pull_request.PR
	prsByReviewer []pull_request.PullRequestShort
	stats         map[string]int64
}

func (r *stubPRRepo) Create(_ context.Context, pr *pull_request.PR) error {
	if r.prByID == nil {
		r.prByID = make(map[string]*pull_request.PR)
	}
	r.prByID[pr.PullRequestId] = pr
	return nil
}

func (r *stubPRRepo) GetByID(_ context.Context, id string) (*pull_request.PR, error) {
	if pr, ok := r.prByID[id]; ok {
		return pr, nil
	}
	return nil, pull_request.ErrNotFound
}

func (r *stubPRRepo) Update(_ context.Context, pr *pull_request.PR) error {
	if r.prByID == nil {
		r.prByID = make(map[string]*pull_request.PR)
	}
	r.prByID[pr.PullRequestId] = pr
	return nil
}

func (r *stubPRRepo) GetByReviewerID(_ context.Context, _ string) ([]pull_request.PullRequestShort, error) {
	return r.prsByReviewer, nil
}

func (r *stubPRRepo) GetReviewerStats(_ context.Context) (map[string]int64, error) {
	return r.stats, nil
}

func buildRouter() (*gin.Engine, *stubTeamStorage, *stubUserStorage, *stubPRRepo) {
	gin.SetMode(gin.TestMode)

	teamStorage := &stubTeamStorage{}
	userStorage := &stubUserStorage{
		users: map[string]*user.User{
			"author": {UserId: "author", UserName: "Author", TeamName: "backend", IsActive: true},
			"u2":     {UserId: "u2", UserName: "Alice", TeamName: "backend", IsActive: true},
		},
	}
	prRepo := &stubPRRepo{
		stats: map[string]int64{"u2": 2},
		prsByReviewer: []pull_request.PullRequestShort{
			*pull_request.NewPullRequestShort("pr-1", "PR 1", "author", pull_request.OPEN),
		},
	}

	teamSvc := team.NewService(teamStorage)
	userSvc := user.NewService(userStorage)
	prSvc := pull_request.NewService(prRepo, userSvc, teamSvc)

	r := gin.Default()
	RegisterRoutes(r, teamSvc, userSvc, prSvc)
	return r, teamStorage, userStorage, prRepo
}

func TestHealthHandler_OK(t *testing.T) {
	r, _, _, _ := buildRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestStatsHandler_OK(t *testing.T) {
	r, _, _, _ := buildRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/stats", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestCreateTeamHandler_Success(t *testing.T) {
	r, teamStorage, _, _ := buildRouter()

	body := dto.TeamDTO{
		TeamName: "backend",
		Members: []dto.TeamMemberDTO{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}
	data, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body=%s", w.Code, w.Body.String())
	}
	if teamStorage.createdTeam == nil || teamStorage.createdTeam.TeamName != "backend" {
		t.Fatalf("expected team to be created in storage")
	}
}

func TestGetTeamHandler_Success(t *testing.T) {
	r, teamStorage, _, _ := buildRouter()

	teamStorage.teamByName = map[string]team.Team{
		"backend": *team.NewTeam("backend"),
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestSetUserIsActiveHandler_Success(t *testing.T) {
	r, _, userStorage, _ := buildRouter()

	body := dto.SetUserActiveRequest{
		UserID:   "u2",
		IsActive: false,
	}
	data, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
	if userStorage.users["u2"] == nil || userStorage.users["u2"].IsActive != false {
		t.Fatalf("expected user u2 to be set inactive")
	}
}

func TestGetUserReviewsHandler_Success(t *testing.T) {
	r, _, _, _ := buildRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestCreatePullRequestHandler_Success(t *testing.T) {
	r, teamStorage, userStorage, prRepo := buildRouter()

	teamStorage.teamByName = map[string]team.Team{
		"backend": {
			TeamName: "backend",
			Members: map[uint]*user.User{
				0: userStorage.users["author"],
				1: userStorage.users["u2"],
			},
		},
	}

	body := dto.CreatePullRequestRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        "author",
	}
	data, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body=%s", w.Code, w.Body.String())
	}
	if prRepo.prByID["pr-1"] == nil {
		t.Fatalf("expected pr-1 to be created in repo")
	}
}

func TestMergePullRequestHandler_Success(t *testing.T) {
	r, _, _, prRepo := buildRouter()

	prRepo.prByID = map[string]*pull_request.PR{
		"pr-1": pull_request.NewPR("pr-1", "Test", "author", pull_request.OPEN),
	}

	body := dto.MergePullRequestRequest{PullRequestID: "pr-1"}
	data, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestReassignPullRequestHandler_Success(t *testing.T) {
	r, teamStorage, userStorage, prRepo := buildRouter()

	teamStorage.teamByName = map[string]team.Team{
		"backend": {
			TeamName: "backend",
			Members: map[uint]*user.User{
				0: userStorage.users["author"],
				1: userStorage.users["u2"],
			},
		},
	}

	pr := pull_request.NewPR("pr-1", "Test", "author", pull_request.OPEN)
	pr.AssignedReviewers = []string{"u2"}
	prRepo.prByID = map[string]*pull_request.PR{"pr-1": pr}

	body := dto.ReassignPullRequestRequest{
		PullRequestID: "pr-1",
		OldUserID:     "u2",
	}
	data, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
}
