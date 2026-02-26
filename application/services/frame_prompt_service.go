package services

import (
	"fmt"
	"strings"

	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"gorm.io/gorm"
)

// FramePromptService 处理帧提示词生成
type FramePromptService struct {
	db          *gorm.DB
	aiService   *AIService
	log         *logger.Logger
	config      *config.Config
	promptI18n  *PromptI18n
	taskService *TaskService
}

// NewFramePromptService 创建帧提示词服务
func NewFramePromptService(db *gorm.DB, cfg *config.Config, log *logger.Logger) *FramePromptService {
	return &FramePromptService{
		db:          db,
		aiService:   NewAIService(db, log),
		log:         log,
		config:      cfg,
		promptI18n:  NewPromptI18n(cfg),
		taskService: NewTaskService(db, log),
	}
}

// FrameType 帧类型
type FrameType string

const (
	FrameTypeFirst  FrameType = "first"  // 首帧
	FrameTypeKey    FrameType = "key"    // 关键帧
	FrameTypeLast   FrameType = "last"   // 尾帧
	FrameTypePanel  FrameType = "panel"  // 分镜板（3格组合）
	FrameTypeAction FrameType = "action" // 动作序列（5格）
)

// GenerateFramePromptRequest 生成帧提示词请求
type GenerateFramePromptRequest struct {
	StoryboardID string    `json:"storyboard_id"`
	FrameType    FrameType `json:"frame_type"`
	// 可选参数
	PanelCount int    `json:"panel_count,omitempty"` // 分镜板格数，默认3
	ImageRatio string `json:"image_ratio,omitempty"` // 画面比例：16:9（横屏）或 9:16（竖屏），默认 16:9
}

// FramePromptResponse 帧提示词响应
type FramePromptResponse struct {
	FrameType   FrameType          `json:"frame_type"`
	SingleFrame *SingleFramePrompt `json:"single_frame,omitempty"` // 单帧提示词
	MultiFrame  *MultiFramePrompt  `json:"multi_frame,omitempty"`  // 多帧提示词
}

// SingleFramePrompt 单帧提示词
type SingleFramePrompt struct {
	Prompt      string `json:"prompt"`
	Description string `json:"description"`
}

// MultiFramePrompt 多帧提示词
type MultiFramePrompt struct {
	Layout string              `json:"layout"` // horizontal_3, grid_2x2 等
	Frames []SingleFramePrompt `json:"frames"`
}

// GenerateFramePrompt 生成指定类型的帧提示词并保存到frame_prompts表
func (s *FramePromptService) GenerateFramePrompt(req GenerateFramePromptRequest, model string) (string, error) {
	// 查询分镜信息
	var storyboard models.Storyboard
	if err := s.db.Preload("Characters").Preload("Props").First(&storyboard, req.StoryboardID).Error; err != nil {
		return "", fmt.Errorf("storyboard not found: %w", err)
	}

	// 创建任务
	task, err := s.taskService.CreateTask("frame_prompt_generation", req.StoryboardID)
	if err != nil {
		s.log.Errorw("Failed to create frame prompt generation task", "error", err, "storyboard_id", req.StoryboardID)
		return "", fmt.Errorf("创建任务失败: %w", err)
	}

	// 异步处理帧提示词生成
	go s.processFramePromptGeneration(task.ID, req, model)

	s.log.Infow("Frame prompt generation task created", "task_id", task.ID, "storyboard_id", req.StoryboardID, "frame_type", req.FrameType)
	return task.ID, nil
}

// processFramePromptGeneration 异步处理帧提示词生成
func (s *FramePromptService) processFramePromptGeneration(taskID string, req GenerateFramePromptRequest, model string) {
	// 更新任务状态为处理中
	s.taskService.UpdateTaskStatus(taskID, "processing", 0, "正在生成帧提示词...")

	// 查询分镜信息
	var storyboard models.Storyboard
	if err := s.db.Preload("Characters").Preload("Props").First(&storyboard, req.StoryboardID).Error; err != nil {
		s.log.Errorw("Storyboard not found during frame prompt generation", "error", err, "storyboard_id", req.StoryboardID)
		s.taskService.UpdateTaskStatus(taskID, "failed", 0, "分镜信息不存在")
		return
	}

	// 获取场景信息
	var scene *models.Scene
	if storyboard.SceneID != nil {
		scene = &models.Scene{}
		if err := s.db.First(scene, *storyboard.SceneID).Error; err != nil {
			s.log.Warnw("Scene not found during frame prompt generation", "scene_id", *storyboard.SceneID, "task_id", taskID)
			scene = nil
		}
	}

	// 获取 drama 的 style 信息
	var episode models.Episode
	if err := s.db.Preload("Drama").First(&episode, storyboard.EpisodeID).Error; err != nil {
		s.log.Warnw("Failed to load episode and drama", "error", err, "episode_id", storyboard.EpisodeID)
	}
	dramaStyle := episode.Drama.Style

	// 画面比例，默认 9:16 竖屏
	imageRatio := req.ImageRatio
	if imageRatio == "" {
		imageRatio = "9:16"
	}

	response := &FramePromptResponse{
		FrameType: req.FrameType,
	}

	// 生成提示词
	switch req.FrameType {
	case FrameTypeFirst:
		response.SingleFrame = s.generateFirstFrame(storyboard, scene, dramaStyle, model, imageRatio)
		s.saveFramePrompt(req.StoryboardID, string(req.FrameType), response.SingleFrame.Prompt, response.SingleFrame.Description, "")
	case FrameTypeKey:
		response.SingleFrame = s.generateKeyFrame(storyboard, scene, dramaStyle, model, imageRatio)
		s.saveFramePrompt(req.StoryboardID, string(req.FrameType), response.SingleFrame.Prompt, response.SingleFrame.Description, "")
	case FrameTypeLast:
		response.SingleFrame = s.generateLastFrame(storyboard, scene, dramaStyle, model, imageRatio)
		s.saveFramePrompt(req.StoryboardID, string(req.FrameType), response.SingleFrame.Prompt, response.SingleFrame.Description, "")
	case FrameTypePanel:
		count := req.PanelCount
		if count == 0 {
			count = 3
		}
		response.MultiFrame = s.generatePanelFrames(storyboard, scene, count, dramaStyle, model, imageRatio)
		var prompts []string
		for _, frame := range response.MultiFrame.Frames {
			prompts = append(prompts, frame.Prompt)
		}
		combinedPrompt := strings.Join(prompts, "\n---\n")
		s.saveFramePrompt(req.StoryboardID, string(req.FrameType), combinedPrompt, "分镜板组合提示词", response.MultiFrame.Layout)
	case FrameTypeAction:
		response.MultiFrame = s.generateActionSequence(storyboard, scene, dramaStyle, model, imageRatio)
		var prompts []string
		for _, frame := range response.MultiFrame.Frames {
			prompts = append(prompts, frame.Prompt)
		}
		combinedPrompt := strings.Join(prompts, "\n---\n")
		s.saveFramePrompt(req.StoryboardID, string(req.FrameType), combinedPrompt, "动作序列组合提示词", response.MultiFrame.Layout)
	default:
		s.log.Errorw("Unsupported frame type during frame prompt generation", "frame_type", req.FrameType, "task_id", taskID)
		s.taskService.UpdateTaskStatus(taskID, "failed", 0, "不支持的帧类型")
		return
	}

	// 更新任务状态为完成
	s.taskService.UpdateTaskResult(taskID, map[string]interface{}{
		"response":      response,
		"storyboard_id": req.StoryboardID,
		"frame_type":    string(req.FrameType),
	})

	s.log.Infow("Frame prompt generation completed", "task_id", taskID, "storyboard_id", req.StoryboardID, "frame_type", req.FrameType)
}

// BatchGenerateFramePrompts 批量为一个章节的所有分镜生成帧提示词
func (s *FramePromptService) BatchGenerateFramePrompts(episodeID string, frameType FrameType, model string, skipExisting bool) (string, error) {
	// 查询该章节的所有分镜
	var storyboards []models.Storyboard
	if err := s.db.Where("episode_id = ?", episodeID).Order("storyboard_number ASC, id ASC").Find(&storyboards).Error; err != nil {
		return "", fmt.Errorf("failed to load storyboards: %w", err)
	}

	if len(storyboards) == 0 {
		return "", fmt.Errorf("该章节没有分镜数据")
	}

	// 如果跳过已有的，过滤掉已有提示词的分镜
	if skipExisting {
		var filtered []models.Storyboard
		for _, sb := range storyboards {
			var count int64
			s.db.Model(&models.FramePrompt{}).
				Where("storyboard_id = ? AND frame_type = ?", sb.ID, string(frameType)).
				Count(&count)
			if count == 0 {
				filtered = append(filtered, sb)
			}
		}
		s.log.Infow("Batch frame prompt: skip existing",
			"total", len(storyboards), "to_generate", len(filtered), "skipped", len(storyboards)-len(filtered))
		storyboards = filtered

		if len(storyboards) == 0 {
			return "", fmt.Errorf("所有分镜已有该类型提示词，无需重新生成")
		}
	}

	// 创建一个总任务
	task, err := s.taskService.CreateTask("batch_frame_prompt_generation", episodeID)
	if err != nil {
		return "", fmt.Errorf("创建任务失败: %w", err)
	}

	// 异步批量处理
	go s.processBatchFramePromptGeneration(task.ID, storyboards, frameType, model)

	s.log.Infow("Batch frame prompt generation task created",
		"task_id", task.ID,
		"episode_id", episodeID,
		"storyboard_count", len(storyboards),
		"frame_type", frameType,
		"skip_existing", skipExisting)

	return task.ID, nil
}

// processBatchFramePromptGeneration 异步批量处理帧提示词生成
func (s *FramePromptService) processBatchFramePromptGeneration(taskID string, storyboards []models.Storyboard, frameType FrameType, model string) {
	total := len(storyboards)
	completed := 0
	failed := 0

	s.taskService.UpdateTaskStatus(taskID, "processing", 0, fmt.Sprintf("开始批量生成帧提示词，共 %d 个分镜...", total))

	for i, sb := range storyboards {
		storyboardID := fmt.Sprintf("%d", sb.ID)
		progress := i * 100 / total

		s.taskService.UpdateTaskStatus(taskID, "processing", progress,
			fmt.Sprintf("正在处理第 %d/%d 个分镜 (已完成: %d, 失败: %d)...", i+1, total, completed, failed))

		// 查询完整的分镜信息（包含角色和道具）
		var storyboard models.Storyboard
		if err := s.db.Preload("Characters").Preload("Props").First(&storyboard, sb.ID).Error; err != nil {
			s.log.Warnw("Failed to load storyboard for batch prompt generation",
				"storyboard_id", sb.ID, "error", err)
			failed++
			continue
		}

		// 获取场景信息
		var scene *models.Scene
		if storyboard.SceneID != nil {
			scene = &models.Scene{}
			if err := s.db.First(scene, *storyboard.SceneID).Error; err != nil {
				scene = nil
			}
		}

		// 获取 drama style
		var episode models.Episode
		if err := s.db.Preload("Drama").First(&episode, storyboard.EpisodeID).Error; err != nil {
			s.log.Warnw("Failed to load episode for batch prompt", "error", err)
		}
		dramaStyle := episode.Drama.Style

		// 批量生成默认使用竖屏 9:16
		batchRatio := "9:16"

		// 根据帧类型生成提示词
		switch frameType {
		case FrameTypeFirst:
			result := s.generateFirstFrame(storyboard, scene, dramaStyle, model, batchRatio)
			s.saveFramePrompt(storyboardID, string(frameType), result.Prompt, result.Description, "")
			completed++
		case FrameTypeKey:
			result := s.generateKeyFrame(storyboard, scene, dramaStyle, model, batchRatio)
			s.saveFramePrompt(storyboardID, string(frameType), result.Prompt, result.Description, "")
			completed++
		case FrameTypeLast:
			result := s.generateLastFrame(storyboard, scene, dramaStyle, model, batchRatio)
			s.saveFramePrompt(storyboardID, string(frameType), result.Prompt, result.Description, "")
			completed++
		case FrameTypePanel:
			result := s.generatePanelFrames(storyboard, scene, 3, dramaStyle, model, batchRatio)
			var prompts []string
			for _, frame := range result.Frames {
				prompts = append(prompts, frame.Prompt)
			}
			combinedPrompt := strings.Join(prompts, "\n---\n")
			s.saveFramePrompt(storyboardID, string(frameType), combinedPrompt, "分镜板组合提示词", result.Layout)
			completed++
		case FrameTypeAction:
			result := s.generateActionSequence(storyboard, scene, dramaStyle, model, batchRatio)
			var prompts []string
			for _, frame := range result.Frames {
				prompts = append(prompts, frame.Prompt)
			}
			combinedPrompt := strings.Join(prompts, "\n---\n")
			s.saveFramePrompt(storyboardID, string(frameType), combinedPrompt, "动作序列组合提示词", result.Layout)
			completed++
		default:
			s.log.Warnw("Unsupported frame type in batch generation", "frame_type", frameType)
			failed++
		}
	}

	// 更新任务为完成
	s.taskService.UpdateTaskResult(taskID, map[string]interface{}{
		"total":     total,
		"completed": completed,
		"failed":    failed,
	})

	s.log.Infow("Batch frame prompt generation completed",
		"task_id", taskID,
		"total", total,
		"completed", completed,
		"failed", failed)
}

// saveFramePrompt 保存帧提示词到数据库
func (s *FramePromptService) saveFramePrompt(storyboardID, frameType, prompt, description, layout string) {
	framePrompt := models.FramePrompt{
		StoryboardID: uint(mustParseUint(storyboardID)),
		FrameType:    frameType,
		Prompt:       prompt,
	}

	if description != "" {
		framePrompt.Description = &description
	}
	if layout != "" {
		framePrompt.Layout = &layout
	}

	// 先删除同类型的旧记录（保持最新）
	s.db.Where("storyboard_id = ? AND frame_type = ?", storyboardID, frameType).Delete(&models.FramePrompt{})

	// 插入新记录
	if err := s.db.Create(&framePrompt).Error; err != nil {
		s.log.Warnw("Failed to save frame prompt", "error", err, "storyboard_id", storyboardID, "frame_type", frameType)
	}
}

// mustParseUint 辅助函数
func mustParseUint(s string) uint64 {
	var result uint64
	fmt.Sscanf(s, "%d", &result)
	return result
}

// generateFirstFrame 生成首帧提示词
func (s *FramePromptService) generateFirstFrame(sb models.Storyboard, scene *models.Scene, dramaStyle string, model string, imageRatio string) *SingleFramePrompt {
	// 构建图片生成专用上下文（包含角色外观、场景描述，不含镜头运动语言）
	contextInfo := ""

	// 如果分镜已配置首帧画面描述，作为最高优先级放在最前面
	if sb.FirstFrameDesc != nil && *sb.FirstFrameDesc != "" {
		contextInfo = fmt.Sprintf("⚠️【首帧画面描述 —— 最高优先级，必须严格遵循其景别、构图和可见角色】\n%s\n\n", *sb.FirstFrameDesc)
	}

	contextInfo += s.buildImagePromptContext(sb, scene)

	// 使用国际化提示词
	systemPrompt := s.promptI18n.GetFirstFramePrompt(dramaStyle, imageRatio)
	userPrompt := s.promptI18n.FormatUserPrompt("frame_info", contextInfo)

	aiResponse, err := s.aiService.GenerateTextForModel(userPrompt, systemPrompt, model, "generate_first_frame", nil)
	if err != nil {
		s.log.Warnw("AI generation failed, using fallback", "error", err)
		fallbackPrompt := s.buildFallbackPrompt(sb, scene, "first frame, static shot")
		return &SingleFramePrompt{
			Prompt:      fallbackPrompt,
			Description: "镜头开始的静态画面，展示初始状态",
		}
	}

	// 解析AI返回的JSON
	result := s.parseFramePromptJSON(aiResponse)
	if result == nil {
		// JSON解析失败，使用降级方案
		s.log.Warnw("Failed to parse AI JSON response, using fallback", "storyboard_id", sb.ID, "response", aiResponse)
		fallbackPrompt := s.buildFallbackPrompt(sb, scene, "first frame, static shot")
		return &SingleFramePrompt{
			Prompt:      fallbackPrompt,
			Description: "镜头开始的静态画面，展示初始状态",
		}
	}

	return result
}

// generateKeyFrame 生成关键帧提示词
func (s *FramePromptService) generateKeyFrame(sb models.Storyboard, scene *models.Scene, dramaStyle string, model string, imageRatio string) *SingleFramePrompt {
	// 构建图片生成专用上下文
	contextInfo := s.buildImagePromptContext(sb, scene)

	// 使用国际化提示词
	systemPrompt := s.promptI18n.GetKeyFramePrompt(dramaStyle, imageRatio)
	userPrompt := s.promptI18n.FormatUserPrompt("key_frame_info", contextInfo)

	aiResponse, err := s.aiService.GenerateTextForModel(userPrompt, systemPrompt, model, "generate_key_frame", nil)
	if err != nil {
		s.log.Warnw("AI generation failed, using fallback", "error", err)
		fallbackPrompt := s.buildFallbackPrompt(sb, scene, "key frame, dynamic action")
		return &SingleFramePrompt{
			Prompt:      fallbackPrompt,
			Description: "动作高潮瞬间，展示关键动作",
		}
	}

	// 解析AI返回的JSON
	result := s.parseFramePromptJSON(aiResponse)
	if result == nil {
		// JSON解析失败，使用降级方案
		s.log.Warnw("Failed to parse AI JSON response, using fallback", "storyboard_id", sb.ID, "response", aiResponse)
		fallbackPrompt := s.buildFallbackPrompt(sb, scene, "key frame, dynamic action")
		return &SingleFramePrompt{
			Prompt:      fallbackPrompt,
			Description: "动作高潮瞬间，展示关键动作",
		}
	}

	return result
}

// generateLastFrame 生成尾帧提示词
func (s *FramePromptService) generateLastFrame(sb models.Storyboard, scene *models.Scene, dramaStyle string, model string, imageRatio string) *SingleFramePrompt {
	// 构建图片生成专用上下文
	contextInfo := s.buildImagePromptContext(sb, scene)

	// 如果分镜已配置尾帧画面描述，注入上下文
	if sb.LastFrameDesc != nil && *sb.LastFrameDesc != "" {
		contextInfo += fmt.Sprintf("\n\n【尾帧画面描述（重点参考）】\n%s", *sb.LastFrameDesc)
	}

	// 使用国际化提示词
	systemPrompt := s.promptI18n.GetLastFramePrompt(dramaStyle, imageRatio)

	// 查找该分镜的首帧提示词，用于保持角色位置一致性
	var firstFramePrompt models.FramePrompt
	var userPrompt string
	err := s.db.Where("storyboard_id = ? AND frame_type = ?", sb.ID, "first").
		Order("updated_at DESC").First(&firstFramePrompt).Error
	if err == nil && firstFramePrompt.Prompt != "" {
		// 有首帧提示词，使用带首帧上下文的模板
		userPrompt = s.promptI18n.FormatUserPrompt("last_frame_info_with_first", contextInfo, firstFramePrompt.Prompt)
		s.log.Infow("Using first frame prompt as context for last frame generation",
			"storyboard_id", sb.ID, "first_frame_prompt_length", len(firstFramePrompt.Prompt))
	} else {
		// 没有首帧提示词，使用普通模板
		userPrompt = s.promptI18n.FormatUserPrompt("last_frame_info", contextInfo)
		s.log.Warnw("No first frame prompt found for last frame generation, using standard template",
			"storyboard_id", sb.ID)
	}

	aiResponse, genErr := s.aiService.GenerateTextForModel(userPrompt, systemPrompt, model, "generate_last_frame", nil)
	if genErr != nil {
		s.log.Warnw("AI generation failed, using fallback", "error", genErr)
		fallbackPrompt := s.buildFallbackPrompt(sb, scene, "last frame, final state")
		return &SingleFramePrompt{
			Prompt:      fallbackPrompt,
			Description: "镜头结束画面，展示最终状态和结果",
		}
	}

	// 解析AI返回的JSON
	result := s.parseFramePromptJSON(aiResponse)
	if result == nil {
		// JSON解析失败，使用降级方案
		s.log.Warnw("Failed to parse AI JSON response, using fallback", "storyboard_id", sb.ID, "response", aiResponse)
		fallbackPrompt := s.buildFallbackPrompt(sb, scene, "last frame, final state")
		return &SingleFramePrompt{
			Prompt:      fallbackPrompt,
			Description: "镜头结束画面，展示最终状态和结果",
		}
	}

	return result
}

// generatePanelFrames 生成分镜板提示词（多格组合）
func (s *FramePromptService) generatePanelFrames(sb models.Storyboard, scene *models.Scene, count int, dramaStyle string, model string, imageRatio string) *MultiFramePrompt {
	layout := fmt.Sprintf("horizontal_%d", count)

	frames := make([]SingleFramePrompt, count)

	// 固定生成：首帧 -> 关键帧 -> 尾帧
	if count == 3 {
		frames[0] = *s.generateFirstFrame(sb, scene, dramaStyle, model, imageRatio)
		frames[0].Description = "第1格：初始状态"

		frames[1] = *s.generateKeyFrame(sb, scene, dramaStyle, model, imageRatio)
		frames[1].Description = "第2格：动作高潮"

		frames[2] = *s.generateLastFrame(sb, scene, dramaStyle, model, imageRatio)
		frames[2].Description = "第3格：最终状态"
	} else if count == 4 {
		frames[0] = *s.generateFirstFrame(sb, scene, dramaStyle, model, imageRatio)
		frames[1] = *s.generateKeyFrame(sb, scene, dramaStyle, model, imageRatio)
		frames[2] = *s.generateKeyFrame(sb, scene, dramaStyle, model, imageRatio)
		frames[3] = *s.generateLastFrame(sb, scene, dramaStyle, model, imageRatio)
	}

	return &MultiFramePrompt{
		Layout: layout,
		Frames: frames,
	}
}

// generateActionSequence 生成关键帧提示词（2×2四宫格）
func (s *FramePromptService) generateActionSequence(sb models.Storyboard, scene *models.Scene, dramaStyle string, model string, imageRatio string) *MultiFramePrompt {
	contextInfo := s.buildImagePromptContext(sb, scene)

	systemPrompt := s.promptI18n.GetActionSequenceFramePrompt(dramaStyle)
	userPrompt := s.promptI18n.FormatUserPrompt("frame_info", contextInfo)

	aiResponse, err := s.aiService.GenerateTextForModel(userPrompt, systemPrompt, model, "generate_action_sequence", nil)
	if err != nil {
		s.log.Warnw("AI generation failed for action sequence, using fallback", "error", err)
		fallbackPrompt := s.buildFallbackPrompt(sb, scene, "2x2 keyframe grid, progressive storyline, character consistency")
		return &MultiFramePrompt{
			Layout: "grid_2x2",
			Frames: []SingleFramePrompt{
				{
					Prompt:      fallbackPrompt,
					Description: "2×2四宫格关键帧，展示剧情递进",
				},
			},
		}
	}

	result := s.parseFramePromptJSON(aiResponse)
	if result == nil {
		s.log.Warnw("Failed to parse AI JSON response for action sequence, using fallback", "storyboard_id", sb.ID, "response", aiResponse)
		fallbackPrompt := s.buildFallbackPrompt(sb, scene, "2x2 keyframe grid, progressive storyline, character consistency")
		return &MultiFramePrompt{
			Layout: "grid_2x2",
			Frames: []SingleFramePrompt{
				{
					Prompt:      fallbackPrompt,
					Description: "2×2四宫格关键帧，展示剧情递进",
				},
			},
		}
	}

	return &MultiFramePrompt{
		Layout: "grid_2x2",
		Frames: []SingleFramePrompt{*result},
	}
}

// buildStoryboardContext 构建镜头上下文信息（用于视频提示词等，保留镜头运动语言）
func (s *FramePromptService) buildStoryboardContext(sb models.Storyboard, scene *models.Scene) string {
	var parts []string

	// 镜头描述（最重要）
	if sb.Description != nil && *sb.Description != "" {
		parts = append(parts, s.promptI18n.FormatUserPrompt("shot_description_label", *sb.Description))
	}

	// 场景信息
	if scene != nil {
		parts = append(parts, s.promptI18n.FormatUserPrompt("scene_label", scene.Location, scene.Time))
	} else if sb.Location != nil && sb.Time != nil {
		parts = append(parts, s.promptI18n.FormatUserPrompt("scene_label", *sb.Location, *sb.Time))
	}

	// 角色
	if len(sb.Characters) > 0 {
		var charNames []string
		for _, char := range sb.Characters {
			charNames = append(charNames, char.Name)
		}
		parts = append(parts, s.promptI18n.FormatUserPrompt("characters_label", strings.Join(charNames, ", ")))
	}

	// 动作
	if sb.Action != nil && *sb.Action != "" {
		parts = append(parts, s.promptI18n.FormatUserPrompt("action_label", *sb.Action))
	}

	// 结果
	if sb.Result != nil && *sb.Result != "" {
		parts = append(parts, s.promptI18n.FormatUserPrompt("result_label", *sb.Result))
	}

	// 对白
	if sb.Dialogue != nil && *sb.Dialogue != "" {
		parts = append(parts, s.promptI18n.FormatUserPrompt("dialogue_label", *sb.Dialogue))
	}

	// 氛围
	if sb.Atmosphere != nil && *sb.Atmosphere != "" {
		parts = append(parts, s.promptI18n.FormatUserPrompt("atmosphere_label", *sb.Atmosphere))
	}

	// 镜头参数
	if sb.ShotType != nil {
		parts = append(parts, s.promptI18n.FormatUserPrompt("shot_type_label", *sb.ShotType))
	}
	if sb.Angle != nil {
		parts = append(parts, s.promptI18n.FormatUserPrompt("angle_label", *sb.Angle))
	}
	if sb.Movement != nil {
		parts = append(parts, s.promptI18n.FormatUserPrompt("movement_label", *sb.Movement))
	}

	return strings.Join(parts, "\n")
}

// buildImagePromptContext 构建图片生成专用上下文（包含角色外观、场景描述，不含镜头运动语言）
func (s *FramePromptService) buildImagePromptContext(sb models.Storyboard, scene *models.Scene) string {
	var parts []string

	// 镜头描述（核心内容，分镜合并后的完整描述）
	shotDesc := ""
	if sb.Action != nil && *sb.Action != "" {
		shotDesc = *sb.Action
	}
	if shotDesc != "" {
		parts = append(parts, fmt.Sprintf("【镜头描述】\n%s", shotDesc))
	}

	// 场景环境（只用 Description 字段，Prompt 字段是图片生成专用提示词，不传给帧 AI）
	if scene != nil {
		sceneDesc := fmt.Sprintf("【场景环境】地点：%s，时间：%s", scene.Location, scene.Time)
		if scene.Description != nil && *scene.Description != "" {
			sceneDesc += fmt.Sprintf("，环境描写：%s", *scene.Description)
		}
		if scene.Atmosphere != nil && *scene.Atmosphere != "" {
			sceneDesc += fmt.Sprintf("，氛围：%s", *scene.Atmosphere)
		}
		if scene.Lighting != nil && *scene.Lighting != "" {
			sceneDesc += fmt.Sprintf("，光线：%s", *scene.Lighting)
		}
		parts = append(parts, sceneDesc)
	} else if sb.Location != nil && sb.Time != nil {
		parts = append(parts, fmt.Sprintf("【场景环境】地点：%s，时间：%s", *sb.Location, *sb.Time))
	}

	// 出场角色（所有角色都传外貌信息，用于在提示词中替代角色名字）
	if len(sb.Characters) > 0 {
		var charLines []string
		for _, char := range sb.Characters {
			var info strings.Builder
			info.WriteString(fmt.Sprintf("- 角色名：%s", char.Name))
			if char.Gender != nil && *char.Gender != "" {
				info.WriteString(fmt.Sprintf("，性别：%s", *char.Gender))
			}
			if char.AgeDescription != nil && *char.AgeDescription != "" {
				info.WriteString(fmt.Sprintf("，年龄：%s", *char.AgeDescription))
			}
			if char.Appearance != nil && *char.Appearance != "" {
				info.WriteString(fmt.Sprintf("，外貌：%s", *char.Appearance))
			}
			if char.Description != nil && *char.Description != "" {
				info.WriteString(fmt.Sprintf("，身份：%s", *char.Description))
			}
			charLines = append(charLines, info.String())
		}
		header := "【出场角色】\n⚠️ 提示词中禁止使用角色名字，必须用外貌特征指代（如「穿灰衬衫的短发男性」）\n"
		parts = append(parts, header+strings.Join(charLines, "\n"))
	}

	// 道具
	if len(sb.Props) > 0 {
		var propLines []string
		for _, prop := range sb.Props {
			desc := fmt.Sprintf("- %s", prop.Name)
			if prop.Description != nil && *prop.Description != "" {
				desc += fmt.Sprintf("：%s", *prop.Description)
			}
			propLines = append(propLines, desc)
		}
		parts = append(parts, "【道具】\n"+strings.Join(propLines, "\n"))
	}

	return strings.Join(parts, "\n\n")
}

// buildFallbackPrompt 构建降级提示词（AI失败时使用）
func (s *FramePromptService) buildFallbackPrompt(sb models.Storyboard, scene *models.Scene, suffix string) string {
	var parts []string

	// 场景
	if scene != nil {
		parts = append(parts, fmt.Sprintf("%s, %s", scene.Location, scene.Time))
	}

	// 角色
	if len(sb.Characters) > 0 {
		for _, char := range sb.Characters {
			parts = append(parts, char.Name)
		}
	}

	// 氛围
	if sb.Atmosphere != nil {
		parts = append(parts, *sb.Atmosphere)
	}

	parts = append(parts, "anime style", suffix)
	return strings.Join(parts, ", ")
}
