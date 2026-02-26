package handlers

import (
	"fmt"
	"strconv"

	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/infrastructure/storage"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ImageGenerationHandler struct {
	imageService *services.ImageGenerationService
	taskService  *services.TaskService
	log          *logger.Logger
	config       *config.Config
	db           *gorm.DB
}

func NewImageGenerationHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger, transferService *services.ResourceTransferService, localStorage storage.Storage) *ImageGenerationHandler {
	return &ImageGenerationHandler{
		imageService: services.NewImageGenerationService(db, cfg, transferService, localStorage, log),
		taskService:  services.NewTaskService(db, log),
		log:          log,
		config:       cfg,
		db:           db,
	}
}

func (h *ImageGenerationHandler) GenerateImage(c *gin.Context) {

	var req services.GenerateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, req.DramaID, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	imageGen, err := h.imageService.GenerateImage(&req)
	if err != nil {
		h.log.Errorw("Failed to generate image", "error", err)
		response.InternalError(c, "生成图片失败")
		return
	}

	response.Success(c, imageGen)
}

// PreviewImagePrompt 预览最终发送给模型的完整 prompt（Debug 用，不实际生成）
func (h *ImageGenerationHandler) PreviewImagePrompt(c *gin.Context) {
	var req services.GenerateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, req.DramaID, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	result, err := h.imageService.PreviewImagePrompt(&req)
	if err != nil {
		h.log.Errorw("Failed to preview prompt", "error", err)
		response.InternalError(c, "预览失败")
		return
	}

	response.Success(c, result)
}

// DEPRECATED: 注释掉场景图片生成handler - 目前只使用首帧图片生成
// func (h *ImageGenerationHandler) GenerateImagesForScene(c *gin.Context) {
//
// 	sceneID := c.Param("scene_id")
//
// 	images, err := h.imageService.GenerateImagesForScene(sceneID)
// 	if err != nil {
// 		h.log.Errorw("Failed to generate images for scene", "error", err)
// 		response.InternalError(c, err.Error())
// 		return
// 	}
//
// 	response.Success(c, images)
// }

func (h *ImageGenerationHandler) GetBackgroundsForEpisode(c *gin.Context) {

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

	backgrounds, err := h.imageService.GetScencesForEpisode(episodeID)
	if err != nil {
		h.log.Errorw("Failed to get backgrounds", "error", err)
		response.InternalError(c, "获取场景失败")
		return
	}

	response.Success(c, backgrounds)
}

func (h *ImageGenerationHandler) ExtractBackgroundsForEpisode(c *gin.Context) {
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

	// 接收可选的 model 和 style 参数
	var req struct {
		Model string `json:"model"`
		Style string `json:"style"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有提供body或者解析失败，使用空字符串（使用默认模型和风格）
		req.Model = ""
		req.Style = ""
	}
	// 如果style为空，从episode获取drama的style
	if req.Style == "" {
		var episode models.Episode
		if err := h.db.Preload("Drama").First(&episode, episodeID).Error; err == nil {
			req.Style = episode.Drama.Style
		}
	}

	// 直接调用服务层的异步方法，该方法会创建任务并返回任务ID
	taskID, err := h.imageService.ExtractBackgroundsForEpisode(episodeID, req.Model, req.Style)
	if err != nil {
		h.log.Errorw("Failed to extract backgrounds", "error", err, "episode_id", episodeID)
		response.InternalError(c, "提取场景失败")
		return
	}

	// 立即返回任务ID
	response.Success(c, gin.H{
		"task_id": taskID,
		"status":  "pending",
		"message": "场景提取任务已创建，正在后台处理...",
	})
}

func (h *ImageGenerationHandler) BatchGenerateForEpisode(c *gin.Context) {

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

	// 从 query 参数获取选项
	skipExisting := c.Query("skip_existing") == "true"
	frameType := c.Query("frame_type") // first, key, last, panel, action 或空

	images, err := h.imageService.BatchGenerateImagesForEpisode(episodeID, skipExisting, frameType)
	if err != nil {
		h.log.Errorw("Failed to batch generate images", "error", err)
		response.InternalError(c, "批量生成图片失败")
		return
	}

	response.Success(c, images)
}

func (h *ImageGenerationHandler) GetImageGeneration(c *gin.Context) {

	imageGenID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyImageGenTeam(h.db, uint(imageGenID), teamID); err != nil {
		response.Forbidden(c, "无权访问该图片")
		return
	}

	imageGen, err := h.imageService.GetImageGeneration(uint(imageGenID))
	if err != nil {
		response.NotFound(c, "图片生成记录不存在")
		return
	}

	response.Success(c, imageGen)
}

func (h *ImageGenerationHandler) ListImageGenerations(c *gin.Context) {
	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	var sceneID *uint
	if sceneIDStr := c.Query("scene_id"); sceneIDStr != "" {
		id, err := strconv.ParseUint(sceneIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			sceneID = &uid
		}
	}

	var storyboardID *uint
	if storyboardIDStr := c.Query("storyboard_id"); storyboardIDStr != "" {
		id, err := strconv.ParseUint(storyboardIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			storyboardID = &uid
		}
	}

	frameType := c.Query("frame_type")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var dramaIDUint *uint
	if dramaIDStr := c.Query("drama_id"); dramaIDStr != "" {
		did, _ := strconv.ParseUint(dramaIDStr, 10, 32)
		didUint := uint(did)
		dramaIDUint = &didUint
		if _, err := auth.VerifyDramaTeam(h.db, dramaIDStr, teamID); err != nil {
			response.Forbidden(c, "无权访问该剧本")
			return
		}
	}

	var teamIDFilter *uint
	if dramaIDUint == nil {
		teamIDFilter = &teamID
	}

	images, total, err := h.imageService.ListImageGenerations(dramaIDUint, teamIDFilter, sceneID, storyboardID, frameType, status, page, pageSize)

	if err != nil {
		h.log.Errorw("Failed to list images", "error", err)
		response.InternalError(c, "获取图片列表失败")
		return
	}

	response.SuccessWithPagination(c, images, total, page, pageSize)
}

func (h *ImageGenerationHandler) DeleteImageGeneration(c *gin.Context) {

	imageGenID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyImageGenTeam(h.db, uint(imageGenID), teamID); err != nil {
		response.Forbidden(c, "无权访问该图片")
		return
	}

	if err := h.imageService.DeleteImageGeneration(uint(imageGenID)); err != nil {
		h.log.Errorw("Failed to delete image", "error", err)
		response.InternalError(c, "删除图片失败")
		return
	}

	response.Success(c, nil)
}

// EditImage 使用 Seededit 对已有图片进行编辑
func (h *ImageGenerationHandler) EditImage(c *gin.Context) {
	var req services.EditImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	imageGen, err := h.imageService.EditImage(&req)
	if err != nil {
		h.log.Errorw("Failed to edit image", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, imageGen)
}

// ListEditsBySource 列出某张原图的所有编辑结果
func (h *ImageGenerationHandler) ListEditsBySource(c *gin.Context) {
	sourceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyImageGenTeam(h.db, uint(sourceID), teamID); err != nil {
		response.Forbidden(c, "无权访问该图片")
		return
	}

	edits, err := h.imageService.ListEditsBySourceImage(uint(sourceID))
	if err != nil {
		h.log.Errorw("Failed to list edits", "error", err)
		response.InternalError(c, "获取编辑列表失败")
		return
	}

	response.Success(c, edits)
}

// ReplaceWithEdit 用编辑结果替换原图
func (h *ImageGenerationHandler) ReplaceWithEdit(c *gin.Context) {
	sourceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyImageGenTeam(h.db, uint(sourceID), teamID); err != nil {
		response.Forbidden(c, "无权访问该图片")
		return
	}

	var req struct {
		EditImageID uint `json:"edit_image_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效")
		return
	}

	if err := h.imageService.ReplaceImageWithEdit(uint(sourceID), req.EditImageID); err != nil {
		h.log.Errorw("Failed to replace image", "error", err)
		response.InternalError(c, "替换图片失败")
		return
	}

	response.Success(c, nil)
}

// UploadImage 上传图片并创建图片生成记录
func (h *ImageGenerationHandler) UploadImage(c *gin.Context) {
	var req struct {
		StoryboardID uint   `json:"storyboard_id" binding:"required"`
		DramaID      uint   `json:"drama_id" binding:"required"`
		FrameType    string `json:"frame_type" binding:"required"`
		ImageURL     string `json:"image_url" binding:"required"`
		Prompt       string `json:"prompt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效")
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

	imageGen, err := h.imageService.CreateImageFromUpload(&services.UploadImageRequest{
		StoryboardID: req.StoryboardID,
		DramaID:      req.DramaID,
		FrameType:    req.FrameType,
		ImageURL:     req.ImageURL,
		Prompt:       req.Prompt,
	})

	if err != nil {
		h.log.Errorw("Failed to create image from upload", "error", err)
		response.InternalError(c, "上传图片失败")
		return
	}

	response.Success(c, imageGen)
}
