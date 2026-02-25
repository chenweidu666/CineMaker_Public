package handlers

import (
	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetStoryboardFramePrompts 查询镜头的所有帧提示词
// GET /api/v1/storyboards/:id/frame-prompts
func GetStoryboardFramePrompts(db *gorm.DB, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamID := auth.GetTeamID(c)
		storyboardID := c.Param("id")

		if err := auth.VerifyStoryboardTeam(db, storyboardID, teamID); err != nil {
			response.Forbidden(c, "无权访问该分镜")
			return
		}

		var framePrompts []models.FramePrompt
		if err := db.Where("storyboard_id = ?", storyboardID).
			Order("created_at DESC").
			Find(&framePrompts).Error; err != nil {
			log.Errorw("Failed to query frame prompts", "error", err)
			response.InternalError(c, err.Error())
			return
		}

		response.Success(c, gin.H{
			"frame_prompts": framePrompts,
		})
	}
}
