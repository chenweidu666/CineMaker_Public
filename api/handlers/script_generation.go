package handlers

import (
	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ScriptGenerationHandler struct {
	scriptService *services.ScriptGenerationService
	taskService   *services.TaskService
	log           *logger.Logger
}

func NewScriptGenerationHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *ScriptGenerationHandler {
	return &ScriptGenerationHandler{
		scriptService: services.NewScriptGenerationService(db, cfg, log),
		taskService:   services.NewTaskService(db, log),
		log:           log,
	}
}

// RewriteScript 使用 AI 改写剧本（添加对话、动作描写等）
func (h *ScriptGenerationHandler) RewriteScript(c *gin.Context) {
	var req services.RewriteScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	taskID, err := h.scriptService.RewriteScript(&req)
	if err != nil {
		h.log.Errorw("Failed to rewrite script", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"task_id": taskID,
		"status":  "pending",
		"message": "剧本改写任务已创建，正在后台处理...",
	})
}

// RevertScriptRewrite 回滚剧本改写，恢复原始内容
func (h *ScriptGenerationHandler) RevertScriptRewrite(c *gin.Context) {
	var req struct {
		EpisodeID uint `json:"episode_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.scriptService.RevertScriptRewrite(req.EpisodeID); err != nil {
		h.log.Errorw("Failed to revert script rewrite", "error", err, "episode_id", req.EpisodeID)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "剧本已回滚到改写前的版本",
	})
}

func (h *ScriptGenerationHandler) GenerateCharacters(c *gin.Context) {
	var req services.GenerateCharactersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 直接调用服务层的异步方法，该方法会创建任务并返回任务ID
	taskID, err := h.scriptService.GenerateCharacters(&req)
	if err != nil {
		h.log.Errorw("Failed to generate characters", "error", err, "drama_id", req.DramaID)
		response.InternalError(c, err.Error())
		return
	}

	// 立即返回任务ID
	response.Success(c, gin.H{
		"task_id": taskID,
		"status":  "pending",
		"message": "角色生成任务已创建，正在后台处理...",
	})
}

// ParseExtract V3：从结构化剧本中程序解析提取角色和场景（同步，不走AI）
func (h *ScriptGenerationHandler) ParseExtract(c *gin.Context) {
	var req struct {
		EpisodeID uint `json:"episode_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.scriptService.ParseExtractFromScript(req.EpisodeID)
	if err != nil {
		h.log.Errorw("Failed to parse extract from script", "error", err, "episode_id", req.EpisodeID)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"characters": result.Characters,
		"scenes":     result.Scenes,
		"message":    "从剧本中解析提取完成",
	})
}
