package http

import (
	"InternshipTask/internal/domain/pull_request"
	"InternshipTask/internal/domain/team"
	"InternshipTask/internal/domain/user"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	teamService *team.Service
	userService *user.Service
	prService   *pull_request.Service
}

func RegisterRoutes(r *gin.Engine, teamSvc *team.Service, userSvc *user.Service, prSvc *pull_request.Service) {
	h := &Handler{
		teamService: teamSvc,
		userService: userSvc,
		prService:   prSvc,
	}

	r.GET("/health", h.health)

	r.GET("/stats", h.stats)

	r.POST("/team/add", h.createTeam)
	r.GET("/team/get", h.getTeam)

	r.POST("/users/setIsActive", h.setUserIsActive)
	r.GET("/users/getReview", h.getUserReviews)

	r.POST("/pullRequest/create", h.createPullRequest)
	r.POST("/pullRequest/merge", h.mergePullRequest)
	r.POST("/pullRequest/reassign", h.reassignPullRequest)
}
