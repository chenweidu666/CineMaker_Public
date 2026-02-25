package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cinemaker/backend/domain"
	models "github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type StoryboardCompositionService struct {
	db        *gorm.DB
	log       *logger.Logger
	imageGen  *ImageGenerationService
	aiService *AIService
	config    *config.Config
}

func NewStoryboardCompositionService(db *gorm.DB, log *logger.Logger, imageGen *ImageGenerationService, aiService *AIService, cfg *config.Config) *StoryboardCompositionService {
	return &StoryboardCompositionService{
		db:        db,
		log:       log,
		imageGen:  imageGen,
		aiService: aiService,
		config:    cfg,
	}
}

type SceneCharacterInfo struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	ImageURL  *string `json:"image_url,omitempty"`
	LocalPath *string `json:"local_path,omitempty"`
}

type ScenePropInfo struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Type        *string `json:"type,omitempty"`
	Description *string `json:"description,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	LocalPath   *string `json:"local_path,omitempty"`
}

type SceneBackgroundInfo struct {
	ID        uint    `json:"id"`
	Location  string  `json:"location"`
	Time      string  `json:"time"`
	ImageURL  *string `json:"image_url,omitempty"`
	LocalPath *string `json:"local_path,omitempty"`
	Status    string  `json:"status"`
}

type SceneCompositionInfo struct {
	ID                    uint                 `json:"id"`
	StoryboardNumber      int                  `json:"storyboard_number"`
	Title                 *string              `json:"title"`
	Description           *string              `json:"description"`
	ShotType              *string              `json:"shot_type"`
	Angle                 *string              `json:"angle"`
	Movement              *string              `json:"movement"`
	Location              *string              `json:"location"`
	Time                  *string              `json:"time"`
	Duration              int                  `json:"duration"`
	Dialogue              *string              `json:"dialogue"`
	Action                *string              `json:"action"`
	FirstFrameDesc        *string              `json:"first_frame_desc,omitempty"`
	MiddleActionDesc      *string              `json:"middle_action_desc,omitempty"`
	LastFrameDesc         *string              `json:"last_frame_desc,omitempty"`
	Result                *string              `json:"result"`
	Atmosphere            *string              `json:"atmosphere"`
	BgmPrompt             *string              `json:"bgm_prompt,omitempty"`
	SoundEffect           *string              `json:"sound_effect,omitempty"`
	ImagePrompt           *string              `json:"image_prompt,omitempty"`
	VideoPrompt           *string              `json:"video_prompt,omitempty"`
	Characters            []SceneCharacterInfo `json:"characters"`
	Props                 []ScenePropInfo      `json:"props"`
	Background            *SceneBackgroundInfo `json:"background"`
	SceneID               *uint                `json:"scene_id"`
	ComposedImage         *string              `json:"composed_image,omitempty"`
	VideoURL              *string              `json:"video_url,omitempty"`
	ImageGenerationID     *uint                `json:"image_generation_id,omitempty"`
	ImageGenerationStatus *string              `json:"image_generation_status,omitempty"`
	VideoGenerationID     *uint                `json:"video_generation_id,omitempty"`
	VideoGenerationStatus *string              `json:"video_generation_status,omitempty"`
}

func (s *StoryboardCompositionService) GetScenesForEpisode(episodeID string) ([]SceneCompositionInfo, error) {
	// 验证权限
	var episode models.Episode
	err := s.db.Preload("Drama").Where("id = ?", episodeID).First(&episode).Error
	if err != nil {
		s.log.Errorw("Episode not found", "episode_id", episodeID, "error", err)
		return nil, domain.ErrEpisodeNotFound
	}

	s.log.Infow("GetScenesForEpisode auth check",
		"episode_id", episodeID,
		"drama_id", episode.DramaID)

	// 获取分镜列表
	var storyboards []models.Storyboard
	if err := s.db.Where("episode_id = ?", episodeID).
		Preload("Characters").
		Preload("Props").
		Order("storyboard_number ASC").
		Find(&storyboards).Error; err != nil {
		return nil, fmt.Errorf("failed to load storyboards: %w", err)
	}

	// 获取所有角色（用于匹配角色信息）
	var characters []models.Character
	if err := s.db.Where("drama_id = ?", episode.DramaID).Find(&characters).Error; err != nil {
		s.log.Warnw("Failed to load characters", "error", err)
	}

	// 创建角色ID到角色信息的映射
	charIDToInfo := make(map[uint]*models.Character)
	for i := range characters {
		charIDToInfo[characters[i].ID] = &characters[i]
	}

	// 获取所有场景ID
	var sceneIDs []uint
	for _, storyboard := range storyboards {
		if storyboard.SceneID != nil {
			sceneIDs = append(sceneIDs, *storyboard.SceneID)
		}
	}

	// 批量获取场景信息
	var scenes []models.Scene
	sceneMap := make(map[uint]*models.Scene)
	if len(sceneIDs) > 0 {
		if err := s.db.Where("id IN ?", sceneIDs).Find(&scenes).Error; err == nil {
			for i := range scenes {
				sceneMap[scenes[i].ID] = &scenes[i]
			}
		}
	}

	// 获取分镜的合成图片（从 image_generations 表）
	storyboardIDs := make([]uint, len(storyboards))
	for i, storyboard := range storyboards {
		storyboardIDs[i] = storyboard.ID
	}

	imageGenMap := make(map[uint]string)                      // storyboard_id -> image_url
	imageGenTaskMap := make(map[uint]*models.ImageGeneration) // storyboard_id -> processing task
	if len(storyboardIDs) > 0 {
		var imageGens []models.ImageGeneration
		// 查询已完成的图片生成记录，每个镜头只取最新的一条
		if err := s.db.Where("storyboard_id IN ? AND status = ?", storyboardIDs, models.ImageStatusCompleted).
			Order("created_at DESC").
			Find(&imageGens).Error; err == nil {
			// 为每个镜头保留最新的一条记录
			for _, ig := range imageGens {
				if ig.StoryboardID != nil {
					if _, exists := imageGenMap[*ig.StoryboardID]; !exists {
						if ig.ImageURL != nil {
							imageGenMap[*ig.StoryboardID] = *ig.ImageURL
						}
					}
				}
			}
		}

		// 查询进行中的图片生成任务
		var processingImageGens []models.ImageGeneration
		if err := s.db.Where("storyboard_id IN ? AND status = ?", storyboardIDs, models.ImageStatusProcessing).
			Order("created_at DESC").
			Find(&processingImageGens).Error; err == nil {
			for _, ig := range processingImageGens {
				if ig.StoryboardID != nil {
					if _, exists := imageGenTaskMap[*ig.StoryboardID]; !exists {
						igCopy := ig
						imageGenTaskMap[*ig.StoryboardID] = &igCopy
					}
				}
			}
		}
	}

	// 批量查询进行中的视频生成任务
	videoGenTaskMap := make(map[uint]*models.VideoGeneration) // storyboard_id -> processing task
	if len(storyboardIDs) > 0 {
		var processingVideoGens []models.VideoGeneration
		if err := s.db.Where("scene_id IN ? AND status = ?", storyboardIDs, models.VideoStatusProcessing).
			Order("created_at DESC").
			Find(&processingVideoGens).Error; err == nil {
			for _, vg := range processingVideoGens {
				if vg.StoryboardID != nil {
					if _, exists := videoGenTaskMap[*vg.StoryboardID]; !exists {
						vgCopy := vg
						videoGenTaskMap[*vg.StoryboardID] = &vgCopy
					}
				}
			}
		}
	}

	// 构建返回结果
	var result []SceneCompositionInfo
	for _, storyboard := range storyboards {
		storyboardInfo := SceneCompositionInfo{
			ID:               storyboard.ID,
			StoryboardNumber: storyboard.StoryboardNumber,
			Title:            storyboard.Title,
			Description:      storyboard.Description,
			ShotType:         storyboard.ShotType,
			Angle:            storyboard.Angle,
			Movement:         storyboard.Movement,
			Location:         storyboard.Location,
			Time:             storyboard.Time,
			Duration:         storyboard.Duration,
			Action:           storyboard.Action,
			FirstFrameDesc:   storyboard.FirstFrameDesc,
			MiddleActionDesc: storyboard.MiddleActionDesc,
			LastFrameDesc:    storyboard.LastFrameDesc,
			Dialogue:         storyboard.Dialogue,
			Result:           storyboard.Result,
			Atmosphere:       storyboard.Atmosphere,
			BgmPrompt:        storyboard.BgmPrompt,
			SoundEffect:      storyboard.SoundEffect,
			ImagePrompt:      storyboard.ImagePrompt,
			VideoPrompt:      storyboard.VideoPrompt,
			SceneID:          storyboard.SceneID,
		}

		// 直接使用关联的角色信息
		if len(storyboard.Characters) > 0 {
			for _, char := range storyboard.Characters {
				storyboardChar := SceneCharacterInfo{
					ID:        char.ID,
					Name:      char.Name,
					ImageURL:  char.ImageURL,
					LocalPath: char.LocalPath,
				}
				storyboardInfo.Characters = append(storyboardInfo.Characters, storyboardChar)
			}
		}

		// 直接使用关联的道具信息
		if len(storyboard.Props) > 0 {
			for _, prop := range storyboard.Props {
				storyboardProp := ScenePropInfo{
					ID:          prop.ID,
					Name:        prop.Name,
					Type:        prop.Type,
					Description: prop.Description,
					ImageURL:    prop.ImageURL,
					LocalPath:   prop.LocalPath,
				}
				storyboardInfo.Props = append(storyboardInfo.Props, storyboardProp)
			}
		}

		// 添加场景信息
		if storyboard.SceneID != nil {
			if scene, ok := sceneMap[*storyboard.SceneID]; ok {
				storyboardInfo.Background = &SceneBackgroundInfo{
					ID:        scene.ID,
					Location:  scene.Location,
					Time:      scene.Time,
					ImageURL:  scene.ImageURL,
					LocalPath: scene.LocalPath,
					Status:    scene.Status,
				}
			}
		}

		// 添加合成图片
		if imageURL, ok := imageGenMap[storyboard.ID]; ok {
			storyboardInfo.ComposedImage = &imageURL
		}

		// 添加视频URL
		if storyboard.VideoURL != nil {
			storyboardInfo.VideoURL = storyboard.VideoURL
		}

		// 添加进行中的图片生成任务信息
		if imageTask, ok := imageGenTaskMap[storyboard.ID]; ok {
			storyboardInfo.ImageGenerationID = &imageTask.ID
			statusStr := string(imageTask.Status)
			storyboardInfo.ImageGenerationStatus = &statusStr
		}

		// 添加进行中的视频生成任务信息
		if videoTask, ok := videoGenTaskMap[storyboard.ID]; ok {
			storyboardInfo.VideoGenerationID = &videoTask.ID
			statusStr := string(videoTask.Status)
			storyboardInfo.VideoGenerationStatus = &statusStr
		}

		result = append(result, storyboardInfo)
	}

	return result, nil
}

type UpdateSceneRequest struct {
	SceneID     *uint   `json:"scene_id"`
	Characters  []uint  `json:"characters"` // 改为存储角色ID数组
	Location    *string `json:"location"`
	Time        *string `json:"time"`
	Action      *string `json:"action"`
	Dialogue    *string `json:"dialogue"`
	Description *string `json:"description"`
	Duration    *int    `json:"duration"`
	ImageURL    *string `json:"image_url"`
	LocalPath   *string `json:"local_path"`
	ImagePrompt *string `json:"image_prompt"`
	VideoPrompt *string `json:"video_prompt"`
}

func (s *StoryboardCompositionService) UpdateScene(sceneID string, req *UpdateSceneRequest) error {
	// 获取分镜并验证权限
	var storyboard models.Storyboard
	err := s.db.Preload("Episode.Drama").Where("id = ?", sceneID).First(&storyboard).Error
	if err != nil {
		return domain.ErrSceneNotFound
	}

	// 构建更新数据
	updates := make(map[string]interface{})

	// 更新背景ID
	if req.SceneID != nil {
		updates["scene_id"] = req.SceneID
	}

	// 更新角色列表（直接存储ID数组）
	if req.Characters != nil {
		charactersJSON, err := json.Marshal(req.Characters)
		if err != nil {
			return fmt.Errorf("failed to serialize characters: %w", err)
		}
		updates["characters"] = charactersJSON
	}

	// 更新场景信息字段
	if req.Location != nil {
		updates["location"] = req.Location
	}
	if req.Time != nil {
		updates["time"] = req.Time
	}
	if req.Action != nil {
		updates["action"] = req.Action
	}
	if req.Dialogue != nil {
		updates["dialogue"] = req.Dialogue
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.ImageURL != nil {
		updates["image_url"] = req.ImageURL
	}
	if req.LocalPath != nil {
		updates["local_path"] = req.LocalPath
	}
	if req.ImagePrompt != nil {
		updates["image_prompt"] = req.ImagePrompt
	}
	if req.VideoPrompt != nil {
		updates["video_prompt"] = req.VideoPrompt
	}

	// 执行更新
	if len(updates) > 0 {
		if err := s.db.Model(&models.Storyboard{}).Where("id = ?", sceneID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update scene: %w", err)
		}
	}

	s.log.Infow("Scene updated", "scene_id", sceneID, "updates", updates)
	return nil
}

type GenerateSceneImageRequest struct {
	SceneID         uint     `json:"scene_id"`
	Prompt          string   `json:"prompt"`
	Model           string   `json:"model"`
	Size            string   `json:"size"`
	ReferenceImages []string `json:"reference_images"`
}

func (s *StoryboardCompositionService) GenerateSceneImage(req *GenerateSceneImageRequest) (*models.ImageGeneration, string, error) {
	// 获取场景并验证权限
	var scene models.Scene
	err := s.db.Where("id = ?", req.SceneID).First(&scene).Error
	if err != nil {
		return nil, "", domain.ErrSceneNotFound
	}

	// 验证权限：通过DramaID查询Drama
	var drama models.Drama
	if err := s.db.Where("id = ? ", scene.DramaID).First(&drama).Error; err != nil {
		return nil, "", domain.ErrUnauthorized
	}

	// 获取剧本风格
	dramaStyle := drama.Style
	if dramaStyle == "" {
		dramaStyle = "realistic"
	}
	sceneStyleNameMap := map[string]string{
		"realistic": "超写实摄影风格",
		"comic":     "漫画风格",
	}
	sceneStyleName := sceneStyleNameMap[dramaStyle]
	if sceneStyleName == "" {
		sceneStyleName = dramaStyle + "风格"
	}

	// 构建场景图片生成提示词
	prompt := req.Prompt
	if prompt == "" {
		prompt = scene.Prompt
		if prompt == "" {
			sceneName := scene.Location
			if scene.Name != nil && *scene.Name != "" {
				sceneName = *scene.Name
			}

			// 收集场景标签信息
			var tags []string
			if scene.Description != nil && *scene.Description != "" {
				tags = append(tags, "环境描述："+*scene.Description)
			}
			if scene.Atmosphere != nil && *scene.Atmosphere != "" {
				tags = append(tags, "氛围："+*scene.Atmosphere)
			}
			if scene.Lighting != nil && *scene.Lighting != "" {
				tags = append(tags, "光线："+*scene.Lighting)
			}
			sceneTypeLabel := "室内"
			if scene.SceneType != nil && *scene.SceneType != "" {
				sceneTypeLabel = *scene.SceneType
			}

			tagInfo := ""
			if len(tags) > 0 {
				tagInfo = strings.Join(tags, "；") + "。"
			}

			timeInfo := ""
			if scene.Time != "" {
				timeInfo = fmt.Sprintf("，时间：%s", scene.Time)
			}

			if sceneTypeLabel == "室外" {
				prompt = fmt.Sprintf("21:9超宽银幕构图，%s，电影级场景空镜头，无人物。"+
					"画面展示「%s」的全貌，采用中远景机位，"+
					"具有电影感的纵深透视与层次分明的前中远景构图。"+
					"场景位置：%s%s。%s"+
					"画面质感细腻，光影真实，色调统一，适合作为影视场景参考。",
					sceneStyleName, sceneName, scene.Location, timeInfo, tagInfo)
			} else {
				prompt = fmt.Sprintf("21:9超宽银幕构图，%s，电影级室内场景空镜头，无人物。"+
					"画面展示「%s」的内景全貌，采用中景偏广的机位，"+
					"展现空间布局、家具陈设与环境氛围，具有电影感的景深与光影层次。"+
					"场景位置：%s%s。%s"+
					"画面质感细腻，光影真实，色调统一，适合作为影视场景参考。",
					sceneStyleName, sceneName, scene.Location, timeInfo, tagInfo)
			}
		}
		s.log.Infow("Using scene prompt", "scene_id", req.SceneID, "prompt", prompt)
	}

	// 动态注入风格标签（与角色生成逻辑一致）
	if !strings.Contains(prompt, "风格") {
		prompt = sceneStyleName + "，" + prompt
	}

	// 场景图使用 21:9 超宽银幕比例，电影感更强
	imageSize := req.Size
	if imageSize == "" {
		imageSize = "3360x1440"
	}
	sceneWidth := 3360
	sceneHeight := 1440
	if imageSize != "3360x1440" {
		fmt.Sscanf(imageSize, "%dx%d", &sceneWidth, &sceneHeight)
	}

	// 使用imageGen服务直接生成
	if s.imageGen != nil {
		genReq := &GenerateImageRequest{
			SceneID:         &req.SceneID,
			DramaID:         fmt.Sprintf("%d", scene.DramaID),
			ImageType:       string(models.ImageTypeScene),
			Prompt:          prompt,
			Model:           req.Model,
			Size:            imageSize,
			Quality:         "standard",
			Width:           &sceneWidth,
			Height:          &sceneHeight,
			ReferenceImages: req.ReferenceImages,
		}
		imageGen, err := s.imageGen.GenerateImage(genReq)
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate image: %w", err)
		}

		// 更新场景的image_url（仅部分更新，避免覆盖其他字段）
		if imageGen.ImageURL != nil {
			sceneUpdates := map[string]interface{}{
				"image_url": *imageGen.ImageURL,
				"status":    "generated",
			}
			if imageGen.LocalPath != nil {
				sceneUpdates["local_path"] = *imageGen.LocalPath
			}
			if err := s.db.Model(&models.Scene{}).Where("id = ?", req.SceneID).Updates(sceneUpdates).Error; err != nil {
				s.log.Errorw("Failed to update scene image url", "error", err)
			}
		}

		s.log.Infow("Scene image generation created", "scene_id", req.SceneID, "image_gen_id", imageGen.ID)
		return imageGen, prompt, nil
	}

	return nil, "", fmt.Errorf("image generation service not available")
}

type UpdateScenePromptRequest struct {
	Name             string  `json:"name"`
	Prompt           string  `json:"prompt"`
	ReferenceImages  []any   `json:"reference_images"`
	ImageOrientation *string `json:"image_orientation"`
}

func (s *StoryboardCompositionService) UpdateScenePrompt(sceneID string, req *UpdateScenePromptRequest) error {
	var scene models.Scene
	if err := s.db.Where("id = ?", sceneID).First(&scene).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ErrSceneNotFound
		}
		return fmt.Errorf("failed to find scene: %w", err)
	}

	updates := map[string]interface{}{
		"prompt": req.Prompt,
	}

	if req.Name != "" {
		updates["name"] = req.Name
	}

	if req.ReferenceImages != nil {
		jsonData, err := json.Marshal(req.ReferenceImages)
		if err == nil {
			updates["reference_images"] = datatypes.JSON(jsonData)
		}
	}

	if req.ImageOrientation != nil && *req.ImageOrientation != "" {
		updates["image_orientation"] = *req.ImageOrientation
	}

	if err := s.db.Model(&models.Scene{}).Where("id = ?", sceneID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update scene prompt: %w", err)
	}

	s.log.Infow("Scene prompt updated", "scene_id", sceneID, "prompt", req.Prompt)
	return nil
}

type UpdateSceneInfoRequest struct {
	Location         *string `json:"location"`
	Time             *string `json:"time"`
	Prompt           *string `json:"prompt"`
	Description      *string `json:"description"`
	ImageURL         *string `json:"image_url"`
	LocalPath        *string `json:"local_path"`
	ReferenceImages  []any   `json:"reference_images"`
	ImageOrientation *string `json:"image_orientation"`
}

func (s *StoryboardCompositionService) UpdateSceneInfo(sceneID string, req *UpdateSceneInfoRequest) error {
	var scene models.Scene
	if err := s.db.Where("id = ?", sceneID).First(&scene).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ErrSceneNotFound
		}
		return fmt.Errorf("failed to find scene: %w", err)
	}

	updates := make(map[string]interface{})
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.Time != nil {
		updates["time"] = *req.Time
	}
	if req.Prompt != nil {
		updates["prompt"] = *req.Prompt
	}
	if req.Description != nil {
		updates["description"] = *req.Description
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

	if len(updates) > 0 {
		if err := s.db.Model(&scene).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update scene: %w", err)
		}
	}

	s.log.Infow("Scene info updated", "scene_id", sceneID, "updates", updates)
	return nil
}

func (s *StoryboardCompositionService) DeleteScene(sceneID string) error {
	var scene models.Scene
	if err := s.db.Where("id = ?", sceneID).First(&scene).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ErrSceneNotFound
		}
		return fmt.Errorf("failed to find scene: %w", err)
	}

	// 删除场景
	if err := s.db.Delete(&scene).Error; err != nil {
		return fmt.Errorf("failed to delete scene: %w", err)
	}

	s.log.Infow("Scene deleted successfully", "scene_id", sceneID)
	return nil
}

func (s *StoryboardCompositionService) PolishPrompt(prompt, promptType, orientation, style string, referenceImages []string) (string, error) {
	promptI18n := NewPromptI18n(s.config)

	stylePrompt := promptI18n.GetStylePrompt(style)
	if stylePrompt == "" {
		stylePrompt = ""
	}

	orientationText := ""
	if orientation == "portrait" {
		orientationText = "（9:16竖屏）"
	} else if orientation == "landscape" {
		orientationText = "（16:9横屏）"
	}

	// 风格中文名映射
	styleNameMap := map[string]string{
		"realistic": "超写实摄影风格",
		"comic":     "漫画风格",
	}
	styleName := styleNameMap[style]
	if styleName == "" && style != "" {
		styleName = style + "风格"
	}
	if styleName == "" {
		styleName = "超写实摄影风格"
	}

	var systemPrompt string
	switch promptType {
	case "refine":
		systemPrompt = "你是一位专业的影视制作助手。请严格按照用户的指示处理文本，按照用户要求的格式返回结果。不要添加额外说明。"
	case "script":
		systemPrompt = "你是一位专业的影视编剧，擅长将剧情构想转化为标准化的分镜输入剧本。\n\n" +
			"你的任务是根据用户提供的完整提示（包含剧情构想、角色、场景、道具信息和格式要求），生成一段标准化的剧本。\n\n" +
			"核心要求：\n" +
			"1. 严格遵守用户提示中指定的格式要求\n" +
			"2. 对话要生动自然，符合角色性格\n" +
			"3. 合理利用提供的角色、场景和道具信息\n" +
			"4. 包含完整的舞台指示和动作描写\n" +
			"5. 直接输出剧本内容，不要添加额外说明或markdown代码块"
	case "character":
		systemPrompt = "你是一位专业的角色设定图提示词优化专家。提示词用于豆包Seedream图片生成模型。\n\n" +
			"用简洁连贯的自然语言描述角色外观。提示词不超过300个汉字。\n\n" +
			"当前剧本画面风格：" + styleName + "\n\n" +
			"核心要求：\n" +
			"1. 提示词开头必须明确写出「" + styleName + "」，风格贯穿整个描述\n" +
			"2. 4:3横向画面，在一张图中从左到右并排展示同一角色的三个全身视角：左边正面、中间3/4侧面、右边背面，必须是三个视角\n" +
			"3. 每个视角都是从头顶到鞋底的完整全身站立像，必须看到鞋子和脚，绝对不能裁切腿部和脚部\n" +
			"4. 纯白色背景，无任何场景、装饰、光影效果（禁止描述任何光线、光照、光影）\n" +
			"5. 三个视角外貌服饰配色完全一致\n" +
			"6. 包含面部特征、服装细节、发型发色、体型比例\n" +
			"7. 重要：只输出纯视觉描述，去掉所有剧情叙事和光影描述\n" +
			"8. 不要出现分辨率数值\n\n" +
			"直接输出提示词，不要任何解释。"
	case "prop":
		systemPrompt = "你是一位专业的道具设定图提示词优化专家。提示词用于豆包Seedream图片生成模型。\n\n" +
			"用简洁连贯的自然语言描述道具外观。提示词不超过300个汉字。\n\n" +
			"当前剧本画面风格：" + styleName + "\n\n" +
			"核心要求：\n" +
			"1. 提示词开头必须明确写出「" + styleName + "」，风格贯穿整个描述\n" +
			"2. 4:3横向构图，三视图布局从左到右并排展示同一道具的三个视角：正面全貌、3/4侧面、背面或内部细节特写，各视角外观材质配色完全一致\n" +
			"3. 纯白色背景，无任何场景、装饰、光影效果（禁止描述任何光线、光照、光影）\n" +
			"4. 各角度外观材质配色完全一致\n" +
			"5. 体现整体造型、材质质感、细节特征\n" +
			"6. 重要：只输出纯视觉描述，去掉所有剧情叙事和光影描述\n" +
			"7. 不要出现分辨率数值\n\n" +
			"直接输出提示词，不要任何解释。"
	default:
		systemPrompt = "你是一位专业的AI图片提示词优化专家。提示词用于豆包Seedream图片生成模型。\n\n" +
			"官方建议：用简洁连贯的自然语言写明主体+行为+环境，再补充风格、色彩、光影、构图等美学元素。提示词不超过300个汉字。\n\n" +
			"当前剧本画面风格：" + styleName + "\n\n" +
			"要求：\n" +
			"1. 保持用户的原始意图\n" +
			"2. 提示词开头必须明确写出「" + styleName + "」，风格贯穿整个描述\n" +
			"3. 21:9超宽银幕构图，电影级场景空镜头，展现空间全貌与氛围，具有电影感的纵深透视与前中远景层次，无人物\n" +
			"4. 如果有参考图片，根据风格优化，不要提及文件路径\n" +
			"5. 输出中文提示词\n" +
			"6. 重要：只输出纯视觉描述，必须去掉所有剧情叙事（如人物关系、故事背景等），Seedream是图片生成模型不理解剧情，叙事内容会浪费prompt空间甚至干扰生成效果\n" +
			"7. 不要出现分辨率数值\n\n" +
			"直接输出提示词，不要任何解释。"
	}

	polishedPrompt, err := s.aiService.GenerateText(prompt, systemPrompt)
	if err != nil {
		s.log.Errorw("Failed to polish prompt with AI", "error", err)
		return "", fmt.Errorf("failed to polish prompt: %w", err)
	}

	if len(referenceImages) > 0 {
		refDesc := "\n\n输入的参考图片说明（按顺序对应输入的图片）：\n"
		for i, imgPath := range referenceImages {
			refDesc += fmt.Sprintf("【参考图片%d】%s\n", i+1, imgPath)
		}
		refDesc += "请严格按照以上参考图片来生成图片，保持风格一致。"
		polishedPrompt += refDesc
	}

	if orientationText != "" {
		polishedPrompt += orientationText
	}

	return polishedPrompt, nil
}

func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

type CreateSceneRequest struct {
	DramaID          uint   `json:"drama_id"`
	EpisodeID        *uint  `json:"episode_id"` // 添加章节ID字段
	Location         string `json:"location"`
	Time             string `json:"time"`
	Prompt           string `json:"prompt"`
	ImageURL         string `json:"image_url"`
	LocalPath        string `json:"local_path"`
	Description      string `json:"description"`
	ReferenceImages  []any  `json:"reference_images"`
	ImageOrientation string `json:"image_orientation"`
}

func (s *StoryboardCompositionService) CreateScene(req *CreateSceneRequest) (*models.Scene, error) {
	scene := &models.Scene{
		DramaID:   req.DramaID,
		EpisodeID: req.EpisodeID, // 设置章节ID
		Location:  req.Location,
		Time:      req.Time,
		Prompt:    req.Prompt,
		Status:    "draft",
	}

	if req.ImageURL != "" {
		scene.ImageURL = &req.ImageURL
		scene.Status = "completed"
	}
	if req.LocalPath != "" {
		scene.LocalPath = &req.LocalPath
	}
	if req.ReferenceImages != nil && len(req.ReferenceImages) > 0 {
		jsonData, err := json.Marshal(req.ReferenceImages)
		if err == nil {
			scene.ReferenceImages = datatypes.JSON(jsonData)
		}
	}
	if req.ImageOrientation != "" {
		scene.ImageOrientation = &req.ImageOrientation
	}

	if err := s.db.Create(scene).Error; err != nil {
		return nil, fmt.Errorf("failed to create scene: %w", err)
	}

	s.log.Infow("Scene created successfully", "scene_id", scene.ID, "drama_id", scene.DramaID, "episode_id", req.EpisodeID)
	return scene, nil
}

func (s *StoryboardCompositionService) GetScenesByDramaID(dramaID string) ([]models.Scene, error) {
	var scenes []models.Scene
	err := s.db.Where("drama_id = ?", dramaID).Find(&scenes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get scenes by drama id: %w", err)
	}
	return scenes, nil
}
