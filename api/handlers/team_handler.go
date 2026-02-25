package handlers

import (
	"strconv"

	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	authService *services.AuthService
	log         *logger.Logger
}

func NewTeamHandler(authService *services.AuthService, log *logger.Logger) *TeamHandler {
	return &TeamHandler{authService: authService, log: log}
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	team, err := h.authService.GetTeam(teamID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, team)
}

func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	userID := auth.GetUserID(c)
	var req struct {
		Name string `json:"name" binding:"required,min=1,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	team, err := h.authService.UpdateTeam(teamID, userID, req.Name)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, team)
}

func (h *TeamHandler) InviteMember(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	userID := auth.GetUserID(c)
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	inv, err := h.authService.InviteMember(teamID, userID, req.Email)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, inv)
}

func (h *TeamHandler) AcceptInvitation(c *gin.Context) {
	userID := auth.GetUserID(c)
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authService.AcceptInvitation(req.Token, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "已加入团队", nil)
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	userID := auth.GetUserID(c)
	memberID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的成员ID")
		return
	}

	if err := h.authService.RemoveMember(teamID, userID, uint(memberID)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "已移除成员", nil)
}
