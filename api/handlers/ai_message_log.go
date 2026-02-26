package handlers

import (
	"strconv"

	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AIMessageLogHandler struct {
	logService *services.AIMessageLogService
	db         *gorm.DB
	log        *logger.Logger
}

func NewAIMessageLogHandler(db *gorm.DB, logService *services.AIMessageLogService, log *logger.Logger) *AIMessageLogHandler {
	return &AIMessageLogHandler{
		logService: logService,
		db:         db,
		log:        log,
	}
}

func (h *AIMessageLogHandler) List(c *gin.Context) {
	teamID := auth.GetTeamID(c)

	var query services.AIMessageLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}
	query.TeamID = teamID

	result, err := h.logService.List(query)
	if err != nil {
		h.log.Errorw("Failed to list AI message logs", "error", err)
		response.InternalError(c, "Failed to list AI message logs")
		return
	}

	response.Success(c, result)
}

func (h *AIMessageLogHandler) GetByID(c *gin.Context) {
	teamID := auth.GetTeamID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := auth.VerifyAIMessageLogTeam(h.db, uint(id), teamID); err != nil {
		response.Forbidden(c, "无权访问该日志")
		return
	}

	log, err := h.logService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "AI message log not found")
		return
	}

	response.Success(c, log)
}

func (h *AIMessageLogHandler) GetStats(c *gin.Context) {
	teamID := auth.GetTeamID(c)

	stats, err := h.logService.GetStats(teamID)
	if err != nil {
		h.log.Errorw("Failed to get AI message log stats", "error", err)
		response.InternalError(c, "Failed to get stats")
		return
	}

	response.Success(c, stats)
}
