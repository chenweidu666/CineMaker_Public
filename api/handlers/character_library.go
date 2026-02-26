package handlers

import (
	"errors"
	"strconv"

	services2 "github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/domain"
	"github.com/cinemaker/backend/infrastructure/storage"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CharacterLibraryHandler struct {
	db             *gorm.DB
	libraryService *services2.CharacterLibraryService
	imageService   *services2.ImageGenerationService
	log            *logger.Logger
}

func NewCharacterLibraryHandler(db *gorm.DB, cfg *config.Config, log *logger.Logger, transferService *services2.ResourceTransferService, localStorage storage.Storage) *CharacterLibraryHandler {
	return &CharacterLibraryHandler{
		db:             db,
		libraryService: services2.NewCharacterLibraryService(db, log, cfg, localStorage),
		imageService:   services2.NewImageGenerationService(db, cfg, transferService, localStorage, log),
		log:            log,
	}
}

// ListLibraryItems 获取角色库列表
func (h *CharacterLibraryHandler) ListLibraryItems(c *gin.Context) {

	var query services2.CharacterLibraryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "请求参数无效")
		return
	}

	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}

	items, total, err := h.libraryService.ListLibraryItems(&query)
	if err != nil {
		h.log.Errorw("Failed to list library items", "error", err)
		response.InternalError(c, "获取角色库失败")
		return
	}

	response.SuccessWithPagination(c, items, total, query.Page, query.PageSize)
}

// CreateLibraryItem 添加到角色库
func (h *CharacterLibraryHandler) CreateLibraryItem(c *gin.Context) {

	var req services2.CreateLibraryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效")
		return
	}

	item, err := h.libraryService.CreateLibraryItem(&req)
	if err != nil {
		h.log.Errorw("Failed to create library item", "error", err)
		response.InternalError(c, "添加到角色库失败")
		return
	}

	response.Created(c, item)
}

// GetLibraryItem 获取角色库项详情
func (h *CharacterLibraryHandler) GetLibraryItem(c *gin.Context) {

	itemID := c.Param("id")

	item, err := h.libraryService.GetLibraryItem(itemID)
	if err != nil {
		if errors.Is(err, domain.ErrLibraryItemNotFound) {
			response.NotFound(c, "角色库项不存在")
			return
		}
		h.log.Errorw("Failed to get library item", "error", err)
		response.InternalError(c, "获取失败")
		return
	}

	response.Success(c, item)
}

// DeleteLibraryItem 删除角色库项
func (h *CharacterLibraryHandler) DeleteLibraryItem(c *gin.Context) {

	itemID := c.Param("id")

	if err := h.libraryService.DeleteLibraryItem(itemID); err != nil {
		if errors.Is(err, domain.ErrLibraryItemNotFound) {
			response.NotFound(c, "角色库项不存在")
			return
		}
		h.log.Errorw("Failed to delete library item", "error", err)
		response.InternalError(c, "删除失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// UploadCharacterImage 上传角色图片
func (h *CharacterLibraryHandler) UploadCharacterImage(c *gin.Context) {

	characterID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyCharacterTeam(h.db, characterID, teamID); err != nil {
		response.Forbidden(c, "无权访问该角色")
		return
	}

	// TODO: 处理文件上传
	// 这里需要实现文件上传逻辑，保存到OSS或本地
	// 暂时使用简单的实现
	var req struct {
		ImageURL string `json:"image_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效")
		return
	}

	if err := h.libraryService.UploadCharacterImage(characterID, req.ImageURL); err != nil {
		if errors.Is(err, domain.ErrCharacterNotFound) {
			response.NotFound(c, "角色不存在")
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			response.Forbidden(c, "无权限")
			return
		}
		h.log.Errorw("Failed to upload character image", "error", err)
		response.InternalError(c, "上传失败")
		return
	}

	response.Success(c, gin.H{"message": "上传成功"})
}

// ApplyLibraryItemToCharacter 从角色库应用形象
func (h *CharacterLibraryHandler) ApplyLibraryItemToCharacter(c *gin.Context) {

	characterID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyCharacterTeam(h.db, characterID, teamID); err != nil {
		response.Forbidden(c, "无权访问该角色")
		return
	}

	var req struct {
		LibraryItemID string `json:"library_item_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效")
		return
	}

	if err := h.libraryService.ApplyLibraryItemToCharacter(characterID, req.LibraryItemID); err != nil {
		if errors.Is(err, domain.ErrLibraryItemNotFound) {
			response.NotFound(c, "角色库项不存在")
			return
		}
		if errors.Is(err, domain.ErrCharacterNotFound) {
			response.NotFound(c, "角色不存在")
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			response.Forbidden(c, "无权限")
			return
		}
		h.log.Errorw("Failed to apply library item", "error", err)
		response.InternalError(c, "应用失败")
		return
	}

	response.Success(c, gin.H{"message": "应用成功"})
}

// AddCharacterToLibrary 将角色添加到角色库
func (h *CharacterLibraryHandler) AddCharacterToLibrary(c *gin.Context) {

	characterID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyCharacterTeam(h.db, characterID, teamID); err != nil {
		response.Forbidden(c, "无权访问该角色")
		return
	}

	var req struct {
		Category *string `json:"category"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空body
		req.Category = nil
	}

	item, err := h.libraryService.AddCharacterToLibrary(characterID, req.Category)
	if err != nil {
		if errors.Is(err, domain.ErrCharacterNotFound) {
			response.NotFound(c, "角色不存在")
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			response.Forbidden(c, "无权限")
			return
		}
		if errors.Is(err, domain.ErrCharacterNoImage) {
			response.BadRequest(c, "角色还没有形象图片")
			return
		}
		h.log.Errorw("Failed to add character to library", "error", err)
		response.InternalError(c, "添加失败")
		return
	}

	response.Created(c, item)
}

// UpdateCharacter 更新角色信息
func (h *CharacterLibraryHandler) UpdateCharacter(c *gin.Context) {

	characterID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyCharacterTeam(h.db, characterID, teamID); err != nil {
		response.Forbidden(c, "无权访问该角色")
		return
	}

	var req services2.UpdateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效")
		return
	}

	if err := h.libraryService.UpdateCharacter(characterID, &req); err != nil {
		if errors.Is(err, domain.ErrCharacterNotFound) {
			response.NotFound(c, "角色不存在")
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			response.Forbidden(c, "无权限")
			return
		}
		h.log.Errorw("Failed to update character", "error", err)
		response.InternalError(c, "更新失败")
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}

// DeleteCharacter 删除单个角色
func (h *CharacterLibraryHandler) DeleteCharacter(c *gin.Context) {

	characterIDStr := c.Param("id")
	characterID, err := strconv.ParseUint(characterIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyCharacterTeam(h.db, characterIDStr, teamID); err != nil {
		response.Forbidden(c, "无权访问该角色")
		return
	}

	if err := h.libraryService.DeleteCharacter(uint(characterID)); err != nil {
		h.log.Errorw("Failed to delete character", "error", err, "id", characterID)
		if errors.Is(err, domain.ErrCharacterNotFound) {
			response.NotFound(c, "角色不存在")
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			response.Forbidden(c, "无权删除此角色")
			return
		}
		response.InternalError(c, "删除失败")
		return
	}

	response.Success(c, gin.H{"message": "角色已删除"})
}

// ExtractCharacters 从剧本提取角色
func (h *CharacterLibraryHandler) ExtractCharacters(c *gin.Context) {
	episodeIDStr := c.Param("episode_id")
	episodeID, err := strconv.ParseUint(episodeIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的章节ID")
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

	taskID, err := h.libraryService.ExtractCharactersFromScript(uint(episodeID))
	if err != nil {
		h.log.Errorw("Failed to extract characters", "error", err)
		response.InternalError(c, "角色提取失败")
		return
	}

	response.Success(c, gin.H{"task_id": taskID, "message": "角色提取任务已提交"})
}

// GenerateCharacterImage 生成角色图片
func (h *CharacterLibraryHandler) GenerateCharacterImage(c *gin.Context) {
	characterID := c.Param("id")

	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}
	if err := auth.VerifyCharacterTeam(h.db, characterID, teamID); err != nil {
		response.Forbidden(c, "无权访问该角色")
		return
	}

	var req struct {
		Model           string   `json:"model"`
		Style           string   `json:"style"`
		Prompt          string   `json:"prompt"`
		Size            string   `json:"size"`
		ReferenceImages []string `json:"reference_images"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Model = ""
		req.Style = ""
		req.Prompt = ""
		req.Size = ""
		req.ReferenceImages = nil
	}

	imageGen, prompt, err := h.libraryService.GenerateCharacterImage(characterID, h.imageService, req.Model, req.Style, req.Prompt, req.Size, req.ReferenceImages)
	if err != nil {
		h.log.Errorw("Failed to generate character image", "error", err, "character_id", characterID)
		response.InternalError(c, "生成角色图片失败")
		return
	}

	response.Success(c, gin.H{
		"image_url": imageGen.ImageURL,
		"prompt":    prompt,
	})
}
