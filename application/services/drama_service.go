package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cinemaker/backend/domain"
	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DramaService struct {
	db      *gorm.DB
	log     *logger.Logger
	baseURL string
}

func NewDramaService(db *gorm.DB, cfg *config.Config, log *logger.Logger) *DramaService {
	return &DramaService{
		db:      db,
		log:     log,
		baseURL: cfg.Storage.BaseURL,
	}
}

type CreateDramaRequest struct {
	Title         string `json:"title" binding:"required,min=1,max=100"`
	Description   string `json:"description"`
	Genre         string `json:"genre"`
	Style         string `json:"style"`
	Tags          string `json:"tags"`
	GenerateAudio *bool  `json:"generate_audio"`
	TeamID        *uint  `json:"-"` // set by handler from auth context
}

type UpdateDramaRequest struct {
	Title         string `json:"title" binding:"omitempty,min=1,max=100"`
	Description   string `json:"description"`
	Genre         string `json:"genre"`
	Style         string `json:"style"`
	Tags          string `json:"tags"`
	Status        string `json:"status" binding:"omitempty,oneof=draft planning production completed archived"`
	GenerateAudio *bool  `json:"generate_audio"`
}

type DramaListQuery struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Status   string `form:"status"`
	Genre    string `form:"genre"`
	Keyword  string `form:"keyword"`
	TeamID   *uint  `form:"-"` // set by handler from auth context
}

func (s *DramaService) CreateDrama(req *CreateDramaRequest) (*models.Drama, error) {
	drama := &models.Drama{
		Title:  req.Title,
		TeamID: req.TeamID,
		Status: models.DramaStatusDraft,
		Style:  models.DefaultStyle,
	}

	if req.Description != "" {
		drama.Description = &req.Description
	}
	if req.Genre != "" {
		drama.Genre = &req.Genre
	}
	if req.Style != "" {
		drama.Style = req.Style
	}
	if req.GenerateAudio != nil {
		drama.GenerateAudio = req.GenerateAudio
	}

	if err := s.db.Create(drama).Error; err != nil {
		s.log.Errorw("Failed to create drama", "error", err)
		return nil, err
	}

	s.log.Infow("Drama created", "drama_id", drama.ID)
	return drama, nil
}

func (s *DramaService) GetDrama(dramaID string) (*models.Drama, error) {
	var drama models.Drama
	err := s.db.Where("id = ? ", dramaID).
		Preload("Characters", "parent_id IS NULL").
		Preload("Characters.Children").
		Preload("Scenes").
		Preload("Props").
		Preload("Episodes.Characters").
		Preload("Episodes.Scenes").     // 加载每个章节关联的场景
		Preload("Episodes.Props").      // 加载每个章节关联的道具
		Preload("Episodes.Storyboards", func(db *gorm.DB) *gorm.DB {
			return db.Order("storyboards.storyboard_number ASC")
		}).
		Preload("Episodes.Storyboards.Characters"). // 加载分镜关联的角色
		Preload("Episodes.Storyboards.Props").      // 加载分镜关联的道具
		First(&drama).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDramaNotFound
		}
		s.log.Errorw("Failed to get drama", "error", err)
		return nil, err
	}

	// 统计每个剧集的时长（基于场景时长之和）
	for i := range drama.Episodes {
		totalDuration := 0
		for _, scene := range drama.Episodes[i].Storyboards {
			totalDuration += scene.Duration
		}
		// 更新剧集时长（秒转分钟，向上取整）
		durationMinutes := (totalDuration + 59) / 60
		drama.Episodes[i].Duration = durationMinutes

		// 如果数据库中的时长与计算的不一致，更新数据库
		if drama.Episodes[i].Duration != durationMinutes {
			s.db.Model(&models.Episode{}).Where("id = ?", drama.Episodes[i].ID).Update("duration", durationMinutes)
		}

		// 查询角色的图片生成状态
		for j := range drama.Episodes[i].Characters {
			var imageGen models.ImageGeneration
			// 查询进行中或失败的任务状态
			err := s.db.Where("character_id = ? AND (status = ? OR status = ?)",
				drama.Episodes[i].Characters[j].ID, "pending", "processing").
				Order("created_at DESC").
				First(&imageGen).Error

			if err == nil {
				// 找到生成中的记录，设置状态
				statusStr := string(imageGen.Status)
				drama.Episodes[i].Characters[j].ImageGenerationStatus = &statusStr
				if imageGen.ErrorMsg != nil {
					drama.Episodes[i].Characters[j].ImageGenerationError = imageGen.ErrorMsg
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				// 检查是否有失败的记录
				err := s.db.Where("character_id = ? AND status = ?",
					drama.Episodes[i].Characters[j].ID, "failed").
					Order("created_at DESC").
					First(&imageGen).Error

				if err == nil {
					statusStr := string(imageGen.Status)
					drama.Episodes[i].Characters[j].ImageGenerationStatus = &statusStr
					if imageGen.ErrorMsg != nil {
						drama.Episodes[i].Characters[j].ImageGenerationError = imageGen.ErrorMsg
					}
				}
			}
		}

		// 查询场景的图片生成状态
		for j := range drama.Episodes[i].Scenes {
			var imageGen models.ImageGeneration
			// 查询进行中或失败的任务状态
			err := s.db.Where("scene_id = ? AND (status = ? OR status = ?)",
				drama.Episodes[i].Scenes[j].ID, "pending", "processing").
				Order("created_at DESC").
				First(&imageGen).Error

			if err == nil {
				// 找到生成中的记录，设置状态
				statusStr := string(imageGen.Status)
				drama.Episodes[i].Scenes[j].ImageGenerationStatus = &statusStr
				if imageGen.ErrorMsg != nil {
					drama.Episodes[i].Scenes[j].ImageGenerationError = imageGen.ErrorMsg
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				// 检查是否有失败的记录
				err := s.db.Where("scene_id = ? AND status = ?",
					drama.Episodes[i].Scenes[j].ID, "failed").
					Order("created_at DESC").
					First(&imageGen).Error

				if err == nil {
					statusStr := string(imageGen.Status)
					drama.Episodes[i].Scenes[j].ImageGenerationStatus = &statusStr
					if imageGen.ErrorMsg != nil {
						drama.Episodes[i].Scenes[j].ImageGenerationError = imageGen.ErrorMsg
					}
				}
			}
		}
	}

	// 整合所有剧集的场景到Drama级别的Scenes字段
	sceneMap := make(map[uint]*models.Scene) // 用于去重
	for i := range drama.Episodes {
		for j := range drama.Episodes[i].Scenes {
			scene := &drama.Episodes[i].Scenes[j]
			sceneMap[scene.ID] = scene
		}
	}

	// 将整合的场景添加到drama.Scenes
	drama.Scenes = make([]models.Scene, 0, len(sceneMap))
	for _, scene := range sceneMap {
		drama.Scenes = append(drama.Scenes, *scene)
	}

	// 为所有场景的 local_path 添加 base_url 前缀
	// s.addBaseURLToScenes(&drama)

	return &drama, nil
}

func (s *DramaService) ListDramas(query *DramaListQuery) ([]models.Drama, int64, error) {
	var dramas []models.Drama
	var total int64

	db := s.db.Model(&models.Drama{})

	if query.TeamID != nil {
		db = db.Where("team_id = ?", *query.TeamID)
	}

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if query.Genre != "" {
		db = db.Where("genre = ?", query.Genre)
	}

	if query.Keyword != "" {
		db = db.Where("title LIKE ? OR description LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		s.log.Errorw("Failed to count dramas", "error", err)
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Order("updated_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Preload("Episodes.Storyboards", func(db *gorm.DB) *gorm.DB {
			return db.Order("storyboards.storyboard_number ASC")
		}).
		Find(&dramas).Error

	if err != nil {
		s.log.Errorw("Failed to list dramas", "error", err)
		return nil, 0, err
	}

	// 统计每个剧本的每个剧集的时长（基于场景时长之和）
	for i := range dramas {
		for j := range dramas[i].Episodes {
			totalDuration := 0
			for _, scene := range dramas[i].Episodes[j].Storyboards {
				totalDuration += scene.Duration
			}
			// 更新剧集时长（秒转分钟，向上取整）
			durationMinutes := (totalDuration + 59) / 60
			dramas[i].Episodes[j].Duration = durationMinutes
		}
	}

	return dramas, total, nil
}

func (s *DramaService) UpdateDrama(dramaID string, req *UpdateDramaRequest) (*models.Drama, error) {
	var drama models.Drama
	if err := s.db.Where("id = ? ", dramaID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDramaNotFound
		}
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Genre != "" {
		updates["genre"] = req.Genre
	}
	if req.Style != "" {
		updates["style"] = req.Style
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.GenerateAudio != nil {
		updates["generate_audio"] = *req.GenerateAudio
	}

	updates["updated_at"] = time.Now()

	if err := s.db.Model(&drama).Updates(updates).Error; err != nil {
		s.log.Errorw("Failed to update drama", "error", err)
		return nil, err
	}

	if err := s.db.Where("id = ?", dramaID).First(&drama).Error; err != nil {
		s.log.Errorw("Failed to reload drama after update", "error", err)
		return nil, err
	}

	s.log.Infow("Drama updated", "drama_id", dramaID)
	return &drama, nil
}

func (s *DramaService) DeleteDrama(dramaID string) error {
	result := s.db.Where("id = ? ", dramaID).Delete(&models.Drama{})

	if result.Error != nil {
		s.log.Errorw("Failed to delete drama", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrDramaNotFound
	}

	s.log.Infow("Drama deleted", "drama_id", dramaID)
	return nil
}

func (s *DramaService) GetDramaStats() (map[string]interface{}, error) {
	var total int64
	var byStatus []struct {
		Status string
		Count  int64
	}

	if err := s.db.Model(&models.Drama{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.Drama{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&byStatus).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total":     total,
		"by_status": byStatus,
	}

	return stats, nil
}

func (s *DramaService) GetDramaStatsByTeam(teamID uint) (map[string]interface{}, error) {
	var total int64
	var byStatus []struct {
		Status string
		Count  int64
	}

	base := s.db.Model(&models.Drama{}).Where("team_id = ?", teamID)

	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.Drama{}).Where("team_id = ?", teamID).
		Select("status, count(*) as count").
		Group("status").
		Scan(&byStatus).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":     total,
		"by_status": byStatus,
	}, nil
}

type SaveOutlineRequest struct {
	Title   string   `json:"title" binding:"required"`
	Summary string   `json:"summary" binding:"required"`
	Genre   string   `json:"genre"`
	Tags    []string `json:"tags"`
}

type CreateCharacterRequest struct {
	DramaID          uint   `json:"drama_id"`
	EpisodeID        *uint  `json:"episode_id"`
	ParentID         *uint  `json:"parent_id"`
	Name             string `json:"name"`
	OutfitName       string `json:"outfit_name"`
	Role             string `json:"role"`
	Gender           string `json:"gender"`
	AgeDescription   string `json:"age_description"`
	Description      string `json:"description"`
	Appearance       string `json:"appearance"`
	Personality      string `json:"personality"`
	VoiceStyle       string `json:"voice_style"`
	Prompt           string `json:"prompt"`
	ImageURL         string `json:"image_url"`
	LocalPath        string `json:"local_path"`
	ReferenceImages  []any  `json:"reference_images"`
	ImageOrientation string `json:"image_orientation"`
}

func (s *DramaService) CreateCharacter(req *CreateCharacterRequest) (*models.Character, error) {
	character := &models.Character{
		DramaID:     req.DramaID,
		ParentID:    req.ParentID,
		Name:        req.Name,
		Appearance:  &req.Appearance,
		Personality: &req.Personality,
	}
	if req.OutfitName != "" {
		character.OutfitName = &req.OutfitName
	}

	if req.Role != "" {
		character.Role = &req.Role
	}
	if req.Gender != "" {
		character.Gender = &req.Gender
	}
	if req.AgeDescription != "" {
		character.AgeDescription = &req.AgeDescription
	}
	if req.VoiceStyle != "" {
		character.VoiceStyle = &req.VoiceStyle
	}
	if req.Prompt != "" {
		character.Prompt = &req.Prompt
	}
	if req.Description != "" {
		character.Description = &req.Description
	}

	if req.ImageURL != "" {
		character.ImageURL = &req.ImageURL
	}
	if req.LocalPath != "" {
		character.LocalPath = &req.LocalPath
	}
	if req.ReferenceImages != nil && len(req.ReferenceImages) > 0 {
		jsonData, err := json.Marshal(req.ReferenceImages)
		if err == nil {
			character.ReferenceImages = datatypes.JSON(jsonData)
		}
	}
	if req.ImageOrientation != "" {
		character.ImageOrientation = &req.ImageOrientation
	}

	if err := s.db.Create(character).Error; err != nil {
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	// 如果指定了 EpisodeID，关联到章节
	if req.EpisodeID != nil {
		var episode models.Episode
		if err := s.db.First(&episode, *req.EpisodeID).Error; err == nil {
			if err := s.db.Model(&episode).Association("Characters").Append(character); err != nil {
				s.log.Errorw("Failed to associate character with episode", "error", err)
			}
		}
	}

	s.log.Infow("Character created successfully", "character_id", character.ID, "drama_id", character.DramaID, "episode_id", req.EpisodeID)
	return character, nil
}

type SaveCharactersRequest struct {
	Characters []models.Character `json:"characters" binding:"required"`
	EpisodeID  *uint              `json:"episode_id"` // 可选：如果提供则关联到指定章节
}

type SaveProgressRequest struct {
	CurrentStep string                 `json:"current_step" binding:"required"`
	StepData    map[string]interface{} `json:"step_data"`
}

type SaveEpisodesRequest struct {
	Episodes []models.Episode `json:"episodes" binding:"required"`
}

func (s *DramaService) SaveOutline(dramaID string, req *SaveOutlineRequest) error {
	var drama models.Drama
	if err := s.db.Where("id = ? ", dramaID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrDramaNotFound
		}
		return err
	}

	updates := map[string]interface{}{
		"title":       req.Title,
		"description": req.Summary,
		"updated_at":  time.Now(),
	}

	if req.Genre != "" {
		updates["genre"] = req.Genre
	}

	if len(req.Tags) > 0 {
		tagsJSON, err := json.Marshal(req.Tags)
		if err != nil {
			s.log.Errorw("Failed to marshal tags", "error", err)
			return err
		}
		updates["tags"] = tagsJSON
	}

	if err := s.db.Model(&drama).Updates(updates).Error; err != nil {
		s.log.Errorw("Failed to save outline", "error", err)
		return err
	}

	s.log.Infow("Outline saved", "drama_id", dramaID)
	return nil
}

func (s *DramaService) GetCharacters(dramaID string, episodeID *string) ([]models.Character, error) {
	var drama models.Drama
	if err := s.db.Where("id = ? ", dramaID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDramaNotFound
		}
		return nil, err
	}

	var characters []models.Character

	if episodeID != nil {
		var episode models.Episode
		if err := s.db.Preload("Characters").Preload("Characters.Children").Where("id = ? AND drama_id = ?", *episodeID, dramaID).First(&episode).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, domain.ErrEpisodeNotFound
			}
			return nil, err
		}
		characters = episode.Characters
	} else {
		if err := s.db.Where("drama_id = ? AND parent_id IS NULL", dramaID).
			Preload("Children").
			Find(&characters).Error; err != nil {
			s.log.Errorw("Failed to get characters", "error", err)
			return nil, err
		}
	}

	// Batch query: get latest image generation per character (including children)
	if len(characters) > 0 {
		var charIDs []uint
		for _, c := range characters {
			charIDs = append(charIDs, c.ID)
			for _, child := range c.Children {
				charIDs = append(charIDs, child.ID)
			}
		}

		var latestGens []models.ImageGeneration
		subQuery := s.db.Model(&models.ImageGeneration{}).
			Select("MAX(id) as id").
			Where("character_id IN ?", charIDs).
			Group("character_id")
		s.db.Where("id IN (?)", subQuery).Find(&latestGens)

		genMap := make(map[uint]*models.ImageGeneration, len(latestGens))
		for i := range latestGens {
			if latestGens[i].CharacterID != nil {
				genMap[*latestGens[i].CharacterID] = &latestGens[i]
			}
		}

		applyGenStatus := func(c *models.Character) {
			imageGen, ok := genMap[c.ID]
			if !ok {
				return
			}
			if imageGen.Status == models.ImageStatusPending || imageGen.Status == models.ImageStatusProcessing {
				statusStr := string(imageGen.Status)
				c.ImageGenerationStatus = &statusStr
			} else if imageGen.Status == models.ImageStatusFailed {
				statusStr := "failed"
				c.ImageGenerationStatus = &statusStr
				if imageGen.ErrorMsg != nil {
					c.ImageGenerationError = imageGen.ErrorMsg
				}
			}
		}
		for i := range characters {
			applyGenStatus(&characters[i])
			for j := range characters[i].Children {
				applyGenStatus(&characters[i].Children[j])
			}
		}
	}

	return characters, nil
}

func (s *DramaService) SaveCharacters(dramaID string, req *SaveCharactersRequest) error {
	// 转换dramaID
	id, err := strconv.ParseUint(dramaID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid drama ID")
	}
	dramaIDUint := uint(id)

	var drama models.Drama
	if err := s.db.Where("id = ? ", dramaIDUint).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrDramaNotFound
		}
		return err
	}

	// 如果指定了EpisodeID，验证章节存在性
	if req.EpisodeID != nil {
		var episode models.Episode
		if err := s.db.Where("id = ? AND drama_id = ?", *req.EpisodeID, dramaIDUint).First(&episode).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrEpisodeNotFound
			}
			return err
		}
	}

	// 获取该项目已存在的所有角色
	var existingCharacters []models.Character
	if err := s.db.Where("drama_id = ?", dramaIDUint).Find(&existingCharacters).Error; err != nil {
		s.log.Errorw("Failed to get existing characters", "error", err)
		return err
	}

	// 创建角色名称到角色的映射
	existingCharMap := make(map[string]*models.Character)
	for i := range existingCharacters {
		existingCharMap[existingCharacters[i].Name] = &existingCharacters[i]
	}

	// 收集需要关联到章节的角色ID
	var characterIDs []uint

	// 创建新角色或复用/更新已有角色
	for _, char := range req.Characters {
		// 1. 如果提供了ID，尝试更新已有角色
		if char.ID > 0 {
			var existing models.Character
			if err := s.db.Where("id = ? AND drama_id = ?", char.ID, dramaIDUint).First(&existing).Error; err == nil {
				// 更新角色信息
				updates := map[string]interface{}{
					"name":        char.Name,
					"role":        char.Role,
					"description": char.Description,
					"personality": char.Personality,
					"appearance":  char.Appearance,
					"image_url":   char.ImageURL,
				}
				if err := s.db.Model(&existing).Updates(updates).Error; err != nil {
					s.log.Errorw("Failed to update character", "error", err, "id", char.ID)
				}
				characterIDs = append(characterIDs, existing.ID)
				continue
			}
		}

		// 2. 如果没有ID但名字已存在，直接复用（可选：也可以选择更新）
		if existingChar, exists := existingCharMap[char.Name]; exists {
			s.log.Infow("Character already exists, reusing", "name", char.Name, "character_id", existingChar.ID)
			characterIDs = append(characterIDs, existingChar.ID)
			continue
		}

		// 3. 角色不存在，创建新角色
		character := models.Character{
			DramaID:     dramaIDUint,
			Name:        char.Name,
			Role:        char.Role,
			Description: char.Description,
			Personality: char.Personality,
			Appearance:  char.Appearance,
			ImageURL:    char.ImageURL,
		}

		if err := s.db.Create(&character).Error; err != nil {
			s.log.Errorw("Failed to create character", "error", err, "name", char.Name)
			continue
		}

		s.log.Infow("New character created", "character_id", character.ID, "name", char.Name)
		characterIDs = append(characterIDs, character.ID)
	}

	// 如果指定了EpisodeID，建立角色与章节的关联
	if req.EpisodeID != nil && len(characterIDs) > 0 {
		var episode models.Episode
		if err := s.db.First(&episode, *req.EpisodeID).Error; err != nil {
			return err
		}

		// 获取角色对象
		var characters []models.Character
		if err := s.db.Where("id IN ?", characterIDs).Find(&characters).Error; err != nil {
			s.log.Errorw("Failed to get characters", "error", err)
			return err
		}

		// 使用GORM的Association API建立多对多关系（会自动去重）
		if err := s.db.Model(&episode).Association("Characters").Append(&characters); err != nil {
			s.log.Errorw("Failed to associate characters with episode", "error", err)
			return err
		}

		s.log.Infow("Characters associated with episode", "episode_id", *req.EpisodeID, "character_count", len(characterIDs))
	}

	if err := s.db.Model(&drama).Update("updated_at", time.Now()).Error; err != nil {
		s.log.Errorw("Failed to update drama timestamp", "error", err)
	}

	s.log.Infow("Characters saved", "drama_id", dramaID, "count", len(req.Characters))
	return nil
}

func (s *DramaService) SaveEpisodes(dramaID string, req *SaveEpisodesRequest) error {
	s.log.Infow("SaveEpisodes called", "drama_id", dramaID, "episodes_count", len(req.Episodes))

	id, err := strconv.ParseUint(dramaID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid drama ID")
	}
	dramaIDUint := uint(id)

	var drama models.Drama
	if err := s.db.Where("id = ?", dramaIDUint).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrDramaNotFound
		}
		return err
	}

	reqEpisodeIDs := make(map[uint]bool)
	for _, ep := range req.Episodes {
		if ep.ID != 0 {
			reqEpisodeIDs[ep.ID] = true
		}
	}

	var allEpisodes []models.Episode
	if err := s.db.Where("drama_id = ?", dramaIDUint).Order("episode_number ASC").Find(&allEpisodes).Error; err != nil {
		return err
	}

	for _, ep := range allEpisodes {
		if !reqEpisodeIDs[ep.ID] {
			s.log.Infow("Deleting episode", "episode_id", ep.ID, "episode_number", ep.EpisodeNum)

			if err := s.db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Where("episode_id = ?", ep.ID).Delete(&models.Storyboard{}).Error; err != nil {
					return err
				}

				if err := tx.Where("episode_id = ?", ep.ID).Delete(&models.Scene{}).Error; err != nil {
					return err
				}

				if err := tx.Delete(&ep).Error; err != nil {
					return err
				}

				return nil
			}); err != nil {
				s.log.Errorw("Failed to delete episode", "error", err, "episode_id", ep.ID)
				return err
			}
		}
	}

	for _, ep := range req.Episodes {
		s.log.Infow("Processing episode", "episode_id", ep.ID, "episode_num", ep.EpisodeNum, "title", ep.Title, "status", ep.Status)

		var existingEpisode models.Episode
		if ep.ID != 0 {
			err := s.db.Where("id = ? AND drama_id = ?", ep.ID, dramaIDUint).First(&existingEpisode).Error

			if err == nil {
				s.log.Infow("Updating existing episode", "episode_id", existingEpisode.ID, "old_episode_number", existingEpisode.EpisodeNum, "new_episode_number", ep.EpisodeNum, "old_title", existingEpisode.Title, "new_title", ep.Title)
				existingEpisode.EpisodeNum = ep.EpisodeNum
				existingEpisode.Title = ep.Title
				existingEpisode.Description = ep.Description
				existingEpisode.ScriptContent = ep.ScriptContent
				existingEpisode.Duration = ep.Duration
				if ep.Status != "" {
					existingEpisode.Status = ep.Status
				}

				if err := s.db.Save(&existingEpisode).Error; err != nil {
					s.log.Errorw("Failed to update episode", "error", err, "episode_id", ep.ID)
				} else {
					s.log.Infow("Episode updated successfully", "episode_id", existingEpisode.ID)
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				s.log.Infow("Creating new episode", "episode_num", ep.EpisodeNum, "title", ep.Title)
				episode := models.Episode{
					DramaID:       dramaIDUint,
					EpisodeNum:    ep.EpisodeNum,
					Title:         ep.Title,
					Description:   ep.Description,
					ScriptContent: ep.ScriptContent,
					Duration:      ep.Duration,
					Status:        "draft",
				}

				if err := s.db.Create(&episode).Error; err != nil {
					s.log.Errorw("Failed to create episode", "error", err, "episode", ep.EpisodeNum)
				} else {
					s.log.Infow("Episode created successfully", "episode_id", episode.ID)
				}
			} else {
				s.log.Errorw("Failed to query episode", "error", err, "episode_id", ep.ID)
			}
		} else {
			s.log.Infow("Creating new episode", "episode_num", ep.EpisodeNum, "title", ep.Title)
			episode := models.Episode{
				DramaID:       dramaIDUint,
				EpisodeNum:    ep.EpisodeNum,
				Title:         ep.Title,
				Description:   ep.Description,
				ScriptContent: ep.ScriptContent,
				Duration:      ep.Duration,
				Status:        "draft",
			}

			if err := s.db.Create(&episode).Error; err != nil {
				s.log.Errorw("Failed to create episode", "error", err, "episode", ep.EpisodeNum)
			} else {
				s.log.Infow("Episode created successfully", "episode_id", episode.ID)
			}
		}
	}

	if err := s.db.Model(&drama).Update("updated_at", time.Now()).Error; err != nil {
		s.log.Errorw("Failed to update drama timestamp", "error", err)
	}

	s.log.Infow("Episodes saved", "drama_id", dramaID, "count", len(req.Episodes))
	return nil
}

type UpdateEpisodeTitleRequest struct {
	EpisodeNumber int    `json:"episode_number" binding:"required"`
	Title         string `json:"title" binding:"required"`
	Status        string `json:"status"`
}

func (s *DramaService) UpdateEpisodeTitle(dramaID string, req *UpdateEpisodeTitleRequest) error {
	s.log.Infow("UpdateEpisodeTitle called", "drama_id", dramaID, "episode_num", req.EpisodeNumber, "new_title", req.Title)

	id, err := strconv.ParseUint(dramaID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid drama ID")
	}
	dramaIDUint := uint(id)

	var episode models.Episode
	err = s.db.Where("drama_id = ? AND episode_number = ?", dramaIDUint, req.EpisodeNumber).First(&episode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrEpisodeNotFound
		}
		s.log.Errorw("Failed to find episode", "error", err)
		return err
	}

	s.log.Infow("Updating episode", "episode_id", episode.ID, "old_title", episode.Title, "new_title", req.Title, "new_status", req.Status)
	episode.Title = req.Title
	if req.Status != "" {
		episode.Status = req.Status
	}

	if err := s.db.Save(&episode).Error; err != nil {
		s.log.Errorw("Failed to update episode title", "error", err)
		return err
	}

	s.log.Infow("Episode title updated successfully", "episode_id", episode.ID)
	return nil
}

func (s *DramaService) SaveProgress(dramaID string, req *SaveProgressRequest) error {
	var drama models.Drama
	if err := s.db.Where("id = ? ", dramaID).First(&drama).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrDramaNotFound
		}
		return err
	}

	// 构建metadata对象
	metadata := make(map[string]interface{})

	// 保留现有metadata
	if drama.Metadata != nil {
		if err := json.Unmarshal(drama.Metadata, &metadata); err != nil {
			s.log.Warnw("Failed to unmarshal existing metadata", "error", err)
		}
	}

	// 更新progress信息
	metadata["current_step"] = req.CurrentStep
	if req.StepData != nil {
		metadata["step_data"] = req.StepData
	}

	// 序列化metadata
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		s.log.Errorw("Failed to marshal metadata", "error", err)
		return err
	}

	updates := map[string]interface{}{
		"metadata":   metadataJSON,
		"updated_at": time.Now(),
	}

	if err := s.db.Model(&drama).Updates(updates).Error; err != nil {
		s.log.Errorw("Failed to save progress", "error", err)
		return err
	}

	s.log.Infow("Progress saved", "drama_id", dramaID, "step", req.CurrentStep)
	return nil
}

// RemoveCharacterFromEpisode 从章节中移除角色关联（不删除角色本身）
func (s *DramaService) RemoveCharacterFromEpisode(episodeID uint, characterID uint) error {
	var episode models.Episode
	if err := s.db.First(&episode, episodeID).Error; err != nil {
		return domain.ErrEpisodeNotFound
	}
	var character models.Character
	if err := s.db.First(&character, characterID).Error; err != nil {
		return domain.ErrCharacterNotFound
	}
	if err := s.db.Model(&episode).Association("Characters").Delete(&character); err != nil {
		s.log.Errorw("Failed to remove character from episode", "error", err, "episode_id", episodeID, "character_id", characterID)
		return err
	}
	s.log.Infow("Character removed from episode", "episode_id", episodeID, "character_id", characterID)
	return nil
}

// RemoveSceneFromEpisode 从章节中移除场景关联（将 episode_id 置空，不删除场景本身）
func (s *DramaService) RemoveSceneFromEpisode(episodeID uint, sceneID uint) error {
	result := s.db.Model(&models.Scene{}).
		Where("id = ? AND episode_id = ?", sceneID, episodeID).
		Update("episode_id", gorm.Expr("NULL"))
	if result.Error != nil {
		s.log.Errorw("Failed to remove scene from episode", "error", result.Error, "episode_id", episodeID, "scene_id", sceneID)
		return result.Error
	}
	if result.RowsAffected == 0 {
		s.log.Warnw("Scene not found in episode, may already be removed", "episode_id", episodeID, "scene_id", sceneID)
		return nil
	}
	s.log.Infow("Scene removed from episode", "episode_id", episodeID, "scene_id", sceneID)
	return nil
}

// AddCharacterToEpisode 将已有角色关联到章节（不创建新角色）
func (s *DramaService) AddCharacterToEpisode(episodeID uint, characterID uint) error {
	var episode models.Episode
	if err := s.db.First(&episode, episodeID).Error; err != nil {
		return domain.ErrEpisodeNotFound
	}
	var character models.Character
	if err := s.db.First(&character, characterID).Error; err != nil {
		return domain.ErrCharacterNotFound
	}
	if err := s.db.Model(&episode).Association("Characters").Append(&character); err != nil {
		s.log.Errorw("Failed to add character to episode", "error", err, "episode_id", episodeID, "character_id", characterID)
		return err
	}
	s.log.Infow("Character added to episode", "episode_id", episodeID, "character_id", characterID)
	return nil
}

// AddSceneToEpisode 将已有场景关联到章节（设置 episode_id）
func (s *DramaService) AddSceneToEpisode(episodeID uint, sceneID uint) error {
	var episode models.Episode
	if err := s.db.First(&episode, episodeID).Error; err != nil {
		return domain.ErrEpisodeNotFound
	}
	result := s.db.Model(&models.Scene{}).
		Where("id = ? AND drama_id = ?", sceneID, episode.DramaID).
		Update("episode_id", episodeID)
	if result.Error != nil {
		s.log.Errorw("Failed to add scene to episode", "error", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("scene not found in this drama")
	}
	s.log.Infow("Scene added to episode", "episode_id", episodeID, "scene_id", sceneID)
	return nil
}

// AddPropToEpisode 将已有道具关联到章节（many-to-many）
func (s *DramaService) AddPropToEpisode(episodeID uint, propID uint) error {
	var episode models.Episode
	if err := s.db.First(&episode, episodeID).Error; err != nil {
		return domain.ErrEpisodeNotFound
	}
	var prop models.Prop
	if err := s.db.First(&prop, propID).Error; err != nil {
		return domain.ErrPropNotFound
	}
	if err := s.db.Model(&episode).Association("Props").Append(&prop); err != nil {
		s.log.Errorw("Failed to add prop to episode", "error", err, "episode_id", episodeID, "prop_id", propID)
		return err
	}
	s.log.Infow("Prop added to episode", "episode_id", episodeID, "prop_id", propID)
	return nil
}

// RemovePropFromEpisode 从章节中移除道具关联（不删除道具本身）
func (s *DramaService) RemovePropFromEpisode(episodeID uint, propID uint) error {
	var episode models.Episode
	if err := s.db.First(&episode, episodeID).Error; err != nil {
		return domain.ErrEpisodeNotFound
	}
	var prop models.Prop
	if err := s.db.First(&prop, propID).Error; err != nil {
		return domain.ErrPropNotFound
	}
	if err := s.db.Model(&episode).Association("Props").Delete(&prop); err != nil {
		s.log.Errorw("Failed to remove prop from episode", "error", err, "episode_id", episodeID, "prop_id", propID)
		return err
	}
	s.log.Infow("Prop removed from episode", "episode_id", episodeID, "prop_id", propID)
	return nil
}

// addBaseURLToScenes 为剧本中所有场景的 local_path 添加 base_url 前缀
func (s *DramaService) addBaseURLToScenes(drama *models.Drama) {
	// 处理 drama.Scenes
	for i := range drama.Scenes {
		if drama.Scenes[i].LocalPath != nil && *drama.Scenes[i].LocalPath != "" {
			fullPath := fmt.Sprintf("%s/%s", s.baseURL, *drama.Scenes[i].LocalPath)
			drama.Scenes[i].LocalPath = &fullPath
		}
	}

	// 处理 drama.Episodes[].Scenes
	for i := range drama.Episodes {
		for j := range drama.Episodes[i].Scenes {
			if drama.Episodes[i].Scenes[j].LocalPath != nil && *drama.Episodes[i].Scenes[j].LocalPath != "" {
				fullPath := fmt.Sprintf("%s/%s", s.baseURL, *drama.Episodes[i].Scenes[j].LocalPath)
				drama.Episodes[i].Scenes[j].LocalPath = &fullPath
			}
		}
	}
}
