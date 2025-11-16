package user

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	UserId   string `gorm:"primaryKey"`
	UserName string
	TeamName string
	IsActive bool
}

func NewUser(id string, name string, teamName string, isActive bool) *User {
	return &User{
		UserId:   id,
		UserName: name,
		TeamName: teamName,
		IsActive: isActive,
	}
}
