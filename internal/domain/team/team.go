package team

import (
	"InternshipTask/internal/domain/user"
	"errors"
)

var (
	ErrTeamNotFound = errors.New("team not found")
)

type Team struct {
	TeamName string              `json:"teamName"  gorm:"primaryKey"`
	Members  map[uint]*user.User `json:"members"`
}

func NewTeam(teamName string) *Team {
	return &Team{
		TeamName: teamName,
		Members:  make(map[uint]*user.User),
	}
}
