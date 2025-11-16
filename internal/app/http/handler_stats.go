package http

import (
	"InternshipTask/internal/app/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) stats(c *gin.Context) {
	stats, err := h.prService.ReviewerStats(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.StatsDTO{
		ReviewAssignments: stats,
	})
}


