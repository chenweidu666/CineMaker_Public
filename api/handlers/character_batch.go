package handlers

import (
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func (h *CharacterLibraryHandler) BatchGenerateCharacterImages(c *gin.Context) {

	var req struct {
		CharacterIDs []string `json:"character_ids" binding:"required,min=1"`
		Model        string   `json:"model"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效")
		return
	}

	if len(req.CharacterIDs) > 10 {
		response.BadRequest(c, "单次最多生成10个角色")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	for _, cid := range req.CharacterIDs {
		if err := auth.VerifyCharacterTeam(h.db, cid, teamID); err != nil {
			response.Forbidden(c, "无权访问部分角色")
			return
		}
	}

	go h.libraryService.BatchGenerateCharacterImages(req.CharacterIDs, h.imageService, req.Model)

	response.Success(c, gin.H{
		"message": "批量生成任务已提交",
		"count":   len(req.CharacterIDs),
	})
}
