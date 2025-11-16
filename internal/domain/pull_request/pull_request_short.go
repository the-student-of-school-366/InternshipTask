package pull_request

type PullRequestShort struct {
	PullRequestId   string
	PullRequestName string
	AuthorId        string
	Status          PullRequestStatus
}

func NewPullRequestShort(id string, name string, authorId string, status PullRequestStatus) *PullRequestShort {
	return &PullRequestShort{
		PullRequestId:   id,
		PullRequestName: name,
		AuthorId:        authorId,
		Status:          status,
	}
}
