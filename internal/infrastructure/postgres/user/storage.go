package user

import (
	domain "InternshipTask/internal/domain/user"
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
	ErrUserNotFound   = errors.New("user not found")
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

var _ domain.Storager = (*postgresStorage)(nil)

func (s *postgresStorage) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`

	row := s.db.QueryRow(ctx, query, id)

	var u domain.User

	err := row.Scan(&u.UserId, &u.UserName, &u.TeamName, &u.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user: %w", err)
	}

	return &u, nil
}

func (s *postgresStorage) SetIsActive(ctx context.Context, id string, isActive bool) (*domain.User, error) {
	query := `
		UPDATE users
		SET is_active = $2
		WHERE user_id = $1
		RETURNING user_id, username, team_name, is_active
	`

	row := s.db.QueryRow(ctx, query, id, isActive)

	var u domain.User

	err := row.Scan(&u.UserId, &u.UserName, &u.TeamName, &u.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update user is_active: %w", err)
	}

	return &u, nil
}

func (s *postgresStorage) Close() error {
	if err := s.db.Close(context.Background()); err != nil {
		return fmt.Errorf("postgres close: %w", err)
	}
	return nil
}
