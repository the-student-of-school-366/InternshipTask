package team

import (
	"InternshipTask/internal/domain/team"
	"InternshipTask/internal/domain/user"
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
	ErrTeamNotFound   = errors.New("team not found")
	ErrQueryExecution = errors.New("query execution failed")
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

func (s *postgresStorage) Create(ctx context.Context, t team.Team) error {
	query := `INSERT INTO teams (team_name) VALUES ($1)
	          ON CONFLICT (team_name) DO NOTHING`
	if _, err := s.db.Exec(ctx, query, t.TeamName); err != nil {
		return fmt.Errorf("insert team: %w", ErrQueryExecution)
	}

	for _, member := range t.Members {
		upsertUserQuery := `
			INSERT INTO users (user_id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id)
			DO UPDATE SET
				username  = EXCLUDED.username,
				team_name = EXCLUDED.team_name,
				is_active = EXCLUDED.is_active
		`
		if _, err := s.db.Exec(ctx, upsertUserQuery,
			member.UserId,
			member.UserName,
			t.TeamName,
			member.IsActive,
		); err != nil {
			return fmt.Errorf("upsert user %s: %w", member.UserId, ErrQueryExecution)
		}
	}

	return nil
}

func (s *postgresStorage) GetByTeamName(ctx context.Context, teamName string) (team.Team, error) {
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)"
	err := s.db.QueryRow(ctx, checkQuery, teamName).Scan(&exists)
	if err != nil {
		return team.Team{}, fmt.Errorf("check team exists: %w", err)
	}
	if !exists {
		return team.Team{}, ErrTeamNotFound
	}
	query := `
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE team_name = $1
		ORDER BY user_id
	`
	rows, err := s.db.Query(ctx, query, teamName)
	if err != nil {
		return team.Team{}, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	members := make(map[uint]*user.User)
	index := uint(0)

	for rows.Next() {
		var u user.User
		err := rows.Scan(&u.UserId, &u.UserName, &u.TeamName, &u.IsActive)
		if err != nil {
			return team.Team{}, fmt.Errorf("scan user: %w", err)
		}
		members[index] = &u
		index++
	}

	if err = rows.Err(); err != nil {
		return team.Team{}, fmt.Errorf("rows iteration: %w", err)
	}

	return team.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}

func (s *postgresStorage) Close() error {
	err := s.db.Close(context.Background())

	if err != nil {
		return fmt.Errorf("postgres close: %w", err)
	}

	return nil
}
