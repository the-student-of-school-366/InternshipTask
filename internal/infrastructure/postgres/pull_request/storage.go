package pull_request

import (
	domain "InternshipTask/internal/domain/pull_request"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"
)

type Config struct {
	DatabaseDSN    string
	ConnectTimeout time.Duration
}

type postgresStorage struct {
	db *pgx.Conn
}

var (
	ErrConnectTimeout = errors.New("connect timeout")
)

func NewPostgresStorage(cfg *Config) (*postgresStorage, error) {
	connCtx, cancel := context.WithTimeoutCause(context.Background(), cfg.ConnectTimeout, ErrConnectTimeout)
	defer cancel()
	conn, err := pgx.Connect(connCtx, cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}

	return &postgresStorage{
		db: conn,
	}, nil
}

var _ domain.Repository = (*postgresStorage)(nil)

func (s *postgresStorage) Create(ctx context.Context, pr *domain.PR) error {
	query := `
		INSERT INTO pull_requests (
			pull_request_id,
			pull_request_name,
			author_id,
			status,
			assigned_reviewers,
			created_at,
			merged_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.Exec(ctx, query,
		pr.PullRequestId,
		pr.PullRequestName,
		pr.AuthorId,
		pr.Status.String(),
		pr.AssignedReviewers,
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		return fmt.Errorf("insert pull_request: %w", err)
	}

	return nil
}

func (s *postgresStorage) GetByID(ctx context.Context, id string) (*domain.PR, error) {
	query := `
		SELECT 
			pull_request_id,
			pull_request_name,
			author_id,
			status,
			assigned_reviewers,
			created_at,
			merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	row := s.db.QueryRow(ctx, query, id)

	var pr domain.PR
	var status string

	err := row.Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&status,
		&pr.AssignedReviewers,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select pull_request: %w", err)
	}

	pr.Status = domain.PullRequestStatus(status)

	return &pr, nil
}

func (s *postgresStorage) Update(ctx context.Context, pr *domain.PR) error {
	query := `
		UPDATE pull_requests
		SET
			status = $2,
			assigned_reviewers = $3,
			merged_at = $4
		WHERE pull_request_id = $1
	`

	_, err := s.db.Exec(ctx, query,
		pr.PullRequestId,
		pr.Status.String(),
		pr.AssignedReviewers,
		pr.MergedAt,
	)
	if err != nil {
		return fmt.Errorf("update pull_request: %w", err)
	}

	return nil
}

func (s *postgresStorage) GetByReviewerID(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error) {
	query := `
		SELECT 
			pull_request_id,
			pull_request_name,
			author_id,
			status
		FROM pull_requests
		WHERE $1 = ANY(assigned_reviewers)
		ORDER BY created_at
	`

	rows, err := s.db.Query(ctx, query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("select pull_requests by reviewer: %w", err)
	}
	defer rows.Close()

	var result []domain.PullRequestShort

	for rows.Next() {
		var (
			id        string
			name      string
			authorID  string
			statusStr string
		)

		if err := rows.Scan(&id, &name, &authorID, &statusStr); err != nil {
			return nil, fmt.Errorf("scan pull_request_short: %w", err)
		}

		result = append(result, *domain.NewPullRequestShort(
			id,
			name,
			authorID,
			domain.PullRequestStatus(statusStr),
		))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return result, nil
}

func (s *postgresStorage) GetReviewerStats(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT reviewer_id, COUNT(*) AS assign_count
		FROM (
			SELECT unnest(assigned_reviewers) AS reviewer_id
			FROM pull_requests
		) AS t
		GROUP BY reviewer_id
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select reviewer stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int64)

	for rows.Next() {
		var reviewerID string
		var count int64
		if err := rows.Scan(&reviewerID, &count); err != nil {
			return nil, fmt.Errorf("scan reviewer stats: %w", err)
		}
		stats[reviewerID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return stats, nil
}
