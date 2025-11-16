package app

import (
	httpapp "InternshipTask/internal/app/http"
	logger "InternshipTask/internal/app/middleware"
	"InternshipTask/internal/domain/pull_request"
	"InternshipTask/internal/domain/team"
	"InternshipTask/internal/domain/user"
	prpg "InternshipTask/internal/infrastructure/postgres/pull_request"
	teampg "InternshipTask/internal/infrastructure/postgres/team"
	userpg "InternshipTask/internal/infrastructure/postgres/user"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type Config struct {
	PostgresDSN    string
	ConnectTimeout time.Duration
}

func New(cfg Config) (*gin.Engine, error) {
	if cfg.PostgresDSN == "" {
		return nil, fmt.Errorf("postgres DSN is empty")
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}

	teamStorage, err := teampg.NewPostgresStorage(&teampg.Config{
		DatabaseDSN:    cfg.PostgresDSN,
		ConnectTimeout: cfg.ConnectTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("init team storage: %w", err)
	}

	userStorage, err := userpg.NewPostgresStorage(&userpg.Config{
		DatabaseDSN:    cfg.PostgresDSN,
		ConnectTimeout: cfg.ConnectTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("init user storage: %w", err)
	}

	prStorage, err := prpg.NewPostgresStorage(&prpg.Config{
		DatabaseDSN:    cfg.PostgresDSN,
		ConnectTimeout: cfg.ConnectTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("init pull_request storage: %w", err)
	}

	teamService := team.NewService(teamStorage)
	userService := user.NewService(userStorage)
	prService := pull_request.NewService(prStorage, userService, teamService)

	router := gin.Default()
	router.Use(logger.LoggerMiddleware())
	httpapp.RegisterRoutes(router, teamService, userService, prService)

	return router, nil
}
