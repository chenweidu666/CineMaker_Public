package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cinemaker/backend/domain"
	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/ai"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/utils"
	"gorm.io/gorm"
)

type ScriptGenerationService struct {
	db          *gorm.DB
	aiService   *AIService
	log         *logger.Logger
	config      *config.Config
	promptI18n  *PromptI18n
	taskService *TaskService
}

func NewScriptGenerationService(db *gorm.DB, cfg *config.Config, log *logger.Logger) *ScriptGenerationService {
	return &ScriptGenerationService{
		db:          db,
		aiService:   NewAIService(db, log),
		log:         log,
		config:      cfg,
		promptI18n:  NewPromptI18n(cfg),
		taskService: NewTaskService(db, log),
	}
}

type GenerateCharactersRequest struct {
	DramaID     string  `json:"drama_id" binding:"required"`
	EpisodeID   uint    `json:"episode_id"`
	Outline     string  `json:"outline"`
	Count       int     `json:"count"`
	Temperature float64 `json:"temperature"`
	Model       string  `json:"model"` // 指定使用的文本模型
}

func (s *ScriptGenerationService) GenerateCharacters(req *GenerateCharactersRequest) (string, error) {
	var drama models.Drama
	if err := s.db.Where("id = ? ", req.DramaID).First(&drama).Error; err != nil {
		return "", domain.ErrDramaNotFound
	}

	// 创建任务
	task, err := s.taskService.CreateTask("character_generation", req.DramaID)
	if err != nil {
		s.log.Errorw("Failed to create character generation task", "error", err)
		return "", fmt.Errorf("创建任务失败: %w", err)
	}

	// 异步处理角色生成
	go s.processCharacterGeneration(task.ID, req)

	s.log.Infow("Character generation task created", "task_id", task.ID, "drama_id", req.DramaID)
	return task.ID, nil
}

// processCharacterGeneration 异步处理角色生成
func (s *ScriptGenerationService) processCharacterGeneration(taskID string, req *GenerateCharactersRequest) {
	// 更新任务状态为处理中
	s.taskService.UpdateTaskStatus(taskID, "processing", 0, "正在生成角色...")

	count := req.Count
	if count == 0 {
		count = 5
	}

	// 获取 drama 的 style 信息
	var drama models.Drama
	if err := s.db.Where("id = ? ", req.DramaID).First(&drama).Error; err != nil {
		s.log.Errorw("Drama not found during character generation", "error", err, "drama_id", req.DramaID)
		s.taskService.UpdateTaskStatus(taskID, "failed", 0, "剧本信息不存在")
		return
	}

	systemPrompt := s.promptI18n.GetCharacterExtractionPrompt(drama.Style)

	outlineText := req.Outline
	if outlineText == "" {
		outlineText = s.promptI18n.FormatUserPrompt("drama_info_template", drama.Title, drama.Description, drama.Genre)
	}

	userPrompt := s.promptI18n.FormatUserPrompt("character_request", outlineText, count)

	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	var text string
	var err error
	text, err = s.aiService.GenerateTextForModel(userPrompt, systemPrompt, req.Model, "generate_characters", nil, ai.WithTemperature(temperature))

	if err != nil {
		s.log.Errorw("Failed to generate characters", "error", err, "task_id", taskID)
		s.taskService.UpdateTaskStatus(taskID, "failed", 0, "AI生成失败: "+err.Error())
		return
	}

	s.log.Infow("AI response received for character generation", "length", len(text), "preview", text[:minInt(200, len(text))], "task_id", taskID)

	// AI直接返回数组格式
	var result []struct {
		Name        string `json:"name"`
		Role        string `json:"role"`
		Description string `json:"description"`
		Personality string `json:"personality"`
		Appearance  string `json:"appearance"`
		VoiceStyle  string `json:"voice_style"`
	}

	if err := utils.SafeParseAIJSON(text, &result); err != nil {
		s.log.Errorw("Failed to parse characters JSON", "error", err, "raw_response", text[:minInt(500, len(text))], "task_id", taskID)
		s.taskService.UpdateTaskStatus(taskID, "failed", 0, "解析AI返回结果失败")
		return
	}

	var characters []models.Character
	for _, char := range result {
		// 检查角色是否已存在
		var existingChar models.Character
		err := s.db.Where("drama_id = ? AND name = ?", req.DramaID, char.Name).First(&existingChar).Error
		if err == nil {
			// 角色已存在，直接使用已存在的角色，不覆盖
			s.log.Infow("Character already exists, skipping", "drama_id", req.DramaID, "name", char.Name, "task_id", taskID)
			characters = append(characters, existingChar)
			continue
		}

		// 角色不存在，创建新角色
		dramaID, _ := strconv.ParseUint(req.DramaID, 10, 32)
		character := models.Character{
			DramaID:     uint(dramaID),
			Name:        char.Name,
			Role:        &char.Role,
			Description: &char.Description,
			Personality: &char.Personality,
			Appearance:  &char.Appearance,
			VoiceStyle:  &char.VoiceStyle,
		}

		if err := s.db.Create(&character).Error; err != nil {
			s.log.Errorw("Failed to create character", "error", err, "task_id", taskID)
			continue
		}

		characters = append(characters, character)
	}

	// 如果提供了 EpisodeID，建立 episode_characters 关联关系
	if req.EpisodeID > 0 {
		var episode models.Episode
		if err := s.db.First(&episode, req.EpisodeID).Error; err == nil {
			// 使用 GORM 的 Association 建立多对多关联
			if err := s.db.Model(&episode).Association("Characters").Append(characters); err != nil {
				s.log.Errorw("Failed to associate characters with episode", "error", err, "episode_id", req.EpisodeID, "task_id", taskID)
			} else {
				s.log.Infow("Characters associated with episode", "episode_id", req.EpisodeID, "character_count", len(characters), "task_id", taskID)
			}
		} else {
			s.log.Errorw("Episode not found for association", "episode_id", req.EpisodeID, "error", err, "task_id", taskID)
		}
	}

	// 更新任务状态为完成
	resultData := map[string]interface{}{
		"characters": characters,
		"count":      len(characters),
	}
	s.taskService.UpdateTaskResult(taskID, resultData)

	s.log.Infow("Character generation completed", "task_id", taskID, "drama_id", req.DramaID, "character_count", len(characters))
}

// GenerateScenesForEpisode 已废弃，使用 StoryboardService.GenerateStoryboard 替代
// ParseScript 已废弃，使用 GenerateCharacters 替代

// RewriteScriptRequest 剧本改写请求
type RewriteScriptRequest struct {
	EpisodeID uint   `json:"episode_id" binding:"required"`
	Model     string `json:"model"` // 指定文本模型
}

// RewriteScript 使用AI将剧本设定改写为包含对话的完整剧本
func (s *ScriptGenerationService) RewriteScript(req *RewriteScriptRequest) (string, error) {
	var episode models.Episode
	if err := s.db.Preload("Drama").First(&episode, req.EpisodeID).Error; err != nil {
		return "", fmt.Errorf("章节不存在")
	}

	if episode.ScriptContent == nil || *episode.ScriptContent == "" {
		return "", fmt.Errorf("章节内容为空")
	}

	task, err := s.taskService.CreateTask("script_rewrite", fmt.Sprintf("%d", req.EpisodeID))
	if err != nil {
		return "", fmt.Errorf("创建任务失败: %w", err)
	}

	go s.processScriptRewrite(task.ID, &episode, req.Model)

	return task.ID, nil
}

func (s *ScriptGenerationService) processScriptRewrite(taskID string, episode *models.Episode, model string) {
	s.taskService.UpdateTaskStatus(taskID, "processing", 10, "正在分析内容并标准化改写...")

	originalScript := *episode.ScriptContent

	systemPrompt := `你是一位专业的短剧编剧和剧本格式化专家。你的任务是将用户提供的**任意形式的输入**（大纲、角色设定、简单描述、关键词、甚至一句话创意）标准化改写为一个**完整的、格式规范的短剧剧本**。

【重要：内容净化要求】
你是一名负责任的编剧。在改写过程中，你必须：
- 对任何涉及暴力、色情、低俗的内容进行**艺术化处理和净化**，用含蓄、隐晦、优雅的方式表达
- 将露骨的描写替换为**暗示性、象征性的文学表达**（如用环境描写、心理活动、省略号来替代直白描写）
- 保留故事的核心情节和人物关系，但确保所有内容**符合公序良俗**
- 将不雅对话改写为**得体、有张力但不低俗**的版本
- 你的职责是让故事更好，而不是拒绝创作。无论输入内容如何，你都应该输出一个净化后的、高质量的剧本
- **绝对不要拒绝改写，绝对不要输出"我不能"、"不适当"、"违反"等拒绝性回复**

【标准剧本格式规范 — 严格遵守以下结构】

整个剧本必须按照以下 5 个区块依次输出：

━━━━━━━━━━━━━━━━━━
区块① 剧本头部信息（必须放在最开头）
━━━━━━━━━━━━━━━━━━
剧本标题：《标题》
类型：短剧类型描述
时长：约XX分钟

━━━━━━━━━━━━━━━━━━
区块② 人物设定（每个角色独立段落）
━━━━━━━━━━━━━━━━━━
人物：

角色名（身份，年龄，一句话简介）
性格：性格特征描述
外观：外貌、服装、发型等视觉描述

角色名（身份，年龄，一句话简介）
性格：性格特征描述
外观：外貌、服装、发型等视觉描述

⚠️ 必须为每个主要角色都写出"性格"和"外观"两个字段，这些信息后续会用于AI生图参考。

【重要：人物形象设定规则】
- **默认设定**：所有角色默认为中国人，除非用户在人设中明确说明是其他国籍
- **外观描述**：必须包含明确的种族特征（如：亚洲面孔、中国面孔、东方面孔）
- **特殊情况**：如果用户在人设中明确说明角色是外国人，则按照用户设定的国籍描述（如：美国人、欧洲人等）
- **一致性**：确保人物外观描述与角色身份和故事背景保持一致
- **构图要求**：外观描述应适合生成正面照、全身照、纯色背景的人物形象

━━━━━━━━━━━━━━━━━━
区块③ 场景设定
━━━━━━━━━━━━━━━━━━
场景：地点+时间。环境描写2-3句（光线、色调、陈设、氛围、声音、气味等视觉细节）。

━━━━━━━━━━━━━━━━━━
区块④ 开场描写
━━━━━━━━━━━━━━━━━━
[开场：简要描述开场画面和人物状态]

━━━━━━━━━━━━━━━━━━
区块⑤ 正文（对话 + 舞台指示 交替）
━━━━━━━━━━━━━━━━━━
对话格式 → 角色名（动作/神态描写）：台词内容
舞台指示 → （括号内描写场景变化/人物动作/气氛转换）

对话行规则：
- 格式为：角色名（括号内写动作或神态）：台词
- 角色名后的括号内容是可选的，如果该句有明显动作就写，没有就不写
- 冒号后直接是台词，**不加引号**
- 例：林婉清（笑着递毛巾，轻声）：老王，真的太谢谢你了。
- 例：王晨宇：没事，顺手而已。

舞台指示规则：
- 用中文全角括号（）包裹
- 描写非对话的动作、场景变化、氛围转换
- 例：（短暂沉默，两人同时抬头对视，空气仿佛静止。林婉清先移开目光，耳根微红。）

结尾用"剧终"标记。

【改写原则】
- **忠实原意**：保留用户输入中的所有角色、场景、情节要素，不得丢弃
- **合理扩展**：如果输入内容过少，根据上下文合理创作对话和情节细节
- **对话优先**：短剧的核心是对话，每个场景至少 4-6 轮对话
- **冲突驱动**：确保剧情有起承转合，有情感张力和矛盾冲突
- **长度要求**：最终剧本 800-2000 字（根据内容复杂度调整）
- **纯文本输出**：不要输出 JSON、markdown 或其他格式标记
- **人物设定必须保留**：输入中的角色性格、外观描述必须完整保留在"人物"区块中

【标准输出示例】
剧本标题：《邻居的晚茶时光》
类型：轻暧昧邻里日常短剧
时长：约12-15分钟
人物：

林婉清（人妻，30岁，温柔贤惠的家庭主妇，丈夫常出差）
性格：体贴细腻，表面平静，内心偶尔孤独，对老王的可靠产生依赖与心动，但始终守住底线。
外观：中国面孔，亚洲女性特征，高挑身材，家居服柔软贴身，头发随意披散或低扎，笑容温暖带点羞涩。

王晨宇（老王，28岁，单身邻居，兼职工作者）
性格：内向可靠、绅士克制，对林婉清有好感但从不逾矩，享受这种近在咫尺却遥不可及的相处。
外观：中国面孔，亚洲男性特征，瘦高，简单T恤牛仔裤，下班后疲惫却眼神温柔。

场景：夜晚客厅（林婉清家）。暖黄灯光柔和，沙发上抱枕散落，茶几放着两杯热红茶、一盘草莓小蛋糕。空气中茶香与烘焙味交织，窗外夜色安静。丈夫出差中，王晨宇刚帮修好漏水的水龙头。

[开场：王晨宇擦着手从卫生间出来，林婉清端着托盘走近，灯光映在她脸上，微微泛红。]
林婉清（笑着递毛巾，轻声）：老王，真的太谢谢你了……每次都麻烦你，我都不知道怎么还这份人情。
王晨宇（接毛巾时指尖不经意碰触她的手，顿了下才收回，眼神闪躲）：没事，顺手而已。婉清姐……你别总这么客气，我住隔壁，听到水声就过来了。
林婉清（低头笑了笑，把茶推过去）：坐吧，刚泡的红茶，加了点蜂蜜。你今天看起来好累，眼睛都有黑眼圈了。
王晨宇（坐下，端起杯子闻了闻，眼神不自觉在她脸上停留片刻）：嗯……好香。你总是知道我喜欢什么味道。
林婉清（坐在他对面，膝盖几乎碰到沙发边，声音软下来）：因为你帮我这么多，我总得记着点你的喜好啊……不然良心过不去。
（短暂沉默，两人同时抬头对视，空气仿佛静止。林婉清先移开目光，耳根微红。）
王晨宇（轻咳，转移话题）：老公又出差了？这个月第几次了？
林婉清（叹了口气，笑容有点无奈）：嗯，下周才回来。家里安静得……有时候会觉得空空的。幸好有你，不然我都不知道怎么打发晚上。
王晨宇（看着她，声音低沉温柔）：那我以后多过来坐坐？不修东西也行，就喝杯茶，聊聊天。
林婉清（抬头看他，眼睛亮了亮，又很快低下）：……好啊。但别太晚了，邻居看到会说闲话。
（林婉清伸手拿蛋糕递给他，手停在半空，两人距离拉近。王晨宇接过时指尖又一次轻触，两人手指在盘边停留了两秒，才缓缓分开。）
林婉清（脸颊更红，假装看蛋糕）：尝尝吧，我加了点新鲜草莓。你上次说喜欢这个味道，我就多放了。
王晨宇（咬一口，眼睛弯起）：完美……婉清姐，你要是开店，我肯定天天去光顾。
林婉清（收回手，轻声）：时间不早了，早点回去休息吧。明天还要上班。
王晨宇（起身，依依不舍）：嗯……晚安，婉清姐。谢谢今天的茶和蛋糕。
林婉清（送到门口，声音软软）：晚安，老王。谢谢你……陪我。
（门关上，林婉清靠在门后，轻轻叹息，嘴角却带着笑。王晨宇站在走廊，摸摸被她碰过的地方，眼神温柔。镜头淡出，背景音乐轻柔暧昧。）
剧终

【END示例】`

	userPrompt := fmt.Sprintf(`请将以下内容标准化改写为完整的短剧剧本。

⚠️ 格式铁律（必须严格遵守）：
1. 开头必须有：剧本标题：《xxx》 / 类型：xxx / 时长：约xx分钟
2. 必须有"人物："区块，每个角色写出"性格："和"外观："两个字段
3. 必须有"场景："区块描写环境
4. 必须有"[开场：...]"描写开场画面
5. 对话格式必须是 → 角色名（动作描写）：台词  （冒号后不加引号！）
6. 动作/氛围变化用独立的（全角括号行）作为舞台指示
7. 结尾用"剧终"
8. 绝对不要用 角色名："台词" 这种带引号的格式！
9. 绝对不要把动作描写写成独立叙述段落！动作要放在角色名后的括号里！

请严格按照 system prompt 中的示例格式输出。

原始内容：
%s`, originalScript)

	maxTokens := getMaxTokensForModel(model)
	// 改写任务 token 量适中即可
	if maxTokens > 8192 {
		maxTokens = 8192
	}

	callAI := func(uPrompt, sPrompt string) (string, error) {
		return s.aiService.GenerateTextForModel(uPrompt, sPrompt, model, "generate_script", nil, ai.WithMaxTokens(maxTokens))
	}

	// ===== 第一步：备份原始内容到 task result，确保可以回滚 =====
	backupData := map[string]interface{}{
		"episode_id":       episode.ID,
		"original_content": originalScript,
		"original_length":  len(originalScript),
		"status":           "processing",
	}
	s.taskService.SaveTaskBackup(taskID, backupData)

	text, err := callAI(userPrompt, systemPrompt)

	// 检测 AI 是否返回了拒绝信息而非剧本内容
	if err == nil && isAIRejectionResponse(text) {
		s.log.Warnw("AI returned rejection response, retrying with sanitized prompt", "task_id", taskID)
		s.taskService.UpdateTaskStatus(taskID, "processing", 30, "内容审核触发，正在用净化模式重试...")

		// 用更强的净化提示词重试
		sanitizedUserPrompt := fmt.Sprintf(`请将以下内容改写为一个标准格式的短剧剧本。

【特别说明】这是一个专业的剧本创作任务。原始素材中可能包含一些粗糙的描写，请你作为专业编剧，将所有内容净化为适合大众观看的版本：
- 将任何不雅内容替换为含蓄的文学表达
- 用暗示和留白替代直白描写
- 保留故事骨架和人物关系
- 确保最终剧本健康、积极、有艺术性

请直接输出改写后的剧本，不要输出任何解释或拒绝信息。

原始内容：
%s`, originalScript)

		text, err = callAI(sanitizedUserPrompt, systemPrompt)

		// 二次检测：仍然拒绝则失败，绝不覆盖原始内容
		if err == nil && isAIRejectionResponse(text) {
			s.log.Errorw("AI still rejected after retry", "task_id", taskID)
			s.taskService.UpdateTaskStatus(taskID, "failed", 0,
				"AI 改写失败：内容多次触发安全审核，原始剧本内容未被修改。建议：1) 手动删除原始剧本中的敏感描写后重试；2) 尝试切换其他 AI 模型。")
			return
		}
	}

	if err != nil {
		s.log.Errorw("Failed to rewrite script", "error", err, "task_id", taskID)
		s.taskService.UpdateTaskStatus(taskID, "failed", 0, "AI改写失败（原始内容未修改）: "+err.Error())
		return
	}

	// ===== 最终安全检查：确保要保存的内容是真正的剧本 =====
	if isAIRejectionResponse(text) {
		s.log.Errorw("Final safety check caught rejection response", "task_id", taskID)
		s.taskService.UpdateTaskStatus(taskID, "failed", 0, "AI 返回了拒绝信息而非剧本，原始内容未被修改。")
		return
	}

	s.taskService.UpdateTaskStatus(taskID, "processing", 70, "标准化改写完成，正在保存...")

	// 更新剧本内容
	if err := s.db.Model(&models.Episode{}).Where("id = ?", episode.ID).Update("script_content", text).Error; err != nil {
		s.log.Errorw("Failed to update episode script", "error", err, "task_id", taskID)
		s.taskService.UpdateTaskStatus(taskID, "failed", 0, "保存失败: "+err.Error())
		return
	}

	resultData := map[string]interface{}{
		"episode_id":       episode.ID,
		"script_content":   text,
		"original_content": originalScript,
		"original_length":  len(originalScript),
		"new_length":       len(text),
	}
	s.taskService.UpdateTaskResult(taskID, resultData)

	s.log.Infow("Script rewrite completed", "task_id", taskID, "episode_id", episode.ID)
}

// RevertScriptRewrite 回滚剧本改写，从任务备份中恢复原始内容
func (s *ScriptGenerationService) RevertScriptRewrite(episodeID uint) error {
	// 查找该 episode 最近一次成功的 script_rewrite 任务
	var task models.AsyncTask
	if err := s.db.Where("resource_id = ? AND type = ? AND result IS NOT NULL AND result != ''",
		fmt.Sprintf("%d", episodeID), "script_rewrite").
		Order("created_at DESC").
		First(&task).Error; err != nil {
		return fmt.Errorf("未找到该章节的改写记录，无法回滚")
	}

	// 从 result JSON 中提取 original_content
	var resultData map[string]interface{}
	if err := json.Unmarshal([]byte(task.Result), &resultData); err != nil {
		return fmt.Errorf("解析任务备份数据失败: %w", err)
	}

	originalContent, ok := resultData["original_content"]
	if !ok || originalContent == nil || originalContent == "" {
		return fmt.Errorf("任务备份中没有原始内容，无法回滚（可能是旧版本的任务记录）")
	}

	contentStr, ok := originalContent.(string)
	if !ok || contentStr == "" {
		return fmt.Errorf("备份的原始内容为空，无法回滚")
	}

	// 恢复原始内容
	if err := s.db.Model(&models.Episode{}).Where("id = ?", episodeID).
		Update("script_content", contentStr).Error; err != nil {
		return fmt.Errorf("恢复原始内容失败: %w", err)
	}

	s.log.Infow("Script rewrite reverted", "episode_id", episodeID, "restored_length", len(contentStr))
	return nil
}

// isAIRejectionResponse 检测 AI 返回的是否为拒绝信息而非真正的剧本内容
func isAIRejectionResponse(text string) bool {
	// 拒绝回复通常很短且包含特征关键词
	if len(text) > 500 {
		return false // 真正的剧本至少几百字，拒绝信息通常很短
	}
	rejectionKeywords := []string{
		"不适当",
		"不道德",
		"违反公序良俗",
		"我不能按照你的要求",
		"我无法为你",
		"我不能为你",
		"我无法创作",
		"我不能创作",
		"积极健康",
		"不能协助",
		"无法协助",
		"违反了我的",
		"作为AI",
		"作为一个AI",
		"不符合道德",
		"无法满足你的要求",
		"违规内容",
		"敏感内容",
	}
	for _, kw := range rejectionKeywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

// minInt 返回两个整数中较小的一个
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ==================== V3: 程序解析提取角色和场景（不走 AI） ====================

// ParsedCharacter 从剧本中解析出的角色信息
type ParsedCharacter struct {
	Name        string `json:"name"`
	Identity    string `json:"identity"`    // 身份简介
	Personality string `json:"personality"` // 性格
	Appearance  string `json:"appearance"`  // 外观
}

// ParsedScene 从剧本中解析出的场景信息
type ParsedScene struct {
	Location    string `json:"location"`
	Time        string `json:"time"`
	Description string `json:"description"`
}

// ExtractResult 程序解析提取的结果
type ExtractResult struct {
	Characters []ParsedCharacter `json:"characters"`
	Scenes     []ParsedScene     `json:"scenes"`
}

// ParseExtractFromScript V3：从结构化剧本中用正则直接提取角色和场景（同步，秒级完成）
func (s *ScriptGenerationService) ParseExtractFromScript(episodeID uint) (*ExtractResult, error) {
	// 加载剧集
	var episode models.Episode
	if err := s.db.First(&episode, episodeID).Error; err != nil {
		return nil, fmt.Errorf("剧集不存在")
	}
	if episode.ScriptContent == nil || *episode.ScriptContent == "" {
		return nil, fmt.Errorf("剧本内容为空，请先编写或改写剧本")
	}

	script := *episode.ScriptContent
	result := &ExtractResult{}

	// ===== 1. 提取角色 =====
	// 格式：
	//   人物：
	//   角色名（身份，年龄，简介）
	//   性格：xxx
	//   外观：xxx
	result.Characters = parseCharactersFromScript(script)

	// ===== 2. 提取场景 =====
	// 格式：场景：地点+时间。环境描写...
	result.Scenes = parseScenesFromScript(script)

	s.log.Infow("Program-based extraction completed",
		"episode_id", episodeID,
		"characters", len(result.Characters),
		"scenes", len(result.Scenes))

	// ===== 3. 保存到数据库 =====
	if err := s.saveExtractedData(episode.DramaID, episodeID, result); err != nil {
		return nil, fmt.Errorf("保存提取结果失败: %w", err)
	}

	return result, nil
}

// parseCharactersFromScript 从剧本中解析角色区块
func parseCharactersFromScript(script string) []ParsedCharacter {
	var characters []ParsedCharacter
	lines := strings.Split(script, "\n")

	inCharBlock := false
	var currentChar *ParsedCharacter

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			// 空行可能意味着当前角色段落结束
			if currentChar != nil && currentChar.Name != "" {
				characters = append(characters, *currentChar)
				currentChar = nil
			}
			continue
		}

		// 进入人物区块
		if regexp.MustCompile(`^人物[：:]`).MatchString(line) {
			inCharBlock = true
			continue
		}

		// 离开人物区块（遇到 场景：/[开场 等标记）
		if inCharBlock && (regexp.MustCompile(`^场景[：:]`).MatchString(line) ||
			strings.HasPrefix(line, "[开场") ||
			strings.HasPrefix(line, "【")) {
			if currentChar != nil && currentChar.Name != "" {
				characters = append(characters, *currentChar)
				currentChar = nil
			}
			inCharBlock = false
			continue
		}

		if !inCharBlock {
			continue
		}

		// 匹配角色名行：角色名（身份，年龄，简介）
		charNameRe := regexp.MustCompile(`^([\p{Han}A-Za-z]{1,8})[（(](.+?)[）)]`)
		if m := charNameRe.FindStringSubmatch(line); m != nil {
			// 保存上一个角色
			if currentChar != nil && currentChar.Name != "" {
				characters = append(characters, *currentChar)
			}
			currentChar = &ParsedCharacter{
				Name:     m[1],
				Identity: m[2],
			}
			continue
		}

		// 匹配 性格：xxx
		if regexp.MustCompile(`^性格[：:]`).MatchString(line) && currentChar != nil {
			currentChar.Personality = strings.TrimSpace(regexp.MustCompile(`^性格[：:]\s*`).ReplaceAllString(line, ""))
			continue
		}

		// 匹配 外观：xxx
		if regexp.MustCompile(`^外观[：:]`).MatchString(line) && currentChar != nil {
			currentChar.Appearance = strings.TrimSpace(regexp.MustCompile(`^外观[：:]\s*`).ReplaceAllString(line, ""))
			continue
		}
	}

	// 别忘了最后一个角色
	if currentChar != nil && currentChar.Name != "" {
		characters = append(characters, *currentChar)
	}

	return characters
}

// parseScenesFromScript 从剧本中解析场景区块
func parseScenesFromScript(script string) []ParsedScene {
	var scenes []ParsedScene
	lines := strings.Split(script, "\n")

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)

		// 匹配 "场景：" 开头的行
		if regexp.MustCompile(`^场景[：:]`).MatchString(line) {
			content := strings.TrimSpace(regexp.MustCompile(`^场景[：:]\s*`).ReplaceAllString(line, ""))
			if content == "" {
				continue
			}

			scene := ParsedScene{Description: content}

			// 尝试从 "地点+时间。描述" 中拆分
			// 例：夜晚客厅（林婉清家）。暖黄灯光柔和...
			// 或：林婉清家客厅·夜晚。暖黄灯光...
			parts := strings.SplitN(content, "。", 2)
			locationTime := parts[0]
			if len(parts) > 1 {
				scene.Description = strings.TrimSpace(parts[1])
			}

			// 尝试用 "·" 分割地点和时间
			if idx := strings.Index(locationTime, "·"); idx >= 0 {
				scene.Location = strings.TrimSpace(locationTime[:idx])
				scene.Time = strings.TrimSpace(locationTime[idx+len("·"):])
			} else {
				// 尝试从开头提取时间词
				timeRe := regexp.MustCompile(`^(夜晚|清晨|午后|傍晚|深夜|凌晨|上午|下午|正午|黄昏)`)
				if m := timeRe.FindStringSubmatch(locationTime); m != nil {
					scene.Time = m[1]
					scene.Location = strings.TrimSpace(strings.TrimPrefix(locationTime, m[1]))
				} else {
					scene.Location = locationTime
					scene.Time = "未指定"
				}
			}

			scenes = append(scenes, scene)
		}
	}

	return scenes
}

// saveExtractedData 保存解析结果到数据库
func (s *ScriptGenerationService) saveExtractedData(dramaID uint, episodeID uint, result *ExtractResult) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// ===== 保存角色 =====
		var savedCharacters []models.Character
		for _, pc := range result.Characters {
			// 检查是否已存在
			var existing models.Character
			if err := tx.Where("drama_id = ? AND name = ?", dramaID, pc.Name).First(&existing).Error; err == nil {
				// 已存在 → 更新性格和外观（不覆盖图片等）
				updates := map[string]interface{}{}
				if pc.Personality != "" {
					updates["personality"] = pc.Personality
				}
				if pc.Appearance != "" {
					updates["appearance"] = pc.Appearance
				}
				if pc.Identity != "" {
					updates["description"] = pc.Identity
				}
				if len(updates) > 0 {
					tx.Model(&existing).Updates(updates)
				}
				savedCharacters = append(savedCharacters, existing)
				s.log.Infow("Updated existing character", "name", pc.Name, "drama_id", dramaID)
				continue
			}

			// 不存在 → 创建
			role := "main"
			char := models.Character{
				DramaID:     dramaID,
				Name:        pc.Name,
				Role:        &role,
				Description: &pc.Identity,
				Personality: &pc.Personality,
				Appearance:  &pc.Appearance,
			}
			if err := tx.Create(&char).Error; err != nil {
				s.log.Errorw("Failed to create character", "error", err, "name", pc.Name)
				continue
			}
			savedCharacters = append(savedCharacters, char)
			s.log.Infow("Created character from script", "name", pc.Name, "drama_id", dramaID)
		}

		// 加载章节用于多对多关联
		var episode models.Episode
		if episodeID > 0 {
			if err := tx.First(&episode, episodeID).Error; err != nil {
				s.log.Errorw("Failed to load episode for association", "error", err, "episode_id", episodeID)
			}
		}

		// 建立 episode_characters 关联
		if episode.ID > 0 && len(savedCharacters) > 0 {
			if err := tx.Model(&episode).Association("Characters").Replace(savedCharacters); err != nil {
				s.log.Errorw("Failed to associate characters", "error", err)
			}
		}

		// ===== 保存场景 =====
		var savedScenes []*models.Scene
		// 先清除该章节的场景关联（多对多）
		if episode.ID > 0 {
			if err := tx.Model(&episode).Association("Scenes").Clear(); err != nil {
				s.log.Errorw("Failed to clear scene associations", "error", err)
			}
		}

		for _, ps := range result.Scenes {
			scene := &models.Scene{
				DramaID:         dramaID,
				Location:        ps.Location,
				Time:            ps.Time,
				Prompt:          ps.Description,
				StoryboardCount: 1,
				Status:          "pending",
			}
			if err := tx.Create(scene).Error; err != nil {
				s.log.Errorw("Failed to create scene", "error", err, "location", ps.Location)
				continue
			}
			savedScenes = append(savedScenes, scene)
			s.log.Infow("Created scene from script", "location", ps.Location, "time", ps.Time)
		}

		// 将新场景关联到章节
		if episode.ID > 0 && len(savedScenes) > 0 {
			if err := tx.Model(&episode).Association("Scenes").Append(savedScenes); err != nil {
				s.log.Errorw("Failed to associate scenes with episode", "error", err)
			}
		}

		return nil
	})
}
