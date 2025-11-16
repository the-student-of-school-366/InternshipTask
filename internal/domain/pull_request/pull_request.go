package pull_request

import "time"

type PullRequestStatus string

const (
	OPEN   PullRequestStatus = "OPEN"
	MERGED PullRequestStatus = "MERGED"
)

func (s PullRequestStatus) String() string {
	return string(s)
}

type PR struct {
	PullRequestId     string `gorm:"primary_key"`
	PullRequestName   string
	AuthorId          string
	Status            PullRequestStatus
	AssignedReviewers []string
	CreatedAt         *time.Time
	MergedAt          *time.Time
}

func NewPR(id string, name string, authorId string, status PullRequestStatus) *PR {

	now := time.Now().UTC()

	return &PR{
		PullRequestId:     id,
		PullRequestName:   name,
		AuthorId:          authorId,
		Status:            status,
		AssignedReviewers: make([]string, 0),
		CreatedAt:         &now,
	}
}
