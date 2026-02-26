package services

import (
	"strconv"

	"fmt"
	"strings"

	models "github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/ai"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StoryboardService struct {
	db          *gorm.DB
	aiService   *AIService
	taskService *TaskService
	log         *logger.Logger
	config      *config.Config
	promptI18n  *PromptI18n
}

func NewStoryboardService(db *gorm.DB, cfg *config.Config, log *logger.Logger) *StoryboardService {
	return &StoryboardService{
		db:          db,
		aiService:   NewAIService(db, log),
		taskService: NewTaskService(db, log),
		log:         log,
		config:      cfg,
		promptI18n:  NewPromptI18n(cfg),
	}
}

type Storyboard struct {
	ShotNumber  int    `json:"shot_number"`
	Title       string `json:"title"`         // 镜头标题
	ShotType    string `json:"shot_type"`     // 景别
	Angle       string `json:"angle"`         // 镜头角度
	Time        string `json:"time"`          // 时间
	Location    string `json:"location"`      // 地点
	SceneID     *uint  `json:"scene_id"`      // 背景ID（AI直接返回，可为null）
	Movement    string `json:"movement"`      // 运镜
	Description    string `json:"description"`      // 镜头描述（完整的一段话）
	Action         string `json:"action"`           // 动作（兼容旧格式）
	Dialogue       string `json:"dialogue"`         // 对话/独白
	Result         string `json:"result"`           // 画面结果
	PrevShotRef      string `json:"prev_shot_ref"`      // 承接上镜描述
	FirstFrameDesc   string `json:"first_frame_desc"`   // 首帧画面描述
	MiddleActionDesc string `json:"middle_action_desc"` // 中间过程描述
	LastFrameDesc    string `json:"last_frame_desc"`    // 尾帧画面描述
	Atmosphere  string `json:"atmosphere"`    // 环境氛围
	Emotion     string `json:"emotion"`       // 情绪
	Duration    int    `json:"duration"`      // 时长（秒）
	BgmPrompt   string `json:"bgm_prompt"`    // 配乐提示词
	SoundEffect string `json:"sound_effect"`  // 音效描述
	VideoPrompt string `json:"video_prompt"`  // Seedance视频生成提示词
	Characters  []uint `json:"characters"`    // 涉及的角色ID列表
	Props       []uint `json:"props"`         // 涉及的道具ID列表
	IsPrimary   bool   `json:"is_primary"`    // 是否主镜
}

type GenerateStoryboardResult struct {
	Storyboards []Storyboard `json:"storyboards"`
	Total       int          `json:"total"`
}

func (s *StoryboardService) GenerateStoryboard(episodeID string, model string, shotCount int) (string, error) {
	s.log.Infow("GenerateStoryboard called", "episode_id", episodeID, "model", model, "shot_count", shotCount)

	// 默认镜头数量
	if shotCount <= 0 {
		shotCount = 1 // V3: 默认1个镜头
	}
	if shotCount < 1 {
		shotCount = 1
	}
	if shotCount > 50 {
		shotCount = 50
	}

	s.log.Infow("Final shot count", "episode_id", episodeID, "shot_count", shotCount)
	// 从数据库获取剧集信息（含 drama.style）
	var episode struct {
		ID            string
		ScriptContent *string
		Description   *string
		DramaID       string
		DramaStyle    string
	}

	err := s.db.Table("episodes").
		Select("episodes.id, episodes.script_content, episodes.description, episodes.drama_id, dramas.style as drama_style").
		Joins("INNER JOIN dramas ON dramas.id = episodes.drama_id").
		Where("episodes.id = ?", episodeID).
		First(&episode).Error

	if err != nil {
		return "", fmt.Errorf("剧集不存在或无权限访问")
	}

	// 获取剧本内容
	var scriptContent string
	if episode.ScriptContent != nil && *episode.ScriptContent != "" {
		scriptContent = *episode.ScriptContent
	} else if episode.Description != nil && *episode.Description != "" {
		scriptContent = *episode.Description
	} else {
		return "", fmt.Errorf("剧本内容为空，请先生成剧集内容")
	}

	// 获取该剧本的所有角色
	var characters []models.Character
	if err := s.db.Where("drama_id = ?", episode.DramaID).Order("name ASC").Find(&characters).Error; err != nil {
		return "", fmt.Errorf("获取角色列表失败: %w", err)
	}

	// 构建角色列表字符串（包含ID和名称）
	characterList := "无角色"
	if len(characters) > 0 {
		var charInfoList []string
		for _, char := range characters {
			charInfoList = append(charInfoList, fmt.Sprintf(`{"id": %d, "name": "%s"}`, char.ID, char.Name))
		}
		characterList = fmt.Sprintf("[%s]", strings.Join(charInfoList, ", "))
	}

	// 获取该项目已提取的场景列表（项目级）
	var scenes []models.Scene
	if err := s.db.Where("drama_id = ?", episode.DramaID).Order("location ASC, time ASC").Find(&scenes).Error; err != nil {
		s.log.Warnw("Failed to get scenes", "error", err)
	}

	// 构建场景列表字符串（包含ID、地点、时间）
	sceneList := "无场景"
	if len(scenes) > 0 {
		var sceneInfoList []string
		for _, bg := range scenes {
			sceneInfoList = append(sceneInfoList, fmt.Sprintf(`{"id": %d, "location": "%s", "time": "%s"}`, bg.ID, bg.Location, bg.Time))
		}
		sceneList = fmt.Sprintf("[%s]", strings.Join(sceneInfoList, ", "))
	}

	// 获取该项目的道具列表
	var props []models.Prop
	if err := s.db.Where("drama_id = ?", episode.DramaID).Order("name ASC").Find(&props).Error; err != nil {
		s.log.Warnw("Failed to get props", "error", err)
	}

	// 构建道具列表字符串（包含ID、名称、类型、描述）
	propList := "无道具"
	if len(props) > 0 {
		var propInfoList []string
		for _, p := range props {
			desc := ""
			if p.Description != nil {
				desc = *p.Description
			}
			propType := ""
			if p.Type != nil {
				propType = *p.Type
			}
			propInfoList = append(propInfoList, fmt.Sprintf(`{"id": %d, "name": "%s", "type": "%s", "description": "%s"}`, p.ID, p.Name, propType, desc))
		}
		propList = fmt.Sprintf("[%s]", strings.Join(propInfoList, ", "))
	}

	// 使用国际化提示词
	systemPrompt := s.promptI18n.GetStoryboardSystemPrompt()

	// 获取风格描述
	styleLabel := "写实摄影风格"
	if episode.DramaStyle == models.StyleComic {
		styleLabel = "漫画/动漫风格"
	}

	userPrompt := fmt.Sprintf(`【画面风格】%s

【角色列表】
%s

【场景列表】
%s

【道具列表】
%s

【剧本】
%s

【分镜要求】
- ⚠️ 如果剧本已经是三段式格式（包含"首帧：""过程：""尾帧："），则严格按照剧本的镜头划分，直接提取每个镜头的三段描述，不要重新拆分或合并
- 如果剧本不是三段式格式，则按以下规则拆分为约 %d 个镜头（允许±20%%浮动）
- 不得遗漏剧本中的任何对话和关键动作
- duration 范围 4-12秒，所有分镜时长之和 = 剧集总时长
- ⚠️ 三段描述中绝对禁止出现角色名字，只能用穿着特征指代
- middle_action_desc 必须包含该镜头的所有台词对话（用「」括起）
- last_frame_desc 只描述相对于首帧的变化
- 所有镜头默认 is_primary: true

【ID匹配】
- characters: 根据剧本中的"造型：""角色："字段，匹配角色列表中的 id（数字数组），不出现则 []
- scene_id: 根据剧本中的"场景："字段，匹配场景列表中 location 最接近的 id（数字），无合适填 null
- props: 根据剧本中的"道具："字段，匹配道具列表中的 id（数字数组），不出现则 []

【输出格式】严格 JSON：
{
  "storyboards": [
    {
      "shot_number": 1,
      "title": "闹钟赖床",
      "time": "清晨",
      "location": "家中卧室",
      "scene_id": 168,
      "prev_shot_ref": "",
      "first_frame_desc": "近景，晨光微弱的卧室，穿浅粉色圆领睡衣（胸口兔子刺绣）的深棕色蓬松短卷发女孩蜷缩在被窝中...",
      "middle_action_desc": "穿粉色睡衣的短卷发女孩皱眉从被窝中缓缓伸出一只手摸索床头，粉色薄唇微张嘟囔：「周一……再躺五分钟。」...",
      "last_frame_desc": "穿粉色睡衣的女孩整个人缩进被窝，只露出蓬松凌乱的短卷发顶，表情从微皱变为放松...",
      "duration": 8,
      "characters": [129],
      "props": [25],
      "is_primary": true
    },
    {
      "shot_number": 2,
      "title": "起床拖延",
      "time": "清晨",
      "location": "家中卧室",
      "scene_id": 168,
      "prev_shot_ref": "同场景，机位从近景拉至中景，女孩已从被窝中坐起",
      "first_frame_desc": "中景，同一卧室晨光渐亮，穿浅粉色睡衣的短卷发女孩已掀开被子半坐在床边，双脚垂在床沿，眼睛半睁，发丝凌乱...",
      "middle_action_desc": "...",
      "last_frame_desc": "...",
      "duration": 6,
      "characters": [129],
      "props": [],
      "is_primary": true
    }
  ]
}`, styleLabel, characterList, sceneList, propList, scriptContent,
		shotCount)

	// 创建异步任务
	task, err := s.taskService.CreateTask("storyboard_generation", episodeID)
	if err != nil {
		s.log.Errorw("Failed to create task", "error", err)
		return "", fmt.Errorf("创建任务失败: %w", err)
	}

	s.log.Infow("Generating storyboard asynchronously",
		"task_id", task.ID,
		"episode_id", episodeID,
		"drama_id", episode.DramaID,
		"script_length", len(scriptContent),
		"character_count", len(characters),
		"characters", characterList,
		"scene_count", len(scenes),
		"scenes", sceneList)

	// 启动后台goroutine处理AI调用和后续逻辑
	go s.processStoryboardGeneration(task.ID, episodeID, model, systemPrompt, userPrompt)

	// 立即返回任务ID
	return task.ID, nil
}

// processStoryboardGeneration 后台处理故事板生成
func (s *StoryboardService) processStoryboardGeneration(taskID, episodeID, model, systemPrompt, prompt string) {
	// 更新任务状态为处理中
	if err := s.taskService.UpdateTaskStatus(taskID, "processing", 10, "开始生成分镜头..."); err != nil {
		s.log.Errorw("Failed to update task status", "error", err, "task_id", taskID)
		return
	}

	s.log.Infow("Processing storyboard generation", "task_id", taskID, "episode_id", episodeID)

	// 调用AI服务生成（如果指定了模型则使用指定的模型）
	// 根据模型动态设置 max_tokens，确保完整返回所有分镜的JSON
	maxTokens := getMaxTokensForModel(model)
	s.log.Infow("Storyboard generation max_tokens", "model", model, "max_tokens", maxTokens, "task_id", taskID)

	var text string
	var err error
	text, err = s.aiService.GenerateTextForModel(prompt, systemPrompt, model, "generate_storyboard", nil, ai.WithMaxTokens(maxTokens))

	if err != nil {
		s.log.Errorw("Failed to generate storyboard", "error", err, "task_id", taskID)
		if updateErr := s.taskService.UpdateTaskError(taskID, fmt.Errorf("生成分镜头失败: %w", err)); updateErr != nil {
			s.log.Errorw("Failed to update task error", "error", updateErr, "task_id", taskID)
		}
		return
	}

	// 更新任务进度
	if err := s.taskService.UpdateTaskStatus(taskID, "processing", 50, "分镜头生成完成，正在解析结果..."); err != nil {
		s.log.Errorw("Failed to update task status", "error", err, "task_id", taskID)
		return
	}

	// 解析JSON结果
	// AI可能返回两种格式：
	// 1. 数组格式: [{...}, {...}]
	// 2. 对象格式: {"storyboards": [{...}, {...}]}
	var result GenerateStoryboardResult

	// 先尝试解析为数组格式
	var storyboards []Storyboard
	if err := utils.SafeParseAIJSON(text, &storyboards); err == nil {
		// 成功解析为数组，包装为对象
		result.Storyboards = storyboards
		result.Total = len(storyboards)
		s.log.Infow("Parsed storyboard as array format", "count", len(storyboards), "task_id", taskID)
	} else {
		// 尝试解析为对象格式
		if err := utils.SafeParseAIJSON(text, &result); err != nil {
			s.log.Errorw("Failed to parse storyboard JSON in both formats", "error", err, "response", text[:min(500, len(text))], "task_id", taskID)
			if updateErr := s.taskService.UpdateTaskError(taskID, fmt.Errorf("解析分镜头结果失败: %w", err)); updateErr != nil {
				s.log.Errorw("Failed to update task error", "error", updateErr, "task_id", taskID)
			}
			return
		}
		result.Total = len(result.Storyboards)
		s.log.Infow("Parsed storyboard as object format", "count", len(result.Storyboards), "task_id", taskID)
	}

	// 计算总时长（所有分镜时长之和）
	totalDuration := 0
	for _, sb := range result.Storyboards {
		totalDuration += sb.Duration
	}

	s.log.Infow("Storyboard generated",
		"task_id", taskID,
		"episode_id", episodeID,
		"count", result.Total,
		"total_duration_seconds", totalDuration)

	// 更新任务进度
	if err := s.taskService.UpdateTaskStatus(taskID, "processing", 70, "正在保存分镜头..."); err != nil {
		s.log.Errorw("Failed to update task status", "error", err, "task_id", taskID)
		return
	}

	// 保存分镜头到数据库
	if err := s.saveStoryboards(episodeID, result.Storyboards); err != nil {
		s.log.Errorw("Failed to save storyboards", "error", err, "task_id", taskID)
		if updateErr := s.taskService.UpdateTaskError(taskID, fmt.Errorf("保存分镜头失败: %w", err)); updateErr != nil {
			s.log.Errorw("Failed to update task error", "error", updateErr, "task_id", taskID)
		}
		return
	}

	// 更新剧集时长（秒转分钟，向上取整）
	if err := s.taskService.UpdateTaskStatus(taskID, "processing", 85, "正在更新剧集时长..."); err != nil {
		s.log.Errorw("Failed to update task status", "error", err, "task_id", taskID)
		return
	}

	durationMinutes := (totalDuration + 59) / 60
	if err := s.db.Model(&models.Episode{}).Where("id = ?", episodeID).Update("duration", durationMinutes).Error; err != nil {
		s.log.Errorw("Failed to update episode duration", "error", err, "task_id", taskID)
	} else {
		s.log.Infow("Episode duration updated",
			"task_id", taskID,
			"episode_id", episodeID,
			"duration_seconds", totalDuration,
			"duration_minutes", durationMinutes)
	}

	// 更新任务结果
	resultData := gin.H{
		"storyboards":      result.Storyboards,
		"total":            result.Total,
		"total_duration":   totalDuration,
		"duration_minutes": durationMinutes,
	}

	if err := s.taskService.UpdateTaskResult(taskID, resultData); err != nil {
		s.log.Errorw("Failed to update task result", "error", err, "task_id", taskID)
		return
	}

	s.log.Infow("Storyboard generation completed", "task_id", taskID, "episode_id", episodeID)
}

// generateImagePrompt 生成专门用于图片生成的提示词（首帧静态画面）
func (s *StoryboardService) generateImagePrompt(sb Storyboard, style string) string {
	var parts []string

	// 1. 完整的场景背景描述
	if sb.Location != "" {
		locationDesc := sb.Location
		if sb.Time != "" {
			locationDesc += ", " + sb.Time
		}
		parts = append(parts, locationDesc)
	}

	// 2. 角色初始静态姿态（去除动作过程，只保留起始状态）
	if sb.Action != "" {
		initialPose := extractInitialPose(sb.Action)
		if initialPose != "" {
			parts = append(parts, initialPose)
		}
	}

	// 3. 情绪氛围
	if sb.Emotion != "" {
		parts = append(parts, sb.Emotion)
	}

	// 4. 风格标签
	if style == models.StyleComic {
		parts = append(parts, "anime style, first frame")
	} else {
		parts = append(parts, "photorealistic style, first frame")
	}

	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}
	if style == models.StyleComic {
		return "anime scene"
	}
	return "photorealistic scene"
}

// extractInitialPose 提取初始静态姿态（去除动作过程）
func extractInitialPose(action string) string {
	// 去除动作过程关键词，保留初始状态描述
	processWords := []string{
		"然后", "接着", "接下来", "随后", "紧接着",
		"向下", "向上", "向前", "向后", "向左", "向右",
		"开始", "继续", "逐渐", "慢慢", "快速", "突然", "猛然",
	}

	result := action
	for _, word := range processWords {
		if idx := strings.Index(result, word); idx > 0 {
			// 在动作过程词之前截断
			result = result[:idx]
			break
		}
	}

	// 清理末尾标点
	result = strings.TrimRight(result, "，。,. ")
	return strings.TrimSpace(result)
}

// extractSimpleLocation 提取简化的场景地点（去除详细描述）
func extractSimpleLocation(location string) string {
	// 在"·"符号处截断，只保留主场景名称
	if idx := strings.Index(location, "·"); idx > 0 {
		return strings.TrimSpace(location[:idx])
	}

	// 如果有逗号，只保留第一部分
	if idx := strings.Index(location, "，"); idx > 0 {
		return strings.TrimSpace(location[:idx])
	}
	if idx := strings.Index(location, ","); idx > 0 {
		return strings.TrimSpace(location[:idx])
	}

	// 限制长度不超过15个字符
	maxLen := 15
	if len(location) > maxLen {
		return strings.TrimSpace(location[:maxLen])
	}

	return strings.TrimSpace(location)
}

// extractSimplePose 提取简单的核心姿态关键词（不超过10个字）
func extractSimplePose(action string) string {
	// 只提取前面最多10个字符作为核心姿态
	runes := []rune(action)
	maxLen := 10
	if len(runes) > maxLen {
		// 在标点符号处截断
		truncated := runes[:maxLen]
		for i := maxLen - 1; i >= 0; i-- {
			if truncated[i] == '，' || truncated[i] == '。' || truncated[i] == ',' || truncated[i] == '.' {
				truncated = runes[:i]
				break
			}
		}
		return strings.TrimSpace(string(truncated))
	}
	return strings.TrimSpace(action)
}

// extractFirstFramePose 从动作描述中提取首帧静态姿态
func extractFirstFramePose(action string) string {
	// 去除表示动作过程的关键词，保留初始状态
	processWords := []string{
		"然后", "接着", "向下", "向前", "走向", "冲向", "转身",
		"开始", "继续", "逐渐", "慢慢", "快速", "突然",
	}

	pose := action
	for _, word := range processWords {
		// 简单处理：在这些词之前截断
		if idx := strings.Index(pose, word); idx > 0 {
			pose = pose[:idx]
			break
		}
	}

	// 清理末尾标点
	pose = strings.TrimRight(pose, "，。,.")
	return strings.TrimSpace(pose)
}

// extractCompositionType 从镜头类型中提取构图类型（去除运镜）
func extractCompositionType(shotType string) string {
	// 去除运镜相关描述
	cameraMovements := []string{
		"晃动", "摇晃", "推进", "拉远", "跟随", "环绕",
		"运镜", "摄影", "移动", "旋转",
	}

	comp := shotType
	for _, movement := range cameraMovements {
		comp = strings.ReplaceAll(comp, movement, "")
	}

	// 清理多余的标点和空格
	comp = strings.ReplaceAll(comp, "··", "·")
	comp = strings.ReplaceAll(comp, "·", " ")
	comp = strings.TrimSpace(comp)

	return comp
}

// generateVideoPrompt 生成 Seedance 1.5 Pro 格式的视频提示词（回退方案）
// 公式：主体 + 运动 + 环境 + 运镜 + 美学 + 声音
func (s *StoryboardService) generateVideoPrompt(sb Storyboard) string {
	var sentences []string

	// 1. 运镜/景别 + 环境（开头）
	var opening string
	if sb.ShotType != "" && sb.Angle != "" {
		opening = fmt.Sprintf("%s，%s", sb.Angle, sb.ShotType)
	} else if sb.ShotType != "" {
		opening = sb.ShotType
	} else if sb.Angle != "" {
		opening = sb.Angle
	}

	var env string
	if sb.Time != "" && sb.Location != "" {
		env = fmt.Sprintf("%s，%s", sb.Time, sb.Location)
	} else if sb.Location != "" {
		env = sb.Location
	} else if sb.Time != "" {
		env = sb.Time
	}

	if opening != "" && env != "" {
		sentences = append(sentences, fmt.Sprintf("%s，%s", opening, env))
	} else if opening != "" {
		sentences = append(sentences, opening)
	} else if env != "" {
		sentences = append(sentences, env)
	}

	// 2. 主体 + 运动（动作描述 — 核心）
	if sb.Action != "" {
		sentences = append(sentences, sb.Action)
	}

	// 3. 对话（直接写出台词）
	if sb.Dialogue != "" {
		sentences = append(sentences, sb.Dialogue)
	}

	// 4. 运镜（融入描述）
	if sb.Movement != "" && sb.Movement != "固定镜头" {
		movementMap := map[string]string{
			"推": "镜头推近", "拉": "镜头拉远", "摇": "镜头摇移",
			"移": "镜头横移", "跟": "镜头跟随", "升": "镜头上升",
			"降": "镜头下降", "甩": "镜头快速甩过", "环绕": "镜头环绕",
			"旋转": "镜头旋转", "变焦": "镜头变焦",
		}
		if mapped, ok := movementMap[sb.Movement]; ok {
			sentences = append(sentences, mapped)
		} else {
			sentences = append(sentences, fmt.Sprintf("镜头%s", sb.Movement))
		}
	}

	// 5. 画面结果
	if sb.Result != "" {
		sentences = append(sentences, sb.Result)
	}

	// 6. 氛围/美学
	if sb.Atmosphere != "" {
		sentences = append(sentences, sb.Atmosphere)
	}

	// 7. 声音设计
	if sb.BgmPrompt != "" {
		sentences = append(sentences, fmt.Sprintf("背景音乐是%s", sb.BgmPrompt))
	}
	if sb.SoundEffect != "" {
		sentences = append(sentences, sb.SoundEffect)
	}

	if len(sentences) > 0 {
		return strings.Join(sentences, "。") + "。"
	}
	return "视频场景"
}

// VideoPromptResult 阶段二AI返回的单个分镜视频提示词结果
type VideoPromptResult struct {
	ShotNumber  int    `json:"shot_number"`
	ShotType    string `json:"shot_type"`
	Angle       string `json:"angle"`
	Movement    string `json:"movement"`
	VideoPrompt string `json:"video_prompt"`
	BgmPrompt   string `json:"bgm_prompt"`
	SoundEffect string `json:"sound_effect"`
}

// VideoPromptBatchResult 阶段二AI返回的批量结果
type VideoPromptBatchResult struct {
	Storyboards []VideoPromptResult `json:"storyboards"`
}

// processStage2VideoPrompts 阶段二：为已保存的分镜批量生成视频提示词和镜头属性
func (s *StoryboardService) processStage2VideoPrompts(taskID, episodeID, model string) error {
	s.log.Infow("Stage 2: generating video prompts", "task_id", taskID, "episode_id", episodeID)

	epID, _ := strconv.Atoi(episodeID)

	// 获取 drama 风格
	style := models.DefaultStyle
	var ep models.Episode
	if err := s.db.Select("drama_id").First(&ep, uint(epID)).Error; err == nil {
		var drama models.Drama
		if err := s.db.Select("style").First(&drama, ep.DramaID).Error; err == nil {
			style = drama.Style
		}
	}

	// 加载已保存的分镜（含角色和道具信息）
	var storyboards []models.Storyboard
	if err := s.db.Preload("Characters").Preload("Props").Preload("Background").
		Where("episode_id = ?", uint(epID)).
		Order("storyboard_number ASC").
		Find(&storyboards).Error; err != nil {
		return fmt.Errorf("加载分镜失败: %w", err)
	}

	if len(storyboards) == 0 {
		return fmt.Errorf("未找到分镜数据")
	}

	// 构建角色外貌速查表
	charAppearanceMap := make(map[uint]string)
	for _, sb := range storyboards {
		for _, c := range sb.Characters {
			if _, exists := charAppearanceMap[c.ID]; !exists {
				var parts []string
				if c.Gender != nil && *c.Gender != "" {
					parts = append(parts, *c.Gender)
				}
				if c.AgeDescription != nil && *c.AgeDescription != "" {
					parts = append(parts, *c.AgeDescription)
				}
				if c.Appearance != nil && *c.Appearance != "" {
					parts = append(parts, *c.Appearance)
				}
				if len(parts) > 0 {
					charAppearanceMap[c.ID] = fmt.Sprintf("%s（%s）", c.Name, strings.Join(parts, "，"))
				} else {
					charAppearanceMap[c.ID] = c.Name
				}
			}
		}
	}

	// 构建批量输入
	var sbDescriptions []string
	for _, sb := range storyboards {
		var charDescs []string
		for _, c := range sb.Characters {
			if desc, ok := charAppearanceMap[c.ID]; ok {
				charDescs = append(charDescs, desc)
			}
		}

		action := ""
		if sb.Action != nil {
			action = *sb.Action
		}
		result := ""
		if sb.Result != nil {
			result = *sb.Result
		}
		atmosphere := ""
		if sb.Atmosphere != nil {
			atmosphere = *sb.Atmosphere
		}
		location := ""
		if sb.Location != nil {
			location = *sb.Location
		}
		timeStr := ""
		if sb.Time != nil {
			timeStr = *sb.Time
		}

		desc := fmt.Sprintf(`--- 分镜 %d ---
时间：%s | 地点：%s | 氛围：%s | 时长：%d秒
角色：%s
动作：%s
结果：%s`,
			sb.StoryboardNumber,
			timeStr, location, atmosphere, sb.Duration,
			strings.Join(charDescs, "；"),
			action,
			result)

		sbDescriptions = append(sbDescriptions, desc)
	}

	styleDesc := "写实摄影风格"
	if style == models.StyleComic {
		styleDesc = "漫画/动漫风格"
	}

	systemPrompt := fmt.Sprintf(`你是一位专业的影视摄影指导和AI视频生成提示词专家。你的任务是为每个分镜设计最佳的镜头属性，并生成高质量的 Seedance 1.5 Pro 视频生成提示词。

**画面风格**：%s（video_prompt 中的画面描述必须符合此风格）

**镜头属性可选值（⚠️ 必须严格使用以下枚举值，禁止自创或变体）**：
- 景别(shot_type)，只能从以下5个值中选1个：远景、全景、中景、近景、特写
- 镜头角度(angle)，只能从以下7个值中选1个：平视、俯视、仰视、高机位、低机位、过肩视角、主观视角
- 运镜(movement)，只能从以下12个值中选1个：固定镜头、推、拉、摇、移、跟、升、降、甩、环绕、旋转、变焦

**video_prompt 提示词公式**：主体 + 运动 + 环境（可选） + 运镜/切镜（可选） + 美学描述（可选） + 声音（可选）

**video_prompt 核心规则**：
1. 用外貌特征指定角色（不要用角色名）：例如"穿黑色职业套装的干练女性"
2. ⚠️ 对话只能来源于分镜的"动作"字段中已有的台词（用「」或""括起的内容），绝对禁止自行编造任何台词！没有台词就不要写对话
3. 如果动作字段包含台词，用外貌特征引出说话人并描述嘴部动作
4. 多人对话使用切镜描述：「镜头从两人同框中景开始...切镜到男子近景...镜头切回女子中近景」
5. 运镜融入描述：使用专业术语（推/拉/摇/移/跟/升/降/环绕），描述起幅→运镜→落幅
6. 环境氛围具体：包含光线、色调、声音细节
7. 声音设计融入尾部：背景音乐风格、关键音效、环境音
8. 长度根据时长调整：4-6秒约100-200字，7-10秒约200-350字，10秒以上约350-500字
9. ⚠️ 禁止使用"缓缓""慢慢""轻轻"等减速副词！动作要干脆利落节奏紧凑，后期可通过剪辑调速。用"随即""紧接着""顺势"串联动作，在时长内尽可能多地展现完整动作

**运镜描述公式**：起幅构图描述 + 运镜方式 + 运镜幅度 + 落幅构图描述

**输出格式**：严格JSON，必须包含所有分镜的结果：
{"storyboards": [{"shot_number": 1, "shot_type": "...", "angle": "...", "movement": "...", "video_prompt": "...", "bgm_prompt": "...", "sound_effect": "..."}, ...]}`, styleDesc)

	userPrompt := fmt.Sprintf("请为以下 %d 个分镜设计镜头属性并生成视频提示词：\n\n%s",
		len(storyboards), strings.Join(sbDescriptions, "\n\n"))

	maxTokens := getMaxTokensForModel(model)
	text, err := s.aiService.GenerateTextForModel(userPrompt, systemPrompt, model, "generate_video_prompts_batch", nil, ai.WithMaxTokens(maxTokens))
	if err != nil {
		return fmt.Errorf("阶段二AI调用失败: %w", err)
	}

	// 解析结果
	var batchResult VideoPromptBatchResult
	var vpResults []VideoPromptResult
	if err := utils.SafeParseAIJSON(text, &vpResults); err == nil {
		batchResult.Storyboards = vpResults
	} else if err := utils.SafeParseAIJSON(text, &batchResult); err != nil {
		s.log.Errorw("Stage 2: failed to parse video prompt results", "error", err, "response_preview", text[:min(500, len(text))])
		return fmt.Errorf("解析视频提示词结果失败: %w", err)
	}

	s.log.Infow("Stage 2: parsed video prompt results", "count", len(batchResult.Storyboards), "task_id", taskID)

	// 合法枚举值
	validShotTypes := map[string]bool{"远景": true, "全景": true, "中景": true, "近景": true, "特写": true}
	validAngles := map[string]bool{"平视": true, "俯视": true, "仰视": true, "高机位": true, "低机位": true, "过肩视角": true, "主观视角": true}
	validMovements := map[string]bool{"固定镜头": true, "推": true, "拉": true, "摇": true, "移": true, "跟": true, "升": true, "降": true, "甩": true, "环绕": true, "旋转": true, "变焦": true}

	// 建立 shot_number -> result 映射，同时校验枚举值
	resultMap := make(map[int]VideoPromptResult)
	for _, vp := range batchResult.Storyboards {
		if !validShotTypes[vp.ShotType] {
			s.log.Warnw("Stage 2: invalid shot_type, resetting", "shot_number", vp.ShotNumber, "got", vp.ShotType)
			vp.ShotType = "中景"
		}
		if !validAngles[vp.Angle] {
			s.log.Warnw("Stage 2: invalid angle, resetting", "shot_number", vp.ShotNumber, "got", vp.Angle)
			vp.Angle = "平视"
		}
		if !validMovements[vp.Movement] {
			s.log.Warnw("Stage 2: invalid movement, resetting", "shot_number", vp.ShotNumber, "got", vp.Movement)
			vp.Movement = "固定镜头"
		}
		resultMap[vp.ShotNumber] = vp
	}

	// 更新数据库中的每个分镜
	for _, sb := range storyboards {
		vp, ok := resultMap[sb.StoryboardNumber]
		if !ok {
			s.log.Warnw("Stage 2: no result for storyboard", "shot_number", sb.StoryboardNumber)
			continue
		}

		updates := map[string]interface{}{}
		if vp.ShotType != "" {
			updates["shot_type"] = vp.ShotType
		}
		if vp.Angle != "" {
			updates["angle"] = vp.Angle
		}
		if vp.Movement != "" {
			updates["movement"] = vp.Movement
		}
		if vp.VideoPrompt != "" {
			updates["video_prompt"] = vp.VideoPrompt
		}
		if vp.BgmPrompt != "" {
			updates["bgm_prompt"] = vp.BgmPrompt
		}
		if vp.SoundEffect != "" {
			updates["sound_effect"] = vp.SoundEffect
		}

		if len(updates) > 0 {
			if err := s.db.Model(&models.Storyboard{}).Where("id = ?", sb.ID).Updates(updates).Error; err != nil {
				s.log.Errorw("Stage 2: failed to update storyboard", "error", err, "storyboard_id", sb.ID)
			} else {
				s.log.Infow("Stage 2: updated storyboard",
					"shot_number", sb.StoryboardNumber,
					"shot_type", vp.ShotType,
					"angle", vp.Angle,
					"movement", vp.Movement)
			}
		}
	}

	return nil
}

// generateVideoPromptWithAI 使用AI生成详细的视频提示词
func (s *StoryboardService) generateVideoPromptWithAI(sb models.Storyboard, firstFrameImage *models.ImageGeneration, lastFrameImage *models.ImageGeneration, duration int, enableSubtitle *bool, generateAudio *bool, aspectRatio string, includeDialogue *bool, style string) (string, error) {
	hasFirstLast := firstFrameImage != nil && lastFrameImage != nil
	middleDesc := safeString(sb.MiddleActionDesc)
	hasDialogueInMiddle := strings.Contains(middleDesc, "「") || strings.Contains(middleDesc, "\u201c") || strings.Contains(middleDesc, "\"")

	var sbInfo strings.Builder

	// 风格信息
	if style == models.StyleComic {
		sbInfo.WriteString("【画面风格】漫画/动漫风格\n")
	} else {
		sbInfo.WriteString("【画面风格】写实摄影风格\n")
	}
	sbInfo.WriteString(fmt.Sprintf("【视频时长】%d秒\n", duration))
	sbInfo.WriteString(fmt.Sprintf("【画面比例】%s\n", aspectRatio))
	if hasDialogueInMiddle {
		sbInfo.WriteString("【人物对话】中间过程描述中包含台词，必须完整保留并描述角色说话时的嘴部张合动作和表情变化。禁止自行编造额外台词。\n")
	} else {
		sbInfo.WriteString("【人物对话】中间过程描述中没有台词，不要写任何台词对话内容，只描述动作和表情变化。禁止自行编造台词。\n")
	}
	if hasFirstLast {
		sbInfo.WriteString("【生成模式】首尾帧模式（Seedance 已有首帧和尾帧图片，提示词只需描述两帧之间的动作过程）\n")
	}

	// 角色名→穿着映射（让 AI 替换名字为穿着特征）
	if len(sb.Characters) > 0 {
		sbInfo.WriteString("\n【角色名→穿着映射】（提示词中必须使用右侧的穿着特征，绝不使用左侧的名字）\n")
		for _, c := range sb.Characters {
			gender := "人物"
			if c.Gender != nil && *c.Gender != "" {
				gender = *c.Gender
			}
			clothing := extractClothingFeature(c.Appearance, gender)
			sbInfo.WriteString(fmt.Sprintf("- %s → %s", c.Name, clothing))
			if hasDialogueInMiddle && c.VoiceStyle != nil && *c.VoiceStyle != "" {
				sbInfo.WriteString(fmt.Sprintf("，声音属性：%s", *c.VoiceStyle))
			}
			sbInfo.WriteString("\n")
		}
	}

	// 中间过程描述（唯一的对话来源）
	if middleDesc != "" {
		sbInfo.WriteString(fmt.Sprintf("\n【中间过程（核心动态素材，注意替换角色名字为穿着特征）】\n%s\n", middleDesc))
	}

	// 当 middle_action_desc 存在时，忽略旧的 action 字段（可能包含过时的台词和描述）
	// 对话只从 middle_action_desc 中提取

	// 首帧/尾帧画面仅供理解起止状态
	firstFrameDesc := safeString(sb.FirstFrameDesc)
	if firstFrameDesc != "" {
		sbInfo.WriteString(fmt.Sprintf("\n【首帧起始状态】（仅供参考，不要在提示词中描述）\n%s\n", firstFrameDesc))
	}
	lastFrameDesc := safeString(sb.LastFrameDesc)
	if lastFrameDesc != "" {
		sbInfo.WriteString(fmt.Sprintf("\n【尾帧结束状态】（仅供参考，不要在提示词中描述）\n%s\n", lastFrameDesc))
	}

	// 场景（仅一句话）
	if sb.Background != nil {
		sbInfo.WriteString(fmt.Sprintf("\n【场景】%s", sb.Background.Location))
		if sb.Background.Time != "" {
			sbInfo.WriteString(fmt.Sprintf("，%s", sb.Background.Time))
		}
		sbInfo.WriteString("\n")
	}

	systemPrompt := s.promptI18n.GetVideoPromptGenerationPrompt(aspectRatio)

	userMessage := fmt.Sprintf(`%s

%s`, s.promptI18n.GetVideoPromptGenerationUserPrompt(), sbInfo.String())

	videoPrompt, err := s.aiService.GenerateTextWithLog(userMessage, systemPrompt, "generate_video_prompt", nil)
	if err != nil {
		s.log.Errorw("Failed to generate video prompt with AI", "error", err, "storyboard_id", sb.StoryboardNumber)
		return "", err
	}

	videoPrompt = strings.TrimSpace(videoPrompt)
	videoPrompt = strings.TrimPrefix(videoPrompt, "```")
	videoPrompt = strings.TrimPrefix(videoPrompt, "```text")
	videoPrompt = strings.TrimPrefix(videoPrompt, "```markdown")
	videoPrompt = strings.TrimSuffix(videoPrompt, "```")
	videoPrompt = strings.TrimSpace(videoPrompt)

	return videoPrompt, nil
}

// GenerateVideoPromptWithAI 使用AI生成详细的视频提示词（公开方法）
func (s *StoryboardService) GenerateVideoPromptWithAI(storyboardID uint, model string, duration int, enableSubtitle *bool, generateAudio *bool, aspectRatio string, includeDialogue *bool) (string, error) {
	// 查询分镜信息，预加载人物、道具、场景
	var storyboard models.Storyboard
	if err := s.db.Preload("Characters").Preload("Props").Preload("Background").Preload("Episode").First(&storyboard, storyboardID).Error; err != nil {
		return "", fmt.Errorf("分镜不存在: %w", err)
	}

	// 获取 drama 风格
	style := models.DefaultStyle
	if storyboard.EpisodeID > 0 {
		var drama models.Drama
		if err := s.db.Select("style").First(&drama, storyboard.Episode.DramaID).Error; err == nil {
			style = drama.Style
		}
	}

	// 查询首帧图片
	var firstFrameImage *models.ImageGeneration
	err := s.db.Where("storyboard_id = ? AND frame_type = ? AND status = ?",
		storyboardID, "first", "completed").
		Order("created_at DESC").First(&firstFrameImage).Error
	if err != nil {
		s.log.Warnw("No first frame image found for video prompt generation", "storyboard_id", storyboardID)
		firstFrameImage = nil
	}

	// 查询尾帧图片
	var lastFrameImage *models.ImageGeneration
	err = s.db.Where("storyboard_id = ? AND frame_type = ? AND status = ?",
		storyboardID, "last", "completed").
		Order("created_at DESC").First(&lastFrameImage).Error
	if err != nil {
		lastFrameImage = nil
	}

	// 构建完整的分镜信息
	return s.generateVideoPromptWithAI(storyboard, firstFrameImage, lastFrameImage, duration, enableSubtitle, generateAudio, aspectRatio, includeDialogue, style)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// genderToChinese 将英文性别转为中文
func genderToChinese(gender string) string {
	switch strings.ToLower(strings.TrimSpace(gender)) {
	case "male", "男", "男性":
		return "男性"
	case "female", "女", "女性":
		return "女性"
	default:
		if gender != "" {
			return gender
		}
		return "人物"
	}
}

// extractClothingFeature 从角色 Appearance 提取简短穿着特征作为视觉标识
// 如 "穿灰衬衫的男性"、"穿黑西装的女性"
func extractClothingFeature(appearance *string, gender string) string {
	genderCN := genderToChinese(gender)
	if appearance == nil || *appearance == "" {
		return genderCN
	}
	desc := *appearance
	// 尝试从 Appearance 中找到"身着/穿着/穿"后面的服装关键词
	clothingPrefixes := []string{"身着", "穿着", "身穿", "穿"}
	for _, prefix := range clothingPrefixes {
		idx := strings.Index(desc, prefix)
		if idx >= 0 {
			after := desc[idx+len(prefix):]
			// 取到标点符号处截断（不包含"的"，避免过早截断）
			endChars := []string{"，", "、", "。", "；", "（", "和", "并", "\n"}
			endIdx := len(after)
			for _, ec := range endChars {
				if i := strings.Index(after, ec); i > 0 && i < endIdx {
					endIdx = i
				}
			}
			clothing := strings.TrimSpace(after[:endIdx])
			runes := []rune(clothing)
			if len(runes) > 25 {
				runes = runes[:25]
			}
			clothing = string(runes)
			if clothing != "" {
				return "穿" + clothing + "的" + genderCN
			}
		}
	}
	// 没有找到穿着信息，截取前20个字作为特征
	runes := []rune(desc)
	if len(runes) > 20 {
		runes = runes[:20]
	}
	return string(runes) + "的" + genderCN
}

// composeActionFromFrameDescs 从首帧/中间过程/尾帧三段描述拼合为一段完整的镜头描述
func composeActionFromFrameDescs(firstFrame, middleAction, lastFrame string) string {
	var parts []string
	if firstFrame != "" {
		parts = append(parts, firstFrame)
	}
	if middleAction != "" {
		parts = append(parts, middleAction)
	}
	if lastFrame != "" {
		parts = append(parts, lastFrame)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

func (s *StoryboardService) saveStoryboards(episodeID string, storyboards []Storyboard) error {
	// 验证 episodeID
	epID, err := strconv.ParseUint(episodeID, 10, 32)
	if err != nil {
		s.log.Errorw("Invalid episode ID", "episode_id", episodeID, "error", err)
		return fmt.Errorf("无效的章节ID: %s", episodeID)
	}

	// 防御性检查：如果AI返回的分镜数量为0，不应该删除旧分镜
	if len(storyboards) == 0 {
		s.log.Errorw("AI返回的分镜数量为0，拒绝保存以避免删除现有分镜", "episode_id", episodeID)
		return fmt.Errorf("AI生成分镜失败：返回的分镜数量为0")
	}

	s.log.Infow("开始保存分镜头",
		"episode_id", episodeID,
		"episode_id_uint", uint(epID),
		"storyboard_count", len(storyboards))

	// 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 验证该章节是否存在
		var episode models.Episode
		if err := tx.First(&episode, epID).Error; err != nil {
			s.log.Errorw("Episode not found", "episode_id", episodeID, "error", err)
			return fmt.Errorf("章节不存在: %s", episodeID)
		}

		s.log.Infow("找到章节信息",
			"episode_id", episode.ID,
			"episode_number", episode.EpisodeNum,
			"drama_id", episode.DramaID,
			"title", episode.Title)

		// 获取该剧集所有的分镜ID（使用 uint 类型）
		var storyboardIDs []uint
		if err := tx.Model(&models.Storyboard{}).
			Where("episode_id = ?", uint(epID)).
			Pluck("id", &storyboardIDs).Error; err != nil {
			return err
		}

		s.log.Infow("查询到现有分镜",
			"episode_id_string", episodeID,
			"episode_id_uint", uint(epID),
			"existing_storyboard_count", len(storyboardIDs),
			"storyboard_ids", storyboardIDs)

		// 如果有分镜，先清理关联的image_generations的storyboard_id
		if len(storyboardIDs) > 0 {
			if err := tx.Model(&models.ImageGeneration{}).
				Where("storyboard_id IN ?", storyboardIDs).
				Update("storyboard_id", nil).Error; err != nil {
				return err
			}
			s.log.Infow("已清理关联的图片生成记录", "count", len(storyboardIDs))
		}

		// 删除该剧集已有的分镜头（使用 uint 类型确保类型匹配）
		s.log.Warnw("准备删除分镜数据",
			"episode_id_string", episodeID,
			"episode_id_uint", uint(epID),
			"episode_id_from_db", episode.ID,
			"will_delete_count", len(storyboardIDs))

		result := tx.Where("episode_id = ?", uint(epID)).Delete(&models.Storyboard{})
		if result.Error != nil {
			s.log.Errorw("删除旧分镜失败", "episode_id", uint(epID), "error", result.Error)
			return result.Error
		}

		s.log.Infow("已删除旧分镜头",
			"episode_id", uint(epID),
			"deleted_count", result.RowsAffected)

		// 注意：不删除背景，因为背景是在分镜拆解前就提取好的
		// AI会直接返回scene_id，不需要在这里做字符串匹配

		// 保存新的分镜头
		for _, sb := range storyboards {
			// 从三段结构化描述自动拼合 action 字段（供下游视频/图片提示词生成使用）
			shotDescription := composeActionFromFrameDescs(sb.FirstFrameDesc, sb.MiddleActionDesc, sb.LastFrameDesc)
			if shotDescription == "" {
				// 兼容：如果 AI 仍返回了 description 或 action
				shotDescription = sb.Description
				if shotDescription == "" {
					shotDescription = sb.Action
				}
			}

			// 使用AI直接返回的SceneID
			if sb.SceneID != nil {
				s.log.Infow("Background ID from AI",
					"shot_number", sb.ShotNumber,
					"scene_id", *sb.SceneID)
			}

			var titlePtr *string
			if sb.Title != "" {
				titlePtr = &sb.Title
			}

		// 把 dialogue 和 action 融合到 middle_action_desc
		if sb.MiddleActionDesc != "" {
			var extras []string
			if sb.Action != "" && !strings.Contains(sb.MiddleActionDesc, sb.Action) {
				extras = append(extras, sb.Action)
			}
			if sb.Dialogue != "" && !strings.Contains(sb.MiddleActionDesc, "「") {
				extras = append(extras, sb.Dialogue)
			}
			if len(extras) > 0 {
				sb.MiddleActionDesc = sb.MiddleActionDesc + "\n" + strings.Join(extras, "\n")
			}
		} else {
			// middle_action_desc 为空时，用 action + dialogue 拼合
			var parts []string
			if sb.Action != "" {
				parts = append(parts, sb.Action)
			}
			if sb.Dialogue != "" {
				parts = append(parts, sb.Dialogue)
			}
			if len(parts) > 0 {
				sb.MiddleActionDesc = strings.Join(parts, "\n")
			}
		}

		var prevShotRefPtr, firstFrameDescPtr, middleActionDescPtr, lastFrameDescPtr *string
		if sb.PrevShotRef != "" {
			prevShotRefPtr = &sb.PrevShotRef
		}
		if sb.FirstFrameDesc != "" {
			firstFrameDescPtr = &sb.FirstFrameDesc
		}
		if sb.MiddleActionDesc != "" {
			middleActionDescPtr = &sb.MiddleActionDesc
		}
		if sb.LastFrameDesc != "" {
			lastFrameDescPtr = &sb.LastFrameDesc
		}

		scene := models.Storyboard{
			EpisodeID:        uint(epID),
			SceneID:          sb.SceneID,
			StoryboardNumber: sb.ShotNumber,
			Title:            titlePtr,
			Location:         &sb.Location,
			Time:             &sb.Time,
			Action:           &shotDescription,
			PrevShotRef:      prevShotRefPtr,
			FirstFrameDesc:   firstFrameDescPtr,
			MiddleActionDesc: middleActionDescPtr,
			LastFrameDesc:    lastFrameDescPtr,
			Duration:         sb.Duration,
		}

			if err := tx.Create(&scene).Error; err != nil {
				s.log.Errorw("Failed to create scene", "error", err, "shot_number", sb.ShotNumber)
				return err
			}

			// 关联角色
			if len(sb.Characters) > 0 {
				var characters []models.Character
				if err := tx.Where("id IN ?", sb.Characters).Find(&characters).Error; err != nil {
					s.log.Warnw("Failed to load characters for association", "error", err, "character_ids", sb.Characters)
				} else if len(characters) > 0 {
					if err := tx.Model(&scene).Association("Characters").Append(characters); err != nil {
						s.log.Warnw("Failed to associate characters", "error", err, "shot_number", sb.ShotNumber)
					} else {
						s.log.Infow("Characters associated successfully",
							"shot_number", sb.ShotNumber,
							"character_ids", sb.Characters,
							"count", len(characters))
					}
				}
			}

			// 关联道具
			if len(sb.Props) > 0 {
				var propModels []models.Prop
				if err := tx.Where("id IN ?", sb.Props).Find(&propModels).Error; err != nil {
					s.log.Warnw("Failed to load props for association", "error", err, "prop_ids", sb.Props)
				} else if len(propModels) > 0 {
					if err := tx.Model(&scene).Association("Props").Append(propModels); err != nil {
						s.log.Warnw("Failed to associate props", "error", err, "shot_number", sb.ShotNumber)
					} else {
						s.log.Infow("Props associated successfully",
							"shot_number", sb.ShotNumber,
							"prop_ids", sb.Props,
							"count", len(propModels))
					}
				}
			}
		}

		s.log.Infow("Storyboards saved successfully", "episode_id", episodeID, "count", len(storyboards))
		return nil
	})
}

// CreateStoryboardRequest 创建分镜请求
type CreateStoryboardRequest struct {
	EpisodeID        uint    `json:"episode_id"`
	SceneID          *uint   `json:"scene_id"`
	StoryboardNumber int     `json:"storyboard_number"`
	Title            *string `json:"title"`
	Location         *string `json:"location"`
	Time             *string `json:"time"`
	ShotType         *string `json:"shot_type"`
	Angle            *string `json:"angle"`
	Movement         *string `json:"movement"`
	Description      *string `json:"description"`
	Action           *string `json:"action"`
	Result           *string `json:"result"`
	Atmosphere       *string `json:"atmosphere"`
	Dialogue         *string `json:"dialogue"`
	BgmPrompt        *string `json:"bgm_prompt"`
	SoundEffect      *string `json:"sound_effect"`
	Duration         int     `json:"duration"`
	Characters       []uint  `json:"characters"`
	Props            []uint  `json:"props"`
}

// CreateStoryboard 创建单个分镜
func (s *StoryboardService) CreateStoryboard(req *CreateStoryboardRequest) (*models.Storyboard, error) {
	// 获取 drama 风格
	style := models.DefaultStyle
	var ep models.Episode
	if err := s.db.Select("drama_id").First(&ep, req.EpisodeID).Error; err == nil {
		var drama models.Drama
		if err := s.db.Select("style").First(&drama, ep.DramaID).Error; err == nil {
			style = drama.Style
		}
	}

	// 构建Storyboard对象
	sb := Storyboard{
		ShotNumber:  req.StoryboardNumber,
		ShotType:    getString(req.ShotType),
		Angle:       getString(req.Angle),
		Time:        getString(req.Time),
		Location:    getString(req.Location),
		SceneID:     req.SceneID,
		Movement:    getString(req.Movement),
		Action:      getString(req.Action),
		Dialogue:    getString(req.Dialogue),
		Result:      getString(req.Result),
		Atmosphere:  getString(req.Atmosphere),
		Emotion:     "",
		Duration:    req.Duration,
		BgmPrompt:   getString(req.BgmPrompt),
		SoundEffect: getString(req.SoundEffect),
		Characters:  req.Characters,
	}
	if req.Title != nil {
		sb.Title = *req.Title
	}

	// 生成提示词
	imagePrompt := s.generateImagePrompt(sb, style)
	videoPrompt := s.generateVideoPrompt(sb)

	// 构建 description
	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}

	modelSB := &models.Storyboard{
		EpisodeID:        req.EpisodeID,
		SceneID:          req.SceneID,
		StoryboardNumber: req.StoryboardNumber,
		Title:            req.Title,
		Location:         req.Location,
		Time:             req.Time,
		ShotType:         req.ShotType,
		Angle:            req.Angle,
		Movement:         req.Movement,
		Description:      &desc,
		Action:           req.Action,
		Result:           req.Result,
		Atmosphere:       req.Atmosphere,
		Dialogue:         req.Dialogue,
		ImagePrompt:      &imagePrompt,
		VideoPrompt:      &videoPrompt,
		BgmPrompt:        req.BgmPrompt,
		SoundEffect:      req.SoundEffect,
		Duration:         req.Duration,
	}

	if err := s.db.Create(modelSB).Error; err != nil {
		return nil, fmt.Errorf("failed to create storyboard: %w", err)
	}

	// 关联角色
	if len(req.Characters) > 0 {
		var characters []models.Character
		if err := s.db.Where("id IN ?", req.Characters).Find(&characters).Error; err != nil {
			s.log.Warnw("Failed to find characters for new storyboard", "error", err)
		} else if len(characters) > 0 {
			s.db.Model(modelSB).Association("Characters").Append(characters)
		}
	}

	// 关联道具
	if len(req.Props) > 0 {
		var propModels []models.Prop
		if err := s.db.Where("id IN ?", req.Props).Find(&propModels).Error; err != nil {
			s.log.Warnw("Failed to find props for new storyboard", "error", err)
		} else if len(propModels) > 0 {
			s.db.Model(modelSB).Association("Props").Append(propModels)
		}
	}

	s.log.Infow("Storyboard created", "id", modelSB.ID, "episode_id", req.EpisodeID)
	return modelSB, nil
}

// DeleteStoryboard 删除分镜
func (s *StoryboardService) DeleteStoryboard(storyboardID uint) error {
	result := s.db.Where("id = ? ", storyboardID).Delete(&models.Storyboard{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("storyboard not found")
	}
	return nil
}

// getMaxTokensForModel 根据模型名称返回适合的 max_tokens
// 大模型用大 token，小模型用小 token，避免截断或超限
func getMaxTokensForModel(model string) int {
	model = strings.ToLower(model)

	// 豆包/火山引擎模型：max_tokens 上限为 12288
	if strings.Contains(model, "doubao") ||
		strings.Contains(model, "skylark") {
		return 12288
	}

	// Lite / 小尺寸模型：限制在安全范围内
	if strings.Contains(model, "lite") ||
		strings.Contains(model, "mini") ||
		strings.Contains(model, "tiny") {
		return 8192
	}

	// Pro / 大尺寸模型：支持更大的 max_tokens
	if strings.Contains(model, "pro") ||
		strings.Contains(model, "plus") ||
		strings.Contains(model, "max") ||
		strings.Contains(model, "gpt-4") ||
		strings.Contains(model, "claude") ||
		strings.Contains(model, "gemini") ||
		strings.Contains(model, "qwen-long") ||
		strings.Contains(model, "qwen2.5-72b") ||
		strings.Contains(model, "deepseek") {
		return 16384
	}

	// 默认值：使用大尺寸，确保分镜完整输出
	return 16384
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
