package handlers

import (
	"fmt"
	"strconv"

	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StoryboardHandler struct {
	storyboardService *services.StoryboardService
	taskService       *services.TaskService
	db                *gorm.DB
	log               *logger.Logger
}

func NewStoryboardHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *StoryboardHandler {
	return &StoryboardHandler{
		storyboardService: services.NewStoryboardService(db, cfg, log),
		taskService:       services.NewTaskService(db, log),
		db:                db,
		log:               log,
	}
}

// GenerateStoryboard 生成分镜头（异步）
func (h *StoryboardHandler) GenerateStoryboard(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	episodeID := c.Param("episode_id")

	if _, err := auth.VerifyEpisodeTeam(h.db, episodeID, teamID); err != nil {
		response.Forbidden(c, "无权操作该章节")
		return
	}

	// 接收可选的 model 和 shot_count 参数
	var req struct {
		Model     string `json:"model"`
		ShotCount int    `json:"shot_count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有提供body或者解析失败，使用默认值
		req.Model = ""
		req.ShotCount = 10
	}

	// 调用生成服务，该服务已经是异步的，会返回任务ID
	taskID, err := h.storyboardService.GenerateStoryboard(episodeID, req.Model, req.ShotCount)
	if err != nil {
		h.log.Errorw("Failed to generate storyboard", "error", err, "episode_id", episodeID)
		response.InternalError(c, err.Error())
		return
	}

	// 立即返回任务ID
	response.Success(c, gin.H{
		"task_id": taskID,
		"status":  "pending",
		"message": "分镜头生成任务已创建，正在后台处理...",
	})
}

// UpdateStoryboard 更新分镜
func (h *StoryboardHandler) UpdateStoryboard(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	storyboardID := c.Param("id")

	if err := auth.VerifyStoryboardTeam(h.db, storyboardID, teamID); err != nil {
		response.Forbidden(c, "无权操作该分镜")
		return
	}

	var req map[string]interface{}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	err := h.storyboardService.UpdateStoryboard(storyboardID, req)
	if err != nil {
		h.log.Errorw("Failed to update storyboard", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Storyboard updated successfully"})
}

// CreateStoryboard 创建分镜
func (h *StoryboardHandler) CreateStoryboard(c *gin.Context) {
	teamID := auth.GetTeamID(c)

	var req services.CreateStoryboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if _, err := auth.VerifyEpisodeTeam(h.db, fmt.Sprintf("%d", req.EpisodeID), teamID); err != nil {
		response.Forbidden(c, "无权操作该章节")
		return
	}

	sb, err := h.storyboardService.CreateStoryboard(&req)
	if err != nil {
		h.log.Errorw("Failed to create storyboard", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, sb)
}

// DeleteStoryboard 删除分镜
func (h *StoryboardHandler) DeleteStoryboard(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	storyboardIDStr := c.Param("id")

	if err := auth.VerifyStoryboardTeam(h.db, storyboardIDStr, teamID); err != nil {
		response.Forbidden(c, "无权操作该分镜")
		return
	}

	storyboardID, err := strconv.ParseUint(storyboardIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := h.storyboardService.DeleteStoryboard(uint(storyboardID)); err != nil {
		h.log.Errorw("Failed to delete storyboard", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GenerateVideoPrompt 生成视频提示词（使用AI生成详细格式）
func (h *StoryboardHandler) GenerateVideoPrompt(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	storyboardIDStr := c.Param("id")

	if err := auth.VerifyStoryboardTeam(h.db, storyboardIDStr, teamID); err != nil {
		response.Forbidden(c, "无权操作该分镜")
		return
	}

	storyboardID, err := strconv.ParseUint(storyboardIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid storyboard ID")
		return
	}

	var req struct {
		Model           string `json:"model"`
		Duration        int    `json:"duration"`
		EnableSubtitle  *bool  `json:"enable_subtitle"`
		GenerateAudio   *bool  `json:"generate_audio"`
		IncludeDialogue *bool  `json:"include_dialogue"`
		AspectRatio     string `json:"aspect_ratio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Model = ""
		req.Duration = 5
		defaultEnableSubtitle := true
		req.EnableSubtitle = &defaultEnableSubtitle
		defaultGenerateAudio := true
		req.GenerateAudio = &defaultGenerateAudio
		req.AspectRatio = "16:9"
	}

	// 记录接收到的参数
	enableSubValue := false
	if req.EnableSubtitle != nil {
		enableSubValue = *req.EnableSubtitle
	}
	enableAudioValue := false
	if req.GenerateAudio != nil {
		enableAudioValue = *req.GenerateAudio
	}
	includeDialogueValue := false
	if req.IncludeDialogue != nil {
		includeDialogueValue = *req.IncludeDialogue
	}
	h.log.Infow("GenerateVideoPrompt received parameters", "model", req.Model, "duration", req.Duration, "enable_subtitle", enableSubValue, "generate_audio", enableAudioValue, "include_dialogue", includeDialogueValue, "aspect_ratio", req.AspectRatio)

	// 调用服务生成视频提示词
	videoPrompt, err := h.storyboardService.GenerateVideoPromptWithAI(uint(storyboardID), req.Model, req.Duration, req.EnableSubtitle, req.GenerateAudio, req.AspectRatio, req.IncludeDialogue)
	if err != nil {
		h.log.Errorw("Failed to generate video prompt", "error", err, "storyboard_id", storyboardID)
		response.InternalError(c, err.Error())
		return
	}

	// 保存视频提示词到数据库
	if err := h.storyboardService.UpdateStoryboard(storyboardIDStr, map[string]interface{}{
		"video_prompt": videoPrompt,
	}); err != nil {
		h.log.Warnw("Failed to save video prompt to database", "error", err, "storyboard_id", storyboardID)
	}

	response.Success(c, gin.H{
		"video_prompt": videoPrompt,
	})
}
