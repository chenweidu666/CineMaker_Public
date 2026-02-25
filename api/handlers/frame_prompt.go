package handlers

import (
	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// FramePromptHandler 处理帧提示词生成请求
type FramePromptHandler struct {
	framePromptService *services.FramePromptService
	log                *logger.Logger
}

// NewFramePromptHandler 创建帧提示词处理器
func NewFramePromptHandler(framePromptService *services.FramePromptService, log *logger.Logger) *FramePromptHandler {
	return &FramePromptHandler{
		framePromptService: framePromptService,
		log:                log,
	}
}

// BatchGenerateFramePrompts 批量为章节所有分镜生成帧提示词
// POST /api/v1/episodes/:episode_id/batch-frame-prompts
func (h *FramePromptHandler) BatchGenerateFramePrompts(c *gin.Context) {
	episodeID := c.Param("episode_id")

	var req struct {
		FrameType    string `json:"frame_type"`
		Model        string `json:"model"`
		SkipExisting bool   `json:"skip_existing"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.FrameType == "" {
		req.FrameType = "first" // 默认生成首帧提示词
	}

	taskID, err := h.framePromptService.BatchGenerateFramePrompts(episodeID, services.FrameType(req.FrameType), req.Model, req.SkipExisting)
	if err != nil {
		h.log.Errorw("Failed to batch generate frame prompts", "error", err, "episode_id", episodeID)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"task_id": taskID,
		"status":  "pending",
		"message": "批量帧提示词生成任务已创建，正在后台处理...",
	})
}

// GenerateFramePrompt 生成指定类型的帧提示词
// POST /api/v1/storyboards/:id/frame-prompt
func (h *FramePromptHandler) GenerateFramePrompt(c *gin.Context) {
	storyboardID := c.Param("id")

	var req struct {
		FrameType  string `json:"frame_type"`
		PanelCount int    `json:"panel_count"`
		Model      string `json:"model"`
		ImageRatio string `json:"image_ratio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	h.log.Infow("GenerateFramePrompt request parsed", "storyboard_id", storyboardID, "frame_type", req.FrameType, "image_ratio", req.ImageRatio)

	serviceReq := services.GenerateFramePromptRequest{
		StoryboardID: storyboardID,
		FrameType:    services.FrameType(req.FrameType),
		PanelCount:   req.PanelCount,
		ImageRatio:   req.ImageRatio,
	}

	// 直接调用服务层的异步方法，该方法会创建任务并返回任务ID
	taskID, err := h.framePromptService.GenerateFramePrompt(serviceReq, req.Model)
	if err != nil {
		h.log.Errorw("Failed to generate frame prompt", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	// 立即返回任务ID
	response.Success(c, gin.H{
		"task_id": taskID,
		"status":  "pending",
		"message": "帧提示词生成任务已创建，正在后台处理...",
	})
}
