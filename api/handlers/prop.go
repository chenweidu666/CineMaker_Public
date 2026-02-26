package handlers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/domain"
	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PropHandler struct {
	db          *gorm.DB
	propService *services.PropService
	log         *logger.Logger
}

func NewPropHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger, aiService *services.AIService, imageGenerationService *services.ImageGenerationService) *PropHandler {
	return &PropHandler{
		db:          db,
		propService: services.NewPropService(db, aiService, services.NewTaskService(db, log), imageGenerationService, log, cfg),
		log:         log,
	}
}

// ListProps 获取道具列表
func (h *PropHandler) ListProps(c *gin.Context) {
	dramaIDStr := c.Param("id")
	if dramaIDStr == "" {
		response.BadRequest(c, "drama_id is required")
		return
	}

	dramaID, err := strconv.ParseUint(dramaIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid drama_id")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, dramaIDStr, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	props, err := h.propService.ListProps(uint(dramaID))
	if err != nil {
		response.InternalError(c, "获取道具列表失败")
		return
	}

	response.Success(c, props)
}

// CreateProp 创建道具
func (h *PropHandler) CreateProp(c *gin.Context) {
	var prop models.Prop
	if err := c.ShouldBindJSON(&prop); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, fmt.Sprintf("%d", prop.DramaID), teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	if err := h.propService.CreateProp(&prop); err != nil {
		response.InternalError(c, "创建道具失败")
		return
	}

	response.Created(c, prop)
}

// UpdateProp 更新道具
func (h *PropHandler) UpdateProp(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyPropTeam(h.db, idStr, teamID); err != nil {
		response.Forbidden(c, "无权访问该道具")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.propService.UpdateProp(uint(id), updates); err != nil {
		response.InternalError(c, "更新道具失败")
		return
	}

	response.Success(c, nil)
}

// DeleteProp 删除道具
func (h *PropHandler) DeleteProp(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyPropTeam(h.db, idStr, teamID); err != nil {
		response.Forbidden(c, "无权访问该道具")
		return
	}

	if err := h.propService.DeleteProp(uint(id)); err != nil {
		h.log.Errorw("Failed to delete prop", "error", err, "id", id)
		response.InternalError(c, "删除道具失败")
		return
	}

	response.Success(c, gin.H{"message": "道具已删除"})
}

// ExtractProps 提取道具
func (h *PropHandler) ExtractProps(c *gin.Context) {
	episodeIDStr := c.Param("episode_id")
	episodeID, err := strconv.ParseUint(episodeIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid episode_id")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyEpisodeTeam(h.db, episodeIDStr, teamID); err != nil {
		response.Forbidden(c, "无权访问该章节")
		return
	}

	taskID, err := h.propService.ExtractPropsFromScript(uint(episodeID))
	if err != nil {
		response.InternalError(c, "提取道具失败")
		return
	}

	response.Success(c, gin.H{"task_id": taskID})
}

// GenerateImage 生成道具图片
func (h *PropHandler) GenerateImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyPropTeam(h.db, idStr, teamID); err != nil {
		response.Forbidden(c, "无权访问该道具")
		return
	}

	var req struct {
		Prompt          string   `json:"prompt"`
		ReferenceImages []string `json:"reference_images"`
	}
	// 允许空body，使用数据库中的默认值
	_ = c.ShouldBindJSON(&req)

	taskID, err := h.propService.GeneratePropImage(uint(id), req.Prompt, req.ReferenceImages)
	if err != nil {
		h.log.Errorw("Failed to generate prop image", "error", err)
		if errors.Is(err, domain.ErrPropNotFound) {
			response.NotFound(c, "道具不存在")
			return
		}
		response.InternalError(c, "生成道具图片失败")
		return
	}

	response.Success(c, gin.H{"task_id": taskID, "message": "图片生成任务已提交"})
}

// AssociateProps 关联道具
func (h *PropHandler) AssociateProps(c *gin.Context) {
	storyboardIDStr := c.Param("id")
	storyboardID, err := strconv.ParseUint(storyboardIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid storyboard_id")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyStoryboardTeam(h.db, storyboardIDStr, teamID); err != nil {
		response.Forbidden(c, "无权访问该分镜")
		return
	}

	var req struct {
		PropIDs []uint `json:"prop_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.propService.AssociatePropsWithStoryboard(uint(storyboardID), req.PropIDs); err != nil {
		response.InternalError(c, "关联道具失败")
		return
	}

	response.Success(c, nil)
}
