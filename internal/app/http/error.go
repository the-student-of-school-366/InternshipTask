package http

import (
	"InternshipTask/internal/app/dto"

	"github.com/gin-gonic/gin"
)

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, dto.ErrorDTO{
		Error: dto.ErrorContent{
			Code:    code,
			Message: message,
		},
	})
}
