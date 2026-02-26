package handlers

import (
	"fmt"
	"strconv"

	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/infrastructure/storage"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VideoGenerationHandler struct {
	videoService *services.VideoGenerationService
	db           *gorm.DB
	log          *logger.Logger
}

func NewVideoGenerationHandler(db *gorm.DB, transferService *services.ResourceTransferService, localStorage storage.Storage, aiService *services.AIService, log *logger.Logger, promptI18n *services.PromptI18n) *VideoGenerationHandler {
	return &VideoGenerationHandler{
		videoService: services.NewVideoGenerationService(db, transferService, localStorage, aiService, log, promptI18n),
		db:           db,
		log:          log,
	}
}

func (h *VideoGenerationHandler) GenerateVideo(c *gin.Context) {
	teamID := auth.GetTeamID(c)

	var req services.GenerateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if _, err := auth.VerifyDramaTeam(h.db, req.DramaID, teamID); err != nil {
		response.Forbidden(c, "无权操作该剧本")
		return
	}

	videoGen, err := h.videoService.GenerateVideo(&req)
	if err != nil {
		h.log.Errorw("Failed to generate video", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, videoGen)
}

func (h *VideoGenerationHandler) GenerateVideoFromImage(c *gin.Context) {
	teamID := auth.GetTeamID(c)

	imageGenID, err := strconv.ParseUint(c.Param("image_gen_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的图片ID")
		return
	}

	if err := auth.VerifyImageGenTeam(h.db, uint(imageGenID), teamID); err != nil {
		response.Forbidden(c, "无权操作该图片")
		return
	}

	videoGen, err := h.videoService.GenerateVideoFromImage(uint(imageGenID))
	if err != nil {
		h.log.Errorw("Failed to generate video from image", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, videoGen)
}

func (h *VideoGenerationHandler) BatchGenerateForEpisode(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	episodeID := c.Param("episode_id")

	if _, err := auth.VerifyEpisodeTeam(h.db, episodeID, teamID); err != nil {
		response.Forbidden(c, "无权操作该章节")
		return
	}

	skipExisting := c.Query("skip_existing") == "true"

	videos, err := h.videoService.BatchGenerateVideosForEpisode(episodeID, skipExisting)
	if err != nil {
		h.log.Errorw("Failed to batch generate videos", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, videos)
}

func (h *VideoGenerationHandler) GetVideoGeneration(c *gin.Context) {
	teamID := auth.GetTeamID(c)

	videoGenID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := auth.VerifyVideoGenTeam(h.db, uint(videoGenID), teamID); err != nil {
		response.Forbidden(c, "无权访问该视频")
		return
	}

	videoGen, err := h.videoService.GetVideoGeneration(uint(videoGenID))
	if err != nil {
		response.NotFound(c, "视频生成记录不存在")
		return
	}

	response.Success(c, videoGen)
}

func (h *VideoGenerationHandler) ListVideoGenerations(c *gin.Context) {
	teamID := auth.GetTeamID(c)

	var storyboardID *uint
	if storyboardIDStr := c.Query("storyboard_id"); storyboardIDStr != "" {
		id, err := strconv.ParseUint(storyboardIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			storyboardID = &uid
		}
	}
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
		if _, err := auth.VerifyDramaTeam(h.db, dramaIDStr, teamID); err != nil {
			response.Forbidden(c, "无权访问该剧本")
			return
		}
		did, _ := strconv.ParseUint(dramaIDStr, 10, 32)
		didUint := uint(did)
		dramaIDUint = &didUint
	}

	if storyboardID != nil {
		if err := auth.VerifyStoryboardTeam(h.db, fmt.Sprintf("%d", *storyboardID), teamID); err != nil {
			response.Forbidden(c, "无权访问该分镜")
			return
		}
	}

	offset := (page - 1) * pageSize
	videos, total, err := h.videoService.ListVideoGenerations(dramaIDUint, storyboardID, status, pageSize, offset)

	if err != nil {
		h.log.Errorw("Failed to list videos", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPagination(c, videos, total, page, pageSize)
}

// ChainGenerateForEpisode V3 链式视频生成（已禁用）
// func (h *VideoGenerationHandler) ChainGenerateForEpisode(c *gin.Context) {
// 	...
// }

// ExtractLastFrame 从指定分镜的已完成视频截取尾帧
// 支持 query 参数 video_id 指定具体视频，同时返回该分镜所有已完成视频列表
func (h *VideoGenerationHandler) ExtractLastFrame(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	storyboardID, err := strconv.ParseUint(c.Param("storyboard_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的分镜ID")
		return
	}

	if err := auth.VerifyStoryboardTeam(h.db, fmt.Sprintf("%d", storyboardID), teamID); err != nil {
		response.Forbidden(c, "无权操作该分镜")
		return
	}

	var videoID uint
	if vidStr := c.Query("video_id"); vidStr != "" {
		if vid, err := strconv.ParseUint(vidStr, 10, 32); err == nil {
			videoID = uint(vid)
		}
	}

	// 查询该分镜所有已完成视频（用于前端选择器）
	var completedVideos []struct {
		ID        uint   `json:"id"`
		CreatedAt string `json:"created_at"`
		Duration  *int   `json:"duration"`
		Model     string `json:"model"`
	}
	h.db.Raw(`SELECT id, created_at, duration, model FROM video_generations 
		WHERE storyboard_id = ? AND status = 'completed' ORDER BY created_at DESC`, storyboardID).Scan(&completedVideos)

	framePath, video, err := h.videoService.GetLastFrameFromStoryboardVideo(uint(storyboardID), videoID)
	if err != nil {
		h.log.Warnw("Failed to extract last frame", "storyboard_id", storyboardID, "error", err)
		if video == nil {
			response.Success(c, gin.H{
				"success":    false,
				"has_video":  false,
				"message":    err.Error(),
				"videos":     completedVideos,
			})
		} else {
			response.Success(c, gin.H{
				"success":    false,
				"has_video":  true,
				"message":    err.Error(),
				"videos":     completedVideos,
			})
		}
		return
	}

	response.Success(c, gin.H{
		"success":    true,
		"has_video":  true,
		"frame_path": framePath,
		"video_id":   video.ID,
		"message":    "尾帧截取成功",
		"videos":     completedVideos,
	})
}

func (h *VideoGenerationHandler) DeleteVideoGeneration(c *gin.Context) {
	teamID := auth.GetTeamID(c)

	videoGenID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := auth.VerifyVideoGenTeam(h.db, uint(videoGenID), teamID); err != nil {
		response.Forbidden(c, "无权删除该视频")
		return
	}

	if err := h.videoService.DeleteVideoGeneration(uint(videoGenID)); err != nil {
		h.log.Errorw("Failed to delete video", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *VideoGenerationHandler) BackfillVideosToCOS(c *gin.Context) {
	uploaded, skipped, err := h.videoService.BackfillVideosToCOS()
	if err != nil {
		h.log.Errorw("Backfill failed", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"uploaded": uploaded,
		"skipped":  skipped,
	})
}
