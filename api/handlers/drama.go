package handlers

import (
	"encoding/json"
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

type DramaHandler struct {
	db                *gorm.DB
	dramaService      *services.DramaService
	videoMergeService *services.VideoMergeService
	log               *logger.Logger
}

func NewDramaHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger, transferService *services.ResourceTransferService) *DramaHandler {
	return &DramaHandler{
		db:                db,
		dramaService:      services.NewDramaService(db, cfg, log),
		videoMergeService: services.NewVideoMergeService(db, transferService, cfg.Storage.LocalPath, cfg.Storage.BaseURL, log),
		log:               log,
	}
}

func (h *DramaHandler) CreateDrama(c *gin.Context) {

	var req services.CreateDramaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	req.TeamID = &teamID

	drama, err := h.dramaService.CreateDrama(&req)
	if err != nil {
		response.InternalError(c, "创建失败")
		return
	}

	response.Created(c, drama)
}

func (h *DramaHandler) GetDrama(c *gin.Context) {

	dramaID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, dramaID, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	drama, err := h.dramaService.GetDrama(dramaID)
	if err != nil {
		if errors.Is(err, domain.ErrDramaNotFound) {
			response.NotFound(c, "剧本不存在")
			return
		}
		response.InternalError(c, "获取失败")
		return
	}

	response.Success(c, drama)
}

func (h *DramaHandler) ListDramas(c *gin.Context) {

	var query services.DramaListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	query.TeamID = &teamID

	if query.Page < models.MinPageSize {
		query.Page = models.MinPageSize
	}
	if query.PageSize < models.MinPageSize || query.PageSize > models.MaxPageSize {
		query.PageSize = models.DefaultPageSize
	}

	dramas, total, err := h.dramaService.ListDramas(&query)
	if err != nil {
		response.InternalError(c, "获取列表失败")
		return
	}

	response.SuccessWithPagination(c, dramas, total, query.Page, query.PageSize)
}

func (h *DramaHandler) UpdateDrama(c *gin.Context) {

	dramaID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, dramaID, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	var req services.UpdateDramaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	drama, err := h.dramaService.UpdateDrama(dramaID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrDramaNotFound) {
			response.NotFound(c, "剧本不存在")
			return
		}
		response.InternalError(c, "更新失败")
		return
	}

	response.Success(c, drama)
}

func (h *DramaHandler) DeleteDrama(c *gin.Context) {

	dramaID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, dramaID, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	if err := h.dramaService.DeleteDrama(dramaID); err != nil {
		if errors.Is(err, domain.ErrDramaNotFound) {
			response.NotFound(c, "剧本不存在")
			return
		}
		response.InternalError(c, "删除失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

func (h *DramaHandler) GetDramaStats(c *gin.Context) {

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	stats, err := h.dramaService.GetDramaStatsByTeam(teamID)
	if err != nil {
		response.InternalError(c, "获取统计失败")
		return
	}

	response.Success(c, stats)
}

func (h *DramaHandler) SaveOutline(c *gin.Context) {

	dramaID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, dramaID, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	var req services.SaveOutlineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.dramaService.SaveOutline(dramaID, &req); err != nil {
		if errors.Is(err, domain.ErrDramaNotFound) {
			response.NotFound(c, "剧本不存在")
			return
		}
		response.InternalError(c, "保存失败")
		return
	}

	response.Success(c, gin.H{"message": "保存成功"})
}

func (h *DramaHandler) GetCharacters(c *gin.Context) {

	dramaID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, dramaID, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	episodeID := c.Query("episode_id")

	var episodeIDPtr *string
	if episodeID != "" {
		episodeIDPtr = &episodeID
	}

	characters, err := h.dramaService.GetCharacters(dramaID, episodeIDPtr)
	if err != nil {
		if errors.Is(err, domain.ErrDramaNotFound) {
			response.NotFound(c, "剧本不存在")
			return
		}
		if errors.Is(err, domain.ErrEpisodeNotFound) {
			response.NotFound(c, "章节不存在")
			return
		}
		response.InternalError(c, "获取角色失败")
		return
	}

	response.Success(c, characters)
}

func (h *DramaHandler) SaveCharacters(c *gin.Context) {
	dramaID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, dramaID, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	var req services.SaveCharactersRequest

	// 先尝试正常绑定JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果绑定失败，检查是否是因为characters字段是字符串而不是数组
		var rawReq map[string]interface{}
		if err := c.ShouldBindJSON(&rawReq); err != nil {
			// 如果连rawReq都绑定失败，直接返回错误
			response.BadRequest(c, err.Error())
			return
		}

		// 检查characters字段类型
		if charField, ok := rawReq["characters"]; ok {
			if charStr, ok := charField.(string); ok {
				// 如果characters是字符串，尝试解析为JSON数组
				var characters []models.Character
				if err := json.Unmarshal([]byte(charStr), &characters); err != nil {
					// 解析失败，返回错误
					response.BadRequest(c, "characters字段格式错误，需要JSON数组或字符串格式的JSON数组")
					return
				}

				// 手动构造请求对象
				req.Characters = characters

				// 处理episode_id字段
				if epID, ok := rawReq["episode_id"]; ok {
					if epIDStr, ok := epID.(float64); ok {
						epIDUint := uint(epIDStr)
						req.EpisodeID = &epIDUint
					}
				}
			} else {
				// 如果characters不是字符串，直接返回原始错误
				response.BadRequest(c, err.Error())
				return
			}
		} else {
			// 如果没有characters字段，返回原始错误
			response.BadRequest(c, err.Error())
			return
		}
	}

	if err := h.dramaService.SaveCharacters(dramaID, &req); err != nil {
		if errors.Is(err, domain.ErrDramaNotFound) {
			response.NotFound(c, "剧本不存在")
			return
		}
		response.InternalError(c, "保存失败")
		return
	}

	response.Success(c, gin.H{"message": "保存成功"})
}

func (h *DramaHandler) SaveEpisodes(c *gin.Context) {

	dramaID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, dramaID, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	h.log.Infow("SaveEpisodes handler called", "drama_id", dramaID)

	var req services.SaveEpisodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorw("Failed to bind JSON", "error", err)
		response.BadRequest(c, err.Error())
		return
	}

	h.log.Infow("Request received", "episodes_count", len(req.Episodes))
	for i, ep := range req.Episodes {
		h.log.Infow("Episode data", "index", i, "episode_num", ep.EpisodeNum, "title", ep.Title)
	}

	if err := h.dramaService.SaveEpisodes(dramaID, &req); err != nil {
		if errors.Is(err, domain.ErrDramaNotFound) {
			response.NotFound(c, "剧本不存在")
			return
		}
		response.InternalError(c, "保存失败")
		return
	}

	response.Success(c, gin.H{"message": "保存成功"})
}

func (h *DramaHandler) UpdateEpisodeTitle(c *gin.Context) {
	dramaID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, dramaID, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	h.log.Infow("UpdateEpisodeTitle handler called", "drama_id", dramaID)

	var req services.UpdateEpisodeTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Errorw("Failed to bind JSON", "error", err)
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.dramaService.UpdateEpisodeTitle(dramaID, &req); err != nil {
		if errors.Is(err, domain.ErrDramaNotFound) || errors.Is(err, domain.ErrEpisodeNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, "更新失败")
		return
	}

	response.Success(c, gin.H{"message": "章节名称更新成功"})
}

func (h *DramaHandler) SaveProgress(c *gin.Context) {

	dramaID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyDramaTeam(h.db, dramaID, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	var req services.SaveProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.dramaService.SaveProgress(dramaID, &req); err != nil {
		if errors.Is(err, domain.ErrDramaNotFound) {
			response.NotFound(c, "剧本不存在")
			return
		}
		response.InternalError(c, "保存失败")
		return
	}

	response.Success(c, gin.H{"message": "保存成功"})
}

// FinalizeEpisode 完成集数制作（触发视频合成）
func (h *DramaHandler) FinalizeEpisode(c *gin.Context) {

	episodeID := c.Param("episode_id")
	if episodeID == "" {
		response.BadRequest(c, "episode_id不能为空")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyEpisodeTeam(h.db, episodeID, teamID); err != nil {
		response.Forbidden(c, "无权访问该章节")
		return
	}

	// 尝试读取时间线数据（可选）
	var timelineData *services.FinalizeEpisodeRequest
	if err := c.ShouldBindJSON(&timelineData); err != nil {
		// 如果没有请求体或解析失败，使用nil（将使用默认场景顺序）
		h.log.Warnw("No timeline data provided, will use default scene order", "error", err)
		timelineData = nil
	} else if timelineData != nil {
		h.log.Infow("Received timeline data", "clips_count", len(timelineData.Clips), "episode_id", episodeID)
	}

	// 触发视频合成任务
	result, err := h.videoMergeService.FinalizeEpisode(episodeID, timelineData)
	if err != nil {
		h.log.Errorw("Failed to finalize episode", "error", err, "episode_id", episodeID)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// DownloadEpisodeVideo 下载剧集视频
func (h *DramaHandler) DownloadEpisodeVideo(c *gin.Context) {

	episodeID := c.Param("episode_id")
	if episodeID == "" {
		response.BadRequest(c, "episode_id不能为空")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, err := auth.VerifyEpisodeTeam(h.db, episodeID, teamID); err != nil {
		response.Forbidden(c, "无权访问该章节")
		return
	}

	// 查询episode
	var episode models.Episode
	if err := h.db.Preload("Drama").Where("id = ?", episodeID).First(&episode).Error; err != nil {
		response.NotFound(c, "剧集不存在")
		return
	}

	// 检查是否有视频
	if episode.VideoURL == nil || *episode.VideoURL == "" {
		response.BadRequest(c, "该剧集还没有生成视频")
		return
	}

	// 返回视频URL，让前端重定向下载
	c.JSON(200, gin.H{
		"video_url":      *episode.VideoURL,
		"title":          episode.Title,
		"episode_number": episode.EpisodeNum,
	})
}

func (h *DramaHandler) SaveEpisodeScript(c *gin.Context) {
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

	h.log.Infow("SaveEpisodeScript called", "episode_id", episodeID)

	var req struct {
		ScriptContent string `json:"script_content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result := h.db.Model(&models.Episode{}).Where("id = ?", episodeID).Update("script_content", req.ScriptContent)
	if result.Error != nil {
		h.log.Errorw("Failed to save episode script", "error", result.Error)
		response.InternalError(c, "保存剧本失败")
		return
	}
	if result.RowsAffected == 0 {
		response.NotFound(c, "章节不存在")
		return
	}

	response.Success(c, gin.H{"message": "剧本保存成功"})
}

func (h *DramaHandler) SaveEpisodeDescription(c *gin.Context) {
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

	var req struct {
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result := h.db.Model(&models.Episode{}).Where("id = ?", episodeID).Updates(map[string]interface{}{
		"description": req.Description,
	})
	if result.Error != nil {
		response.InternalError(c, "保存剧情描述失败")
		return
	}

	response.Success(c, gin.H{"message": "剧情描述保存成功"})
}

func (h *DramaHandler) AddCharacterToEpisode(c *gin.Context) {
	teamID, tErr := auth.MustGetTeamID(c)
	if tErr != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, tErr := auth.VerifyEpisodeTeam(h.db, c.Param("episode_id"), teamID); tErr != nil {
		response.Forbidden(c, "无权访问该章节")
		return
	}
	episodeID, err := strconv.ParseUint(c.Param("episode_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的章节ID")
		return
	}
	var req struct {
		CharacterID uint `json:"character_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供 character_id")
		return
	}
	if err := h.dramaService.AddCharacterToEpisode(uint(episodeID), req.CharacterID); err != nil {
		h.log.Errorw("Failed to add character to episode", "error", err)
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "角色已关联到章节"})
}

func (h *DramaHandler) AddSceneToEpisode(c *gin.Context) {
	teamID, tErr := auth.MustGetTeamID(c)
	if tErr != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, tErr := auth.VerifyEpisodeTeam(h.db, c.Param("episode_id"), teamID); tErr != nil {
		response.Forbidden(c, "无权访问该章节")
		return
	}
	episodeID, err := strconv.ParseUint(c.Param("episode_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的章节ID")
		return
	}
	var req struct {
		SceneID uint `json:"scene_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供 scene_id")
		return
	}
	if err := h.dramaService.AddSceneToEpisode(uint(episodeID), req.SceneID); err != nil {
		h.log.Errorw("Failed to add scene to episode", "error", err)
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "场景已关联到章节"})
}

func (h *DramaHandler) AddPropToEpisode(c *gin.Context) {
	teamID, tErr := auth.MustGetTeamID(c)
	if tErr != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, tErr := auth.VerifyEpisodeTeam(h.db, c.Param("episode_id"), teamID); tErr != nil {
		response.Forbidden(c, "无权访问该章节")
		return
	}
	episodeID, err := strconv.ParseUint(c.Param("episode_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的章节ID")
		return
	}
	var req struct {
		PropID uint `json:"prop_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供 prop_id")
		return
	}
	if err := h.dramaService.AddPropToEpisode(uint(episodeID), req.PropID); err != nil {
		h.log.Errorw("Failed to add prop to episode", "error", err)
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "道具已关联到章节"})
}

func (h *DramaHandler) RemoveCharacterFromEpisode(c *gin.Context) {
	teamID, tErr := auth.MustGetTeamID(c)
	if tErr != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, tErr := auth.VerifyEpisodeTeam(h.db, c.Param("episode_id"), teamID); tErr != nil {
		response.Forbidden(c, "无权访问该章节")
		return
	}

	episodeID, err := strconv.ParseUint(c.Param("episode_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的章节ID")
		return
	}
	characterID, err := strconv.ParseUint(c.Param("character_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}
	if err := h.dramaService.RemoveCharacterFromEpisode(uint(episodeID), uint(characterID)); err != nil {
		h.log.Errorw("Failed to remove character from episode", "error", err)
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "角色已从章节移除"})
}

func (h *DramaHandler) RemoveSceneFromEpisode(c *gin.Context) {
	teamID, tErr := auth.MustGetTeamID(c)
	if tErr != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, tErr := auth.VerifyEpisodeTeam(h.db, c.Param("episode_id"), teamID); tErr != nil {
		response.Forbidden(c, "无权访问该章节")
		return
	}

	episodeID, err := strconv.ParseUint(c.Param("episode_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的章节ID")
		return
	}
	sceneID, err := strconv.ParseUint(c.Param("scene_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的场景ID")
		return
	}
	if err := h.dramaService.RemoveSceneFromEpisode(uint(episodeID), uint(sceneID)); err != nil {
		h.log.Errorw("Failed to remove scene from episode", "error", err)
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "场景已从章节移除"})
}

func (h *DramaHandler) RemovePropFromEpisode(c *gin.Context) {
	teamID, tErr := auth.MustGetTeamID(c)
	if tErr != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if _, tErr := auth.VerifyEpisodeTeam(h.db, c.Param("episode_id"), teamID); tErr != nil {
		response.Forbidden(c, "无权访问该章节")
		return
	}

	episodeID, err := strconv.ParseUint(c.Param("episode_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的章节ID")
		return
	}
	propID, err := strconv.ParseUint(c.Param("prop_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的道具ID")
		return
	}
	if err := h.dramaService.RemovePropFromEpisode(uint(episodeID), uint(propID)); err != nil {
		h.log.Errorw("Failed to remove prop from episode", "error", err)
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "道具已从章节移除"})
}

func (h *DramaHandler) CreateCharacter(c *gin.Context) {
	var req services.CreateCharacterRequest
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

	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}

	character, err := h.dramaService.CreateCharacter(&req)
	if err != nil {
		h.log.Errorw("Failed to create character", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, character)
}
