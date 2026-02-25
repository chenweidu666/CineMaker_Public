package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cinemaker/backend/domain"
	models "github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/infrastructure/storage"
	"github.com/cinemaker/backend/pkg/ai"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CharacterLibraryService struct {
	db           *gorm.DB
	log          *logger.Logger
	config       *config.Config
	aiService    *AIService
	taskService  *TaskService
	promptI18n   *PromptI18n
	localStorage storage.Storage
}

func NewCharacterLibraryService(db *gorm.DB, log *logger.Logger, cfg *config.Config, localStorage storage.Storage) *CharacterLibraryService {
	return &CharacterLibraryService{
		db:           db,
		log:          log,
		config:       cfg,
		aiService:    NewAIService(db, log),
		taskService:  NewTaskService(db, log),
		promptI18n:   NewPromptI18n(cfg),
		localStorage: localStorage,
	}
}

type CreateLibraryItemRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Category    *string `json:"category"`
	ImageURL    string  `json:"image_url" binding:"required"`
	LocalPath   *string `json:"local_path"`
	Description *string `json:"description"`
	Tags        *string `json:"tags"`
	SourceType  string  `json:"source_type"`
}

type CharacterLibraryQuery struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
	Category   string `form:"category"`
	SourceType string `form:"source_type"`
	Keyword    string `form:"keyword"`
}

// ListLibraryItems 获取用户角色库列表
func (s *CharacterLibraryService) ListLibraryItems(query *CharacterLibraryQuery) ([]models.CharacterLibrary, int64, error) {
	var items []models.CharacterLibrary
	var total int64

	db := s.db.Model(&models.CharacterLibrary{})

	// 筛选条件
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}

	if query.SourceType != "" {
		db = db.Where("source_type = ?", query.SourceType)
	}

	if query.Keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		s.log.Errorw("Failed to count character library", "error", err)
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	err := db.Order("created_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&items).Error

	if err != nil {
		s.log.Errorw("Failed to list character library", "error", err)
		return nil, 0, err
	}

	return items, total, nil
}

// CreateLibraryItem 添加到角色库
func (s *CharacterLibraryService) CreateLibraryItem(req *CreateLibraryItemRequest) (*models.CharacterLibrary, error) {
	sourceType := req.SourceType
	if sourceType == "" {
		sourceType = "generated"
	}

	item := &models.CharacterLibrary{
		Name:        req.Name,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
		LocalPath:   req.LocalPath,
		Description: req.Description,
		Tags:        req.Tags,
		SourceType:  sourceType,
	}

	if err := s.db.Create(item).Error; err != nil {
		s.log.Errorw("Failed to create library item", "error", err)
		return nil, err
	}

	s.log.Infow("Library item created", "item_id", item.ID)
	return item, nil
}

// GetLibraryItem 获取角色库项
func (s *CharacterLibraryService) GetLibraryItem(itemID string) (*models.CharacterLibrary, error) {
	var item models.CharacterLibrary
	err := s.db.Where("id = ? ", itemID).First(&item).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLibraryItemNotFound
		}
		s.log.Errorw("Failed to get library item", "error", err)
		return nil, err
	}

	return &item, nil
}

// DeleteLibraryItem 删除角色库项
func (s *CharacterLibraryService) DeleteLibraryItem(itemID string) error {
	result := s.db.Where("id = ? ", itemID).Delete(&models.CharacterLibrary{})

	if result.Error != nil {
		s.log.Errorw("Failed to delete library item", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrLibraryItemNotFound
	}

	s.log.Infow("Library item deleted", "item_id", itemID)
	return nil
}

// ApplyLibraryItemToCharacter 将角色库形象应用到角色
func (s *CharacterLibraryService) ApplyLibraryItemToCharacter(characterID string, libraryItemID string) error {
	// 验证角色库项存在且属于该用户
	var libraryItem models.CharacterLibrary
	if err := s.db.Where("id = ? ", libraryItemID).First(&libraryItem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrLibraryItemNotFound
		}
		return err
	}

	// 查找角色
	var character models.Character
	if err := s.db.Where("id = ?", characterID).First(&character).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrCharacterNotFound
		}
		return err
	}

	// 查询Drama验证权限
	var drama models.Drama
	if err := s.db.Where("id = ? ", character.DramaID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrUnauthorized
		}
		return err
	}

	// 更新角色的 local_path 和 image_url
	updates := map[string]interface{}{}
	if libraryItem.LocalPath != nil && *libraryItem.LocalPath != "" {
		updates["local_path"] = libraryItem.LocalPath
	}
	if libraryItem.ImageURL != "" {
		updates["image_url"] = libraryItem.ImageURL
	}
	if len(updates) > 0 {
		if err := s.db.Model(&character).Updates(updates).Error; err != nil {
			s.log.Errorw("Failed to update character image", "error", err)
			return err
		}
	}

	s.log.Infow("Library item applied to character", "character_id", characterID, "library_item_id", libraryItemID)
	return nil
}

// UploadCharacterImage 上传角色图片
func (s *CharacterLibraryService) UploadCharacterImage(characterID string, imageURL string) error {
	// 查找角色
	var character models.Character
	if err := s.db.Where("id = ?", characterID).First(&character).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrCharacterNotFound
		}
		return err
	}

	// 查询Drama验证权限
	var drama models.Drama
	if err := s.db.Where("id = ? ", character.DramaID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrUnauthorized
		}
		return err
	}

	// 更新图片URL
	if err := s.db.Model(&character).Update("image_url", imageURL).Error; err != nil {
		s.log.Errorw("Failed to update character image", "error", err)
		return err
	}

	s.log.Infow("Character image uploaded", "character_id", characterID)
	return nil
}

// AddCharacterToLibrary 将角色添加到角色库
func (s *CharacterLibraryService) AddCharacterToLibrary(characterID string, category *string) (*models.CharacterLibrary, error) {
	// 查找角色
	var character models.Character
	if err := s.db.Where("id = ?", characterID).First(&character).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCharacterNotFound
		}
		return nil, err
	}

	// 查询Drama验证权限
	var drama models.Drama
	if err := s.db.Where("id = ? ", character.DramaID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	// 检查是否有图片
	if character.ImageURL == nil || *character.ImageURL == "" {
		return nil, fmt.Errorf("角色还没有形象图片")
	}

	// 创建角色库项
	charLibrary := &models.CharacterLibrary{
		Name:        character.Name,
		ImageURL:    *character.ImageURL,
		LocalPath:   character.LocalPath,
		Description: character.Description,
		SourceType:  "character",
	}

	if err := s.db.Create(charLibrary).Error; err != nil {
		s.log.Errorw("Failed to add character to library", "error", err)
		return nil, err
	}

	s.log.Infow("Character added to library", "character_id", characterID, "library_item_id", charLibrary.ID)
	return charLibrary, nil
}

// DeleteCharacter 删除单个角色
func (s *CharacterLibraryService) DeleteCharacter(characterID uint) error {
	// 查找角色
	var character models.Character
	if err := s.db.Where("id = ?", characterID).First(&character).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrCharacterNotFound
		}
		return err
	}

	// 验证权限：检查角色所属的drama是否属于当前用户
	var drama models.Drama
	if err := s.db.Where("id = ? ", character.DramaID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrUnauthorized
		}
		return err
	}

	// 删除角色
	if err := s.db.Delete(&character).Error; err != nil {
		s.log.Errorw("Failed to delete character", "error", err, "id", characterID)
		return err
	}

	s.log.Infow("Character deleted", "id", characterID)
	return nil
}

// GenerateCharacterImage AI生成角色形象
func (s *CharacterLibraryService) GenerateCharacterImage(characterID string, imageService *ImageGenerationService, modelName string, style string, customPrompt string, size string, referenceImages []string) (*models.ImageGeneration, string, error) {
	// 查找角色
	var character models.Character
	if err := s.db.Where("id = ?", characterID).First(&character).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", domain.ErrCharacterNotFound
		}
		return nil, "", err
	}

	// 查询Drama验证权限
	var drama models.Drama
	if err := s.db.Where("id = ? ", character.DramaID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", domain.ErrUnauthorized
		}
		return nil, "", err
	}

	prompt := ""

	// 确定实际使用的风格：优先用传入的style，否则从drama获取
	effectiveStyle := style
	if effectiveStyle == "" {
		effectiveStyle = drama.Style
	}
	if effectiveStyle == "" {
		effectiveStyle = "realistic"
	}

	// 获取风格的中文名称，用于嵌入prompt
	styleNameMap := map[string]string{
		"realistic": "超写实摄影风格",
		"comic":     "漫画风格",
	}
	styleName := styleNameMap[effectiveStyle]
	if styleName == "" {
		styleName = effectiveStyle + "风格"
	}

	if customPrompt != "" {
		prompt = customPrompt
		if !strings.Contains(prompt, "风格") {
			prompt = styleName + "，" + prompt
		}
	} else {
		if character.Appearance != nil && *character.Appearance != "" {
			prompt = *character.Appearance
		} else if character.Description != nil && *character.Description != "" {
			prompt = *character.Description
		} else {
			prompt = character.Name
		}

		prompt = fmt.Sprintf("4:3横向构图，%s，角色三视图设定图。"+
			"纯白色背景，无场景无装饰无文字。"+
			"一张图内从左到右并排展示同一角色的正面、3/4侧面、背面三个全身站立视角。"+
			"每个视角均为头顶至鞋底的完整全身像，身材修长高挑，头身比1:7.5。"+
			"三视角外貌服饰配色完全统一，面部五官清晰可辨。"+
			"角色描述：%s", styleName, prompt)
	}

	// 注入身材比例与全身构图要求（正面指令，避免负面描述）
	if !strings.Contains(prompt, "头身比") {
		prompt += " 身材修长高挑，头身比1:7.5，腿部修长占身高一半以上，每个视角均为头顶至鞋底完整全身站立像。"
	}

	// 三视图用4:3（2304x1728）：3人并排每人~768x1728，高宽比1:2.25，兼顾宽度和全身高度
	imageSize := size
	if imageSize == "" {
		imageSize = "2304x1728"
	}

	dramaIDStr := fmt.Sprintf("%d", character.DramaID)
	imageType := "character"
	imageStyle := effectiveStyle
	req := &GenerateImageRequest{
		DramaID:         dramaIDStr,
		CharacterID:     &character.ID,
		ImageType:       imageType,
		Prompt:          prompt,
		Provider:        "openai",
		Model:           modelName,
		Size:            imageSize,
		Quality:         "standard",
		Style:           &imageStyle,
		ReferenceImages: referenceImages,
	}

	imageGen, err := imageService.GenerateImage(req)
	if err != nil {
		s.log.Errorw("Failed to generate character image", "error", err)
		return nil, "", fmt.Errorf("图片生成失败: %w", err)
	}

	// 将最终 prompt 回写到角色，前端编辑框可以直接看到完整内容
	s.db.Model(&models.Character{}).Where("id = ?", character.ID).Update("prompt", prompt)

	// 异步处理：在后台监听图片生成完成，然后更新角色image_url
	go s.waitAndUpdateCharacterImage(character.ID, imageGen.ID)

	// 立即返回ImageGeneration对象和提示词，让前端可以轮询状态
	s.log.Infow("Character image generation started", "character_id", characterID, "image_gen_id", imageGen.ID)
	return imageGen, prompt, nil
}

// waitAndUpdateCharacterImage 后台异步等待图片生成完成并更新角色image_url
func (s *CharacterLibraryService) waitAndUpdateCharacterImage(characterID uint, imageGenID uint) {
	maxAttempts := 60
	pollInterval := 5 * time.Second

	for i := 0; i < maxAttempts; i++ {
		time.Sleep(pollInterval)

		// 查询图片生成状态
		var imageGen models.ImageGeneration
		if err := s.db.First(&imageGen, imageGenID).Error; err != nil {
			s.log.Errorw("Failed to query image generation status", "error", err, "image_gen_id", imageGenID)
			continue
		}

		// 检查是否完成
		if imageGen.Status == models.ImageStatusCompleted && imageGen.ImageURL != nil && *imageGen.ImageURL != "" {
			// 更新角色的image_url
			if err := s.db.Model(&models.Character{}).Where("id = ?", characterID).Update("image_url", *imageGen.ImageURL).Error; err != nil {
				s.log.Errorw("Failed to update character image_url", "error", err, "character_id", characterID)
				return
			}
			s.log.Infow("Character image updated successfully", "character_id", characterID, "image_url", *imageGen.ImageURL)
			return
		}

		// 检查是否失败
		if imageGen.Status == models.ImageStatusFailed {
			s.log.Errorw("Character image generation failed", "character_id", characterID, "image_gen_id", imageGenID, "error", imageGen.ErrorMsg)
			return
		}
	}

	s.log.Warnw("Character image generation timeout", "character_id", characterID, "image_gen_id", imageGenID)
}

type UpdateCharacterRequest struct {
	Name             *string `json:"name"`
	Role             *string `json:"role"`
	Appearance       *string `json:"appearance"`
	Personality      *string `json:"personality"`
	Description      *string `json:"description"`
	Prompt           *string `json:"prompt"`
	ImageURL         *string `json:"image_url"`
	LocalPath        *string `json:"local_path"`
	ReferenceImages  []any   `json:"reference_images"`
	ImageOrientation *string `json:"image_orientation"`
}

// UpdateCharacter 更新角色信息
func (s *CharacterLibraryService) UpdateCharacter(characterID string, req *UpdateCharacterRequest) error {
	// 查找角色
	var character models.Character
	if err := s.db.Where("id = ?", characterID).First(&character).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrCharacterNotFound
		}
		return err
	}

	// 验证权限：查询角色所属的drama是否属于该用户
	var drama models.Drama
	if err := s.db.Where("id = ? ", character.DramaID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrUnauthorized
		}
		return err
	}

	// 构建更新数据
	updates := make(map[string]interface{})

	if req.Name != nil && *req.Name != "" {
		updates["name"] = *req.Name
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.Appearance != nil {
		updates["appearance"] = *req.Appearance
	}
	if req.Personality != nil {
		updates["personality"] = *req.Personality
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Prompt != nil {
		updates["prompt"] = *req.Prompt
	}
	if req.ImageURL != nil {
		updates["image_url"] = *req.ImageURL
	}
	if req.LocalPath != nil {
		updates["local_path"] = *req.LocalPath
	}
	if req.ReferenceImages != nil {
		jsonData, err := json.Marshal(req.ReferenceImages)
		if err == nil {
			updates["reference_images"] = datatypes.JSON(jsonData)
		}
	}
	if req.ImageOrientation != nil {
		updates["image_orientation"] = *req.ImageOrientation
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	if req.ImageURL != nil || req.LocalPath != nil {
		if character.LocalPath != nil && *character.LocalPath != "" && s.localStorage != nil {
			if err := s.localStorage.Delete(*character.LocalPath); err != nil {
				s.log.Warnw("Failed to delete old character image", "error", err, "path", *character.LocalPath)
			}
		}
	}

	if err := s.db.Model(&character).Updates(updates).Error; err != nil {
		s.log.Errorw("Failed to update character", "error", err, "character_id", characterID)
		return err
	}

	s.log.Infow("Character updated", "character_id", characterID, "updates", updates)
	return nil
}

// BatchGenerateCharacterImages 批量生成角色图片（并发执行）
func (s *CharacterLibraryService) BatchGenerateCharacterImages(characterIDs []string, imageService *ImageGenerationService, modelName string) {
	s.log.Infow("Starting batch character image generation",
		"count", len(characterIDs),
		"model", modelName)

	// 使用 goroutine 并发生成所有角色图片
	for _, characterID := range characterIDs {
		// 为每个角色启动单独的 goroutine
		go func(charID string) {
			imageGen, _, err := s.GenerateCharacterImage(charID, imageService, modelName, "", "", "", nil)
			if err != nil {
				s.log.Errorw("Failed to generate character image in batch",
					"character_id", charID,
					"error", err)
				return
			}

			s.log.Infow("Character image generated in batch",
				"character_id", charID,
				"image_gen_id", imageGen.ID)
		}(characterID)
	}

	s.log.Infow("Batch character image generation tasks submitted",
		"total", len(characterIDs))
}

// ExtractCharactersFromScript 从分集剧本中提取角色
func (s *CharacterLibraryService) ExtractCharactersFromScript(episodeID uint) (string, error) {
	var episode models.Episode
	if err := s.db.First(&episode, episodeID).Error; err != nil {
		return "", domain.ErrEpisodeNotFound
	}

	if episode.ScriptContent == nil || *episode.ScriptContent == "" {
		return "", fmt.Errorf("剧本内容为空")
	}

	task, err := s.taskService.CreateTask("character_extraction", fmt.Sprintf("%d", episode.DramaID))
	if err != nil {
		return "", fmt.Errorf("创建任务失败: %w", err)
	}

	go s.processCharacterExtraction(task.ID, episode)

	return task.ID, nil
}

func (s *CharacterLibraryService) processCharacterExtraction(taskID string, episode models.Episode) {
	s.taskService.UpdateTaskStatus(taskID, "processing", 0, "正在分析剧本...")

	script := ""
	if episode.ScriptContent != nil {
		script = *episode.ScriptContent
	}

	// 获取 drama 的 style 信息
	var drama models.Drama
	if err := s.db.First(&drama, episode.DramaID).Error; err != nil {
		s.log.Warnw("Failed to load drama", "error", err, "drama_id", episode.DramaID)
	}

	prompt := s.promptI18n.GetCharacterExtractionPrompt(drama.Style)
	userPrompt := fmt.Sprintf("【剧本内容】\n%s", script)

	response, err := s.aiService.GenerateText(userPrompt, prompt, ai.WithMaxTokens(3000))
	if err != nil {
		s.taskService.UpdateTaskError(taskID, err)
		return
	}

	s.taskService.UpdateTaskStatus(taskID, "processing", 50, "正在整理角色数据...")

	var extractedCharacters []struct {
		Name        string `json:"name"`
		Role        string `json:"role"`
		Appearance  string `json:"appearance"`
		Personality string `json:"personality"`
		Description string `json:"description"`
	}

	if err := utils.SafeParseAIJSON(response, &extractedCharacters); err != nil {
		s.log.Errorw("Failed to parse AI response for characters", "error", err, "response", response)
		s.taskService.UpdateTaskError(taskID, fmt.Errorf("解析AI响应失败"))
		return
	}

	var savedCharacters []models.Character
	for _, charData := range extractedCharacters {
		// 检查是否已存在同名角色
		var existingCharacter models.Character
		err := s.db.Where("drama_id = ? AND name = ?", episode.DramaID, charData.Name).First(&existingCharacter).Error

		if err == nil {
			// 如果存在，只关联，不更新（或者可以选更新，这里暂不更新）
			if err := s.db.Model(&episode).Association("Characters").Append(&existingCharacter); err != nil {
				s.log.Warnw("Failed to associate existing character", "error", err)
			}
			savedCharacters = append(savedCharacters, existingCharacter)
		} else {
			// 创建新角色
			newCharacter := models.Character{
				DramaID:     episode.DramaID,
				Name:        charData.Name,
				Role:        &charData.Role,
				Appearance:  &charData.Appearance,
				Personality: &charData.Personality,
				Description: &charData.Description,
			}
			if err := s.db.Create(&newCharacter).Error; err != nil {
				s.log.Errorw("Failed to create extracted character", "error", err)
				continue
			}

			// 关联到分集
			if err := s.db.Model(&episode).Association("Characters").Append(&newCharacter); err != nil {
				s.log.Warnw("Failed to associate new character", "error", err)
			}
			savedCharacters = append(savedCharacters, newCharacter)
		}
	}

	s.taskService.UpdateTaskResult(taskID, map[string]interface{}{
		"characters": savedCharacters,
		"count":      len(savedCharacters),
	})
}
