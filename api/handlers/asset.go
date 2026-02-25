package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AssetHandler struct {
	db           *gorm.DB
	assetService *services.AssetService
	log          *logger.Logger
}

func NewAssetHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger) *AssetHandler {
	return &AssetHandler{
		db:           db,
		assetService: services.NewAssetService(db, log),
		log:          log,
	}
}

func (h *AssetHandler) verifyAssetTeam(assetID uint, teamID uint) error {
	var asset struct {
		TeamID  *uint `gorm:"column:team_id"`
		DramaID *uint `gorm:"column:drama_id"`
	}
	if err := h.db.Table("assets").Select("team_id, drama_id").
		Where("id = ? AND deleted_at IS NULL", assetID).First(&asset).Error; err != nil {
		return fmt.Errorf("asset not found")
	}
	if asset.TeamID != nil && *asset.TeamID != teamID {
		return fmt.Errorf("forbidden")
	}
	if asset.DramaID != nil {
		_, err := auth.VerifyDramaTeam(h.db, fmt.Sprintf("%d", *asset.DramaID), teamID)
		return err
	}
	return nil
}

func (h *AssetHandler) CreateAsset(c *gin.Context) {

	var req services.CreateAssetRequest
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

	asset, err := h.assetService.CreateAsset(&req)
	if err != nil {
		h.log.Errorw("Failed to create asset", "error", err)
		response.InternalError(c, "创建素材失败")
		return
	}

	response.Success(c, asset)
}

func (h *AssetHandler) UpdateAsset(c *gin.Context) {

	assetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	teamID, tErr := auth.MustGetTeamID(c)
	if tErr != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := h.verifyAssetTeam(uint(assetID), teamID); err != nil {
		response.Forbidden(c, "无权访问该素材")
		return
	}

	var req services.UpdateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	asset, err := h.assetService.UpdateAsset(uint(assetID), &req)
	if err != nil {
		h.log.Errorw("Failed to update asset", "error", err)
		response.InternalError(c, "更新素材失败")
		return
	}

	response.Success(c, asset)
}

func (h *AssetHandler) GetAsset(c *gin.Context) {

	assetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	teamID, tErr := auth.MustGetTeamID(c)
	if tErr != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := h.verifyAssetTeam(uint(assetID), teamID); err != nil {
		response.Forbidden(c, "无权访问该素材")
		return
	}

	asset, err := h.assetService.GetAsset(uint(assetID))
	if err != nil {
		response.NotFound(c, "素材不存在")
		return
	}

	response.Success(c, asset)
}

func (h *AssetHandler) ListAssets(c *gin.Context) {

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	var dramaID *string
	if dramaIDStr := c.Query("drama_id"); dramaIDStr != "" {
		if _, err := auth.VerifyDramaTeam(h.db, dramaIDStr, teamID); err != nil {
			response.Forbidden(c, "无权访问该剧本")
			return
		}
		dramaID = &dramaIDStr
	}

	var episodeID *uint
	if episodeIDStr := c.Query("episode_id"); episodeIDStr != "" {
		if id, err := strconv.ParseUint(episodeIDStr, 10, 32); err == nil {
			uid := uint(id)
			episodeID = &uid
		}
	}

	var storyboardID *uint
	if storyboardIDStr := c.Query("storyboard_id"); storyboardIDStr != "" {
		if id, err := strconv.ParseUint(storyboardIDStr, 10, 32); err == nil {
			uid := uint(id)
			storyboardID = &uid
		}
	}

	var assetType *models.AssetType
	if typeStr := c.Query("type"); typeStr != "" {
		t := models.AssetType(typeStr)
		assetType = &t
	}

	var isFavorite *bool
	if favoriteStr := c.Query("is_favorite"); favoriteStr != "" {
		if favoriteStr == "true" {
			fav := true
			isFavorite = &fav
		} else if favoriteStr == "false" {
			fav := false
			isFavorite = &fav
		}
	}

	var tagIDs []uint
	if tagIDsStr := c.Query("tag_ids"); tagIDsStr != "" {
		for _, idStr := range strings.Split(tagIDsStr, ",") {
			if id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 32); err == nil {
				tagIDs = append(tagIDs, uint(id))
			}
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	req := &services.ListAssetsRequest{
		DramaID:      dramaID,
		EpisodeID:    episodeID,
		StoryboardID: storyboardID,
		Type:         assetType,
		Category:     c.Query("category"),
		TagIDs:       tagIDs,
		IsFavorite:   isFavorite,
		Search:       c.Query("search"),
		Page:         page,
		PageSize:     pageSize,
		TeamID:       &teamID,
	}

	assets, total, err := h.assetService.ListAssets(req)
	if err != nil {
		h.log.Errorw("Failed to list assets", "error", err)
		response.InternalError(c, "获取素材列表失败")
		return
	}

	response.SuccessWithPagination(c, assets, total, page, pageSize)
}

func (h *AssetHandler) DeleteAsset(c *gin.Context) {

	assetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	teamID, tErr := auth.MustGetTeamID(c)
	if tErr != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := h.verifyAssetTeam(uint(assetID), teamID); err != nil {
		response.Forbidden(c, "无权访问该素材")
		return
	}

	if err := h.assetService.DeleteAsset(uint(assetID)); err != nil {
		h.log.Errorw("Failed to delete asset", "error", err)
		response.InternalError(c, "删除素材失败")
		return
	}

	response.Success(c, nil)
}

func (h *AssetHandler) ImportFromImageGen(c *gin.Context) {

	imageGenID, err := strconv.ParseUint(c.Param("image_gen_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	teamID, tErr := auth.MustGetTeamID(c)
	if tErr != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyImageGenTeam(h.db, uint(imageGenID), teamID); err != nil {
		response.Forbidden(c, "无权访问该图片")
		return
	}

	asset, err := h.assetService.ImportFromImageGen(uint(imageGenID))
	if err != nil {
		h.log.Errorw("Failed to import from image gen", "error", err)
		response.InternalError(c, "导入素材失败")
		return
	}

	response.Success(c, asset)
}

func (h *AssetHandler) ImportFromVideoGen(c *gin.Context) {

	videoGenID, err := strconv.ParseUint(c.Param("video_gen_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	teamID, _ := auth.MustGetTeamID(c)
	if err := auth.VerifyVideoGenTeam(h.db, uint(videoGenID), teamID); err != nil {
		response.Forbidden(c, "无权操作该视频")
		return
	}

	asset, err := h.assetService.ImportFromVideoGen(uint(videoGenID))
	if err != nil {
		h.log.Errorw("Failed to import from video gen", "error", err)
		response.InternalError(c, "导入素材失败")
		return
	}

	response.Success(c, asset)
}
