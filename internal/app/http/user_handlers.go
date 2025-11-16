package http

import (
	"InternshipTask/internal/app/dto"
	"InternshipTask/internal/domain/user"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) setUserIsActive(c *gin.Context) {
	var req dto.SetUserActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	u, err := h.userService.SetIsActive(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": toUserDTO(u),
	})
}

func (h *Handler) getUserReviews(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "user_id is required")
		return
	}

	prs, err := h.prService.GetByReviewerID(context.Background(), userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	resp := dto.UserReviewsResponse{
		UserID:       userID,
		PullRequests: make([]dto.PullRequestShortDTO, 0, len(prs)),
	}

	for _, p := range prs {
		resp.PullRequests = append(resp.PullRequests, dto.PullRequestShortDTO{
			PullRequestID:   p.PullRequestId,
			PullRequestName: p.PullRequestName,
			AuthorID:        p.AuthorId,
			Status:          p.Status.String(),
		})
	}

	c.JSON(http.StatusOK, resp)
}

func toUserDTO(u *user.User) dto.UserDTO {
	return dto.UserDTO{
		UserID:   u.UserId,
		Username: u.UserName,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}
