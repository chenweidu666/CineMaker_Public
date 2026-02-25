package handlers

import (
	"errors"
	"fmt"

	services2 "github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/domain"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SceneHandler struct {
	db           *gorm.DB
	sceneService *services2.StoryboardCompositionService
	log          *logger.Logger
	aiService    *services2.AIService
	config       *config.Config
}

func NewSceneHandler(db *gorm.DB, log *logger.Logger, imageGenService *services2.ImageGenerationService, aiService *services2.AIService, cfg *config.Config) *SceneHandler {
	return &SceneHandler{
		db:           db,
		sceneService: services2.NewStoryboardCompositionService(db, log, imageGenService, aiService, cfg),
		log:          log,
		aiService:    aiService,
		config:       cfg,
	}
}

func (h *SceneHandler) GetStoryboardsForEpisode(c *gin.Context) {
	episodeID := c.Param("episode_id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyEpisodeTeam(h.db, episodeID, teamID); err != nil {
		response.Forbidden(c, "无权访问该章节")
		return
	}

	storyboards, err := h.sceneService.GetScenesForEpisode(episodeID)
	if err != nil {
		h.log.Errorw("Failed to get storyboards for episode", "error", err, "episode_id", episodeID)
		response.InternalError(c, "获取分镜失败")
		return
	}

	response.Success(c, gin.H{
		"storyboards": storyboards,
		"total":       len(storyboards),
	})
}

func (h *SceneHandler) UpdateScene(c *gin.Context) {
	sceneID := c.Param("scene_id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifySceneTeam(h.db, sceneID, teamID); err != nil {
		response.Forbidden(c, "无权访问该场景")
		return
	}

	var req services2.UpdateSceneInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.sceneService.UpdateSceneInfo(sceneID, &req); err != nil {
		h.log.Errorw("Failed to update scene", "error", err, "scene_id", sceneID)
		response.InternalError(c, "更新场景失败")
		return
	}

	response.Success(c, gin.H{"message": "Scene updated successfully"})
}

func (h *SceneHandler) GenerateSceneImage(c *gin.Context) {
	var req services2.GenerateSceneImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifySceneTeam(h.db, fmt.Sprintf("%d", req.SceneID), teamID); err != nil {
		response.Forbidden(c, "无权访问该场景")
		return
	}

	imageGen, prompt, err := h.sceneService.GenerateSceneImage(&req)
	if err != nil {
		h.log.Errorw("Failed to generate scene image", "error", err)
		response.InternalError(c, "生成场景图片失败")
		return
	}

	response.Success(c, gin.H{
		"message":          "Scene image generation started",
		"image_generation": imageGen,
		"prompt":           prompt,
	})
}

func (h *SceneHandler) UpdateScenePrompt(c *gin.Context) {
	sceneID := c.Param("scene_id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifySceneTeam(h.db, sceneID, teamID); err != nil {
		response.Forbidden(c, "无权访问该场景")
		return
	}

	var req services2.UpdateScenePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.sceneService.UpdateScenePrompt(sceneID, &req); err != nil {
		h.log.Errorw("Failed to update scene prompt", "error", err, "scene_id", sceneID)
		if errors.Is(err, domain.ErrSceneNotFound) {
			response.NotFound(c, "场景不存在")
			return
		}
		response.InternalError(c, "更新场景提示词失败")
		return
	}

	response.Success(c, gin.H{"message": "场景提示词已更新"})
}

func (h *SceneHandler) DeleteScene(c *gin.Context) {
	sceneID := c.Param("scene_id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifySceneTeam(h.db, sceneID, teamID); err != nil {
		response.Forbidden(c, "无权访问该场景")
		return
	}

	if err := h.sceneService.DeleteScene(sceneID); err != nil {
		h.log.Errorw("Failed to delete scene", "error", err, "scene_id", sceneID)
		if errors.Is(err, domain.ErrSceneNotFound) {
			response.NotFound(c, "场景不存在")
			return
		}
		response.InternalError(c, "删除场景失败")
		return
	}

	response.Success(c, gin.H{"message": "场景已删除"})
}

func (h *SceneHandler) ListScenes(c *gin.Context) {
	dramaID := c.Query("drama_id")

	if dramaID == "" {
		response.BadRequest(c, "drama_id is required")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, dramaID, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	scenes, err := h.sceneService.GetScenesByDramaID(dramaID)
	if err != nil {
		h.log.Errorw("Failed to get scenes", "error", err, "drama_id", dramaID)
		response.InternalError(c, "获取场景失败")
		return
	}

	response.Success(c, scenes)
}

func (h *SceneHandler) CreateScene(c *gin.Context) {
	var req services2.CreateSceneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.DramaID == 0 {
		response.BadRequest(c, "drama_id is required")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, fmt.Sprintf("%d", req.DramaID), teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	scene, err := h.sceneService.CreateScene(&req)
	if err != nil {
		h.log.Errorw("Failed to create scene", "error", err)
		response.InternalError(c, "创建场景失败")
		return
	}

	response.Success(c, scene)
}

func (h *SceneHandler) PolishPrompt(c *gin.Context) {
	var req struct {
		Prompt          string   `json:"prompt" binding:"required"`
		Type            string   `json:"type" binding:"required"`
		Orientation     string   `json:"orientation" binding:"required"`
		Style           string   `json:"style" binding:"required"`
		ReferenceImages []string `json:"reference_images"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	polishedPrompt, err := h.sceneService.PolishPrompt(req.Prompt, req.Type, req.Orientation, req.Style, req.ReferenceImages)
	if err != nil {
		h.log.Errorw("Failed to polish prompt", "error", err)
		response.InternalError(c, "优化提示词失败")
		return
	}

	response.Success(c, gin.H{
		"polished_prompt": polishedPrompt,
	})
}
