package handlers

import (
	"strconv"

	services2 "github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VideoMergeHandler struct {
	mergeService *services2.VideoMergeService
	db           *gorm.DB
	log          *logger.Logger
}

func NewVideoMergeHandler(db *gorm.DB, transferService *services2.ResourceTransferService, storagePath, baseURL string, log *logger.Logger) *VideoMergeHandler {
	return &VideoMergeHandler{
		mergeService: services2.NewVideoMergeService(db, transferService, storagePath, baseURL, log),
		db:           db,
		log:          log,
	}
}

func (h *VideoMergeHandler) MergeVideos(c *gin.Context) {
	teamID := auth.GetTeamID(c)

	var req services2.MergeVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	if _, err := auth.VerifyEpisodeTeam(h.db, req.EpisodeID, teamID); err != nil {
		response.Forbidden(c, "无权操作该章节")
		return
	}

	merge, err := h.mergeService.MergeVideos(&req)
	if err != nil {
		h.log.Errorw("Failed to merge videos", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "Video merge task created",
		"merge":   merge,
	})
}

func (h *VideoMergeHandler) GetMerge(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	mergeIDStr := c.Param("merge_id")
	mergeID, err := strconv.ParseUint(mergeIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid merge ID")
		return
	}

	if err := auth.VerifyVideoMergeTeam(h.db, uint(mergeID), teamID); err != nil {
		response.Forbidden(c, "无权访问该合成记录")
		return
	}

	merge, err := h.mergeService.GetMerge(uint(mergeID))
	if err != nil {
		h.log.Errorw("Failed to get merge", "error", err)
		response.NotFound(c, "Merge not found")
		return
	}

	response.Success(c, gin.H{"merge": merge})
}

func (h *VideoMergeHandler) ListMerges(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	episodeID := c.Query("episode_id")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if episodeID != "" {
		if _, err := auth.VerifyEpisodeTeam(h.db, episodeID, teamID); err != nil {
			response.Forbidden(c, "无权访问该章节")
			return
		}
	}

	var episodeIDPtr *string
	if episodeID != "" {
		episodeIDPtr = &episodeID
	}

	merges, total, err := h.mergeService.ListMerges(episodeIDPtr, status, page, pageSize)
	if err != nil {
		h.log.Errorw("Failed to list merges", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"merges":    merges,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *VideoMergeHandler) DeleteMerge(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	mergeIDStr := c.Param("merge_id")
	mergeID, err := strconv.ParseUint(mergeIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid merge ID")
		return
	}

	if err := auth.VerifyVideoMergeTeam(h.db, uint(mergeID), teamID); err != nil {
		response.Forbidden(c, "无权删除该合成记录")
		return
	}

	if err := h.mergeService.DeleteMerge(uint(mergeID)); err != nil {
		h.log.Errorw("Failed to delete merge", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Merge deleted successfully"})
}
