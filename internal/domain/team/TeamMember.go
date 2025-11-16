package team

type TeamMember struct {
	UserId   string `json:"user_id" gorm:"primary_key"`
	UserName string `json:"user_name"`
	IsActive bool   `json:"is_active"`
}

func NewTeamMember(userId string, userName string, isActive bool) *TeamMember {
	return &TeamMember{
		UserId:   userId,
		UserName: userName,
		IsActive: isActive,
	}
}
