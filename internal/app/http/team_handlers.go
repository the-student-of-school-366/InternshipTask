package http

import (
	"InternshipTask/internal/app/dto"
	"InternshipTask/internal/domain/team"
	"InternshipTask/internal/domain/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) createTeam(c *gin.Context) {
	var req dto.TeamDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	domainTeam := team.NewTeam(req.TeamName)
	for i, m := range req.Members {
		domainTeam.Members[uint(i)] = user.NewUser(m.UserID, m.Username, req.TeamName, m.IsActive)
	}

	if err := h.teamService.Create(c.Request.Context(), *domainTeam); err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"team": toTeamDTO(domainTeam),
	})
}

func (h *Handler) getTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	t, err := h.teamService.GetByTeamName(c.Request.Context(), teamName)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	c.JSON(http.StatusOK, toTeamDTOPtr(&t))
}

func toTeamDTO(t *team.Team) dto.TeamDTO {
	return *toTeamDTOPtr(t)
}

func toTeamDTOPtr(t *team.Team) *dto.TeamDTO {
	members := make([]dto.TeamMemberDTO, 0, len(t.Members))
	for _, u := range t.Members {
		if u == nil {
			continue
		}
		members = append(members, dto.TeamMemberDTO{
			UserID:   u.UserId,
			Username: u.UserName,
			IsActive: u.IsActive,
		})
	}

	return &dto.TeamDTO{
		TeamName: t.TeamName,
		Members:  members,
	}
}
