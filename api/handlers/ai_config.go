package handlers

import (
	"errors"
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

type AIConfigHandler struct {
	db        *gorm.DB
	aiService *services.AIService
	log       *logger.Logger
}

func NewAIConfigHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *AIConfigHandler {
	return &AIConfigHandler{
		db:        db,
		aiService: services.NewAIService(db, log),
		log:       log,
	}
}

func (h *AIConfigHandler) CreateConfig(c *gin.Context) {
	var req services.CreateAIConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	req.TeamID = &teamID

	config, err := h.aiService.CreateConfig(&req)
	if err != nil {
		response.InternalError(c, "创建失败")
		return
	}

	response.Created(c, config)
}

func (h *AIConfigHandler) GetConfig(c *gin.Context) {

	teamID, tErr := auth.MustGetTeamID(c)
	if tErr != nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的配置ID")
		return
	}

	cfg, err := h.aiService.GetConfig(uint(configID))
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			response.NotFound(c, "配置不存在")
			return
		}
		response.InternalError(c, "获取失败")
		return
	}
	if cfg.TeamID != nil && *cfg.TeamID != teamID {
		response.Forbidden(c, "无权访问该配置")
		return
	}

	response.Success(c, cfg)
}

func (h *AIConfigHandler) ListConfigs(c *gin.Context) {
	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	serviceType := c.Query("service_type")

	var configs []models.AIServiceConfig
	// 包含 team_id 匹配 或 team_id 为 NULL（旧数据兼容）
	query := h.db.Model(&models.AIServiceConfig{}).Where("team_id = ? OR team_id IS NULL", teamID)
	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}
	if err := query.Order("priority DESC, created_at DESC").Find(&configs).Error; err != nil {
		response.InternalError(c, "获取列表失败")
		return
	}

	response.Success(c, configs)
}

func (h *AIConfigHandler) UpdateConfig(c *gin.Context) {

	teamID, tErr := auth.MustGetTeamID(c)
	if tErr != nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的配置ID")
		return
	}

	existing, err := h.aiService.GetConfig(uint(configID))
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			response.NotFound(c, "配置不存在")
			return
		}
		response.InternalError(c, "获取失败")
		return
	}
	if existing.TeamID != nil && *existing.TeamID != teamID {
		response.Forbidden(c, "无权修改该配置")
		return
	}

	var req services.UpdateAIConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	cfg, err := h.aiService.UpdateConfig(uint(configID), &req)
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			response.NotFound(c, "配置不存在")
			return
		}
		response.InternalError(c, "更新失败")
		return
	}

	response.Success(c, cfg)
}

func (h *AIConfigHandler) DeleteConfig(c *gin.Context) {
	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的配置ID")
		return
	}

	existing, err := h.aiService.GetConfig(uint(configID))
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			response.NotFound(c, "配置不存在")
			return
		}
		response.InternalError(c, "获取失败")
		return
	}
	if existing.TeamID != nil && *existing.TeamID != teamID {
		response.Forbidden(c, "无权删除该配置")
		return
	}

	if err := h.aiService.DeleteConfig(uint(configID)); err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			response.NotFound(c, "配置不存在")
			return
		}
		response.InternalError(c, "删除失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *AIConfigHandler) TestConnection(c *gin.Context) {
	var req services.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.aiService.TestConnection(&req); err != nil {
		response.BadRequest(c, "连接测试失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "连接测试成功"})
}

func (h *AIConfigHandler) TestConnectionAll(c *gin.Context) {
	var req services.TestConnectionAllRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.aiService.TestConnectionAll(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "全部模型连通测试通过"})
}
