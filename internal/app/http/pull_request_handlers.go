package http

import (
	"InternshipTask/internal/app/dto"
	"InternshipTask/internal/domain/pull_request"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) createPullRequest(c *gin.Context) {
	var req dto.CreatePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	pr, err := h.prService.Create(c.Request.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"pr": toPullRequestDTO(pr),
	})
}

func (h *Handler) mergePullRequest(c *gin.Context) {
	var req dto.MergePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	pr, err := h.prService.Merge(c.Request.Context(), req.PullRequestID)
	if err != nil {
		if errors.Is(err, pull_request.ErrNotFound) {
			writeError(c, http.StatusNotFound, "NOT_FOUND", "pr not found")
			return
		}
		writeError(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr": toPullRequestDTO(pr),
	})
}

func (h *Handler) reassignPullRequest(c *gin.Context) {
	var req dto.ReassignPullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	pr, replacedBy, err := h.prService.Reassign(c.Request.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":          toPullRequestDTO(pr),
		"replaced_by": replacedBy,
	})
}

func toPullRequestDTO(pr *pull_request.PR) dto.PullRequestDTO {
	return dto.PullRequestDTO{
		PullRequestID:     pr.PullRequestId,
		PullRequestName:   pr.PullRequestName,
		AuthorID:          pr.AuthorId,
		Status:            pr.Status.String(),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}
