package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/ai"
	"github.com/cinemaker/backend/pkg/asr"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/video"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VideoAnalysisService struct {
	db        *gorm.DB
	log       *logger.Logger
	aiService *AIService
	workDir   string
}

func NewVideoAnalysisService(db *gorm.DB, log *logger.Logger, aiService *AIService, workDir string) *VideoAnalysisService {
	return &VideoAnalysisService{
		db:        db,
		log:       log,
		aiService: aiService,
		workDir:   workDir,
	}
}

// CreateTask creates a new video analysis task record.
func (s *VideoAnalysisService) CreateTask(videoPath, videoURL, title string, teamID uint) (*models.VideoAnalysisTask, error) {
	task := &models.VideoAnalysisTask{
		TaskID:    uuid.New().String(),
		TeamID:    teamID,
		VideoPath: videoPath,
		VideoURL:  videoURL,
		Title:     title,
		Status:    "pending",
		Progress:  0,
		Stage:     "created",
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("create video analysis task: %w", err)
	}
	return task, nil
}

// DownloadAndCreateTask downloads a video from URL, then creates a task.
func (s *VideoAnalysisService) DownloadAndCreateTask(videoURL string, teamID uint) (*models.VideoAnalysisTask, error) {
	taskID := uuid.New().String()
	downloadDir := filepath.Join(s.workDir, "video_analysis", taskID)

	task := &models.VideoAnalysisTask{
		TaskID:   taskID,
		TeamID:   teamID,
		VideoURL: videoURL,
		Status:   "downloading",
		Progress: 0,
		Stage:    "downloading",
	}
	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	go func() {
		result, err := video.DownloadVideo(videoURL, downloadDir)
		if err != nil {
			s.updateTaskError(task.ID, fmt.Sprintf("下载失败: %v", err))
			return
		}

		downloadData := map[string]interface{}{
			"download": map[string]interface{}{
				"status":    "done",
				"title":     result.Title,
				"file_path": result.FilePath,
				"duration":  result.Duration,
			},
		}
		downloadJSON, _ := json.Marshal(downloadData)

		s.db.Model(task).Updates(map[string]interface{}{
			"video_path":  result.FilePath,
			"title":       result.Title,
			"duration":    result.Duration,
			"status":      "processing",
			"progress":    10,
			"stage":       "downloaded",
			"stage_data":  downloadJSON,
		})

		s.ProcessVideo(task.ID, result.FilePath)
	}()

	return task, nil
}

// StartProcessing begins the async video analysis pipeline.
func (s *VideoAnalysisService) StartProcessing(taskID uint) {
	var task models.VideoAnalysisTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		s.log.Errorw("Task not found", "taskID", taskID, "error", err)
		return
	}
	go s.ProcessVideo(task.ID, task.VideoPath)
}

// ProcessVideo runs the full analysis pipeline.
func (s *VideoAnalysisService) ProcessVideo(taskID uint, videoPath string) {
	stageData := make(map[string]interface{})
	// Load existing stage_data (e.g. download info)
	var task models.VideoAnalysisTask
	if err := s.db.First(&task, taskID).Error; err == nil && len(task.StageData) > 0 {
		json.Unmarshal(task.StageData, &stageData)
	}

	s.updateTaskProgress(taskID, "detecting", 15, "正在检测镜头...")

	outputDir := filepath.Join(s.workDir, "video_analysis", fmt.Sprintf("task_%d", taskID))

	// Step 1: Scene detection + keyframe extraction + audio extraction
	sceneResult, err := video.DetectScenes(videoPath, 0.3, outputDir)
	if err != nil {
		s.updateTaskError(taskID, fmt.Sprintf("镜头检测失败: %v", err))
		return
	}

	// Save scene detection results
	shotTimestamps := make([]map[string]interface{}, len(sceneResult.Shots))
	for i, sh := range sceneResult.Shots {
		shotTimestamps[i] = map[string]interface{}{
			"index":      i,
			"start_time": sh.StartTime,
			"end_time":   sh.EndTime,
			"frame_url":  framePathToURL(sh.FramePath),
		}
	}
	stageData["detect"] = map[string]interface{}{
		"shot_count":  len(sceneResult.Shots),
		"duration":    sceneResult.Duration,
		"has_audio":   sceneResult.AudioPath != "",
		"shots":       shotTimestamps,
	}
	s.saveStageData(taskID, stageData)

	s.db.Model(&models.VideoAnalysisTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"shot_count": len(sceneResult.Shots),
		"duration":   sceneResult.Duration,
		"progress":   30,
		"stage":      "analyzing",
	})

	// Step 2: Parallel - ASR transcription + Vision analysis
	var wg sync.WaitGroup
	var sdMu sync.Mutex // protects stageData across goroutines
	var dialogues []asr.TranscriptSegment
	var dialogueErr error
	var frameDescs []shotDescription
	var frameAnalysisErr error

	saveSD := func() {
		s.saveStageData(taskID, stageData)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if sceneResult.AudioPath == "" {
			sdMu.Lock()
			stageData["transcribe"] = map[string]interface{}{"status": "skipped", "reason": "无音频轨道"}
			saveSD()
			sdMu.Unlock()
			return
		}

		dialogues, dialogueErr = s.transcribePerShot(taskID, videoPath, sceneResult.Shots, stageData, &sdMu)

		sdMu.Lock()
		if dialogueErr != nil {
			s.log.Errorw("ASR transcription failed", "error", dialogueErr)
		}
		sdMu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		frameDescs = s.analyzeFramesLive(taskID, sceneResult.Shots, stageData, &sdMu)

		sdMu.Lock()
		failCount := 0
		for _, d := range frameDescs {
			if strings.HasPrefix(d.Description, "[分析失败]") {
				failCount++
			}
		}
		if failCount == len(frameDescs) && len(frameDescs) > 0 {
			frameAnalysisErr = fmt.Errorf("全部 %d 帧画面分析失败", failCount)
		}
		sdMu.Unlock()
	}()

	wg.Wait()

	// If either path had a critical failure, stop the pipeline
	var failures []string
	if dialogueErr != nil {
		failures = append(failures, fmt.Sprintf("语音识别失败: %v", dialogueErr))
	}
	if frameAnalysisErr != nil {
		failures = append(failures, fmt.Sprintf("画面分析失败: %v", frameAnalysisErr))
	}
	if len(failures) > 0 {
		s.updateTaskError(taskID, strings.Join(failures, "；"))
		return
	}

	s.updateTaskProgress(taskID, "synthesizing", 75, "正在生成剧本...")

	// Step 3: LLM synthesizes everything into a structured script
	analysisResult, err := s.synthesizeScript(sceneResult, frameDescs, dialogues)
	if err != nil {
		s.updateTaskError(taskID, fmt.Sprintf("剧本生成失败: %v", err))
		return
	}

	resultJSON, _ := json.Marshal(analysisResult)

	stageData["synthesize"] = map[string]interface{}{"status": "done"}
	s.saveStageData(taskID, stageData)

	updates := map[string]interface{}{
		"status":   "done",
		"progress": 100,
		"stage":    "complete",
		"result":   resultJSON,
	}
	if analysisResult.Title != "" {
		updates["title"] = analysisResult.Title
	}
	s.db.Model(&models.VideoAnalysisTask{}).Where("id = ?", taskID).Updates(updates)

	s.log.Infow("Video analysis complete", "taskID", taskID, "shots", len(sceneResult.Shots))
}

func (s *VideoAnalysisService) saveStageData(taskID uint, stageData map[string]interface{}) {
	j, _ := json.Marshal(stageData)
	s.db.Model(&models.VideoAnalysisTask{}).Where("id = ?", taskID).Update("stage_data", j)
}

// framePathToURL converts a local frame path like "data/storage/video_analysis/task_3/frame_000.jpg"
// to a URL path like "/static/video_analysis/task_3/frame_000.jpg".
func framePathToURL(framePath string) string {
	const prefix = "data/storage/"
	if idx := strings.Index(framePath, prefix); idx >= 0 {
		return "/static/" + framePath[idx+len(prefix):]
	}
	return "/static/" + framePath
}

type shotDescription struct {
	Index       int    `json:"index"`
	Description string `json:"description"`
}

// analyzeFramesLive analyzes frames and updates stageData after each one completes.
func (s *VideoAnalysisService) analyzeFramesLive(
	taskID uint,
	shots []video.Shot,
	stageData map[string]interface{},
	sdMu *sync.Mutex,
) []shotDescription {
	descs := make([]shotDescription, len(shots))
	var mu sync.Mutex
	var wg sync.WaitGroup
	doneCount := 0

	sem := make(chan struct{}, 3)

	for i, shot := range shots {
		wg.Add(1)
		go func(idx int, sh video.Shot) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			desc, err := s.analyzeFrame(sh.FramePath)
			mu.Lock()
			if err != nil {
				descs[idx] = shotDescription{Index: idx, Description: fmt.Sprintf("[分析失败] %v", err)}
			} else {
				descs[idx] = shotDescription{Index: idx, Description: desc}
			}
			doneCount++
			currentDone := doneCount
			mu.Unlock()

			// Build current snapshot and save
			sdMu.Lock()
			framesData := make([]map[string]interface{}, len(shots))
			for j := range shots {
				frameURL := framePathToURL(shots[j].FramePath)
				mu.Lock()
				d := descs[j]
				mu.Unlock()
				if d.Description == "" {
					framesData[j] = map[string]interface{}{
						"index": j, "description": "", "frame_url": frameURL,
					}
				} else {
					framesData[j] = map[string]interface{}{
						"index": d.Index, "description": d.Description, "frame_url": frameURL,
					}
				}
			}
			stageData["analyze"] = map[string]interface{}{
				"status": "processing",
				"done":   currentDone,
				"total":  len(shots),
				"frames": framesData,
			}
			s.saveStageData(taskID, stageData)
			sdMu.Unlock()

			progress := 30 + int(float64(currentDone)/float64(len(shots))*40)
			s.updateTaskProgress(taskID, "analyzing", progress, fmt.Sprintf("分析第 %d/%d 帧...", currentDone, len(shots)))
		}(i, shot)
	}

	wg.Wait()

	// Final save with "done" status
	sdMu.Lock()
	framesData := make([]map[string]interface{}, len(descs))
	for i, d := range descs {
		frameURL := framePathToURL(shots[i].FramePath)
		framesData[i] = map[string]interface{}{
			"index": d.Index, "description": d.Description, "frame_url": frameURL,
		}
	}
	stageData["analyze"] = map[string]interface{}{
		"status": "done", "done": len(descs), "total": len(descs), "frames": framesData,
	}
	s.saveStageData(taskID, stageData)
	sdMu.Unlock()

	return descs
}

func (s *VideoAnalysisService) analyzeFrame(framePath string) (string, error) {
	imgData, err := os.ReadFile(framePath)
	if err != nil {
		return "", fmt.Errorf("read frame: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(framePath))
	mimeType := "image/jpeg"
	if ext == ".png" {
		mimeType = "image/png"
	}
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imgData))

	// Prefer "vision" service type; fall back to "text" and look for a vision-capable model
	visionConfig, _ := s.aiService.GetDefaultConfig("vision")
	if visionConfig == nil {
		visionConfig, _ = s.aiService.GetDefaultConfig("text")
	}
	if visionConfig == nil {
		return "", fmt.Errorf("no vision or text AI config available")
	}

	model := ""
	if len(visionConfig.Model) > 0 {
		model = visionConfig.Model[0]
	}
	for _, m := range visionConfig.Model {
		lower := strings.ToLower(m)
		if strings.Contains(lower, "vision") || strings.Contains(lower, "vl") {
			model = m
			break
		}
	}

	endpoint := visionConfig.Endpoint
	if endpoint == "" {
		endpoint = "/v1/chat/completions"
	}

	client := ai.NewOpenAIClient(visionConfig.BaseURL, visionConfig.APIKey, model, endpoint)

	prompt := `请用中文详细分析这一帧画面，重点关注以下五个维度：
1. 角色设计：人物外貌、发型、服装风格、表情、肢体语言
2. 道具设计：画面中出现的关键物品、道具、文字信息
3. 场景设计：场景类型（室内/室外）、具体环境、色调、光线、氛围
4. 镜头语言：构图方式（远景/中景/近景/特写）、拍摄角度
5. 剧情暗示：这个画面暗示了什么样的故事情节或情感状态
请直接输出描述，不超过300字。`

	desc, _, err := client.GenerateTextWithImages(prompt, []string{dataURL}, "", 500)
	if err != nil {
		return "", fmt.Errorf("vision analysis: %w", err)
	}

	return desc, nil
}

// transcribePerShot transcribes audio per-shot in parallel and updates stageData after each.
func (s *VideoAnalysisService) transcribePerShot(
	taskID uint,
	videoPath string,
	shots []video.Shot,
	stageData map[string]interface{},
	sdMu *sync.Mutex,
) ([]asr.TranscriptSegment, error) {
	asrCfg, err := s.getASRConfig()
	if err != nil {
		return nil, fmt.Errorf("ASR not configured: %w", err)
	}

	outputDir := filepath.Join(s.workDir, "video_analysis", fmt.Sprintf("task_%d", taskID), "audio_segments")
	os.MkdirAll(outputDir, 0755)

	results := make([]shotASRResult, len(shots))
	var mu sync.Mutex
	var asrWg sync.WaitGroup
	doneCount := 0
	sem := make(chan struct{}, 2) // 2 concurrent ASR tasks

	sdMu.Lock()
	stageData["transcribe"] = map[string]interface{}{
		"status": "processing", "done": 0, "total": len(shots),
		"message": fmt.Sprintf("语音识别 0/%d 镜头...", len(shots)),
		"shots":   make([]map[string]interface{}, len(shots)),
	}
	s.saveStageData(taskID, stageData)
	sdMu.Unlock()

	for i, shot := range shots {
		asrWg.Add(1)
		go func(idx int, sh video.Shot) {
			defer asrWg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			segPath := filepath.Join(outputDir, fmt.Sprintf("shot_%03d.wav", idx))
			if err := video.ExtractAudioSegment(videoPath, sh.StartTime, sh.EndTime, segPath); err != nil {
				s.log.Warnw("Skip ASR for short shot", "shot", idx, "error", err)
				mu.Lock()
				results[idx] = shotASRResult{ShotIndex: idx, Segments: nil}
				doneCount++
				mu.Unlock()
				s.updateTranscribeStageData(taskID, shots, results, doneCount, stageData, sdMu)
				return
			}

			client := asr.NewASRClient(asrCfg.appID, asrCfg.accessKey)
			client.ResourceID = asrCfg.resourceID

			segs, err := client.TranscribeFile(segPath)

			mu.Lock()
			if err != nil {
				results[idx] = shotASRResult{ShotIndex: idx, Err: err}
			} else {
				// Offset timestamps to video timeline
				for j := range segs {
					segs[j].StartTime += sh.StartTime
					segs[j].EndTime += sh.StartTime
				}
				results[idx] = shotASRResult{ShotIndex: idx, Segments: segs}
			}
			doneCount++
			currentDone := doneCount
			mu.Unlock()

			s.updateTranscribeStageData(taskID, shots, results, currentDone, stageData, sdMu)
		}(i, shot)
	}

	asrWg.Wait()

	// Collect all segments in order
	var allSegments []asr.TranscriptSegment
	fatalCount := 0
	for _, r := range results {
		if r.Err != nil {
			fatalCount++
			s.log.Warnw("ASR shot failed", "shot", r.ShotIndex, "error", r.Err)
		}
		allSegments = append(allSegments, r.Segments...)
	}

	// Final save
	sdMu.Lock()
	if fatalCount == len(shots) && len(shots) > 0 {
		stageData["transcribe"] = map[string]interface{}{"status": "failed", "error": "全部镜头语音识别失败"}
		s.saveStageData(taskID, stageData)
		sdMu.Unlock()
		return nil, fmt.Errorf("全部 %d 镜头语音识别失败", fatalCount)
	}
	sdMu.Unlock()

	return allSegments, nil
}

func (s *VideoAnalysisService) updateTranscribeStageData(
	taskID uint,
	shots []video.Shot,
	results []shotASRResult,
	doneCount int,
	stageData map[string]interface{},
	sdMu *sync.Mutex,
) {
	sdMu.Lock()
	defer sdMu.Unlock()

	shotsData := make([]map[string]interface{}, len(shots))
	for i, sh := range shots {
		r := results[i]
		entry := map[string]interface{}{
			"shot_index": i,
			"start_time": sh.StartTime,
			"end_time":   sh.EndTime,
		}
		if r.Segments != nil || r.Err != nil {
			if r.Err != nil {
				entry["status"] = "failed"
				entry["error"] = r.Err.Error()
			} else if len(r.Segments) == 0 {
				entry["status"] = "silent"
				entry["text"] = ""
			} else {
				texts := make([]string, len(r.Segments))
				for j, seg := range r.Segments {
					texts[j] = seg.Text
				}
				entry["status"] = "done"
				entry["text"] = strings.Join(texts, " ")
				segsData := make([]map[string]interface{}, len(r.Segments))
				for j, seg := range r.Segments {
					segsData[j] = map[string]interface{}{
						"start": seg.StartTime, "end": seg.EndTime, "text": seg.Text,
					}
				}
				entry["segments"] = segsData
			}
		} else {
			entry["status"] = "pending"
		}
		shotsData[i] = entry
	}

	status := "processing"
	if doneCount >= len(shots) {
		status = "done"
	}
	stageData["transcribe"] = map[string]interface{}{
		"status":  status,
		"done":    doneCount,
		"total":   len(shots),
		"message": fmt.Sprintf("语音识别 %d/%d 镜头", doneCount, len(shots)),
		"shots":   shotsData,
	}
	s.saveStageData(taskID, stageData)
}

type shotASRResult struct {
	ShotIndex int
	Segments  []asr.TranscriptSegment
	Err       error
}

type asrConfigInfo struct {
	appID      string
	accessKey  string
	resourceID string
}

func (s *VideoAnalysisService) getASRConfig() (*asrConfigInfo, error) {
	// Priority: environment variables > database
	appID := os.Getenv("ASR_APP_ID")
	accessKey := os.Getenv("ASR_ACCESS_KEY")
	resourceID := os.Getenv("ASR_RESOURCE_ID")

	if appID != "" && accessKey != "" {
		if resourceID == "" {
			resourceID = "volc.bigasr.auc"
		}
		return &asrConfigInfo{appID: appID, accessKey: accessKey, resourceID: resourceID}, nil
	}

	var config models.AIServiceConfig
	err := s.db.Where("service_type = ? AND is_active = ?", "asr", true).
		Order("priority DESC").First(&config).Error
	if err != nil {
		return nil, fmt.Errorf("no ASR config found: set ASR_APP_ID/ASR_ACCESS_KEY env vars or add asr config in settings")
	}
	resID := "volc.bigasr.auc"
	if config.Endpoint != "" {
		resID = config.Endpoint
	}
	return &asrConfigInfo{
		appID:      config.Provider,
		accessKey:  config.APIKey,
		resourceID: resID,
	}, nil
}

func (s *VideoAnalysisService) synthesizeScript(
	sceneResult *video.SceneDetectResult,
	frameDescs []shotDescription,
	dialogues []asr.TranscriptSegment,
) (*models.AnalysisResult, error) {

	var sb strings.Builder
	sb.WriteString("以下是一段视频的分析数据，请综合所有信息生成一个结构化的剧本。\n\n")

	sb.WriteString("## 镜头画面描述\n")
	for _, fd := range frameDescs {
		shot := sceneResult.Shots[fd.Index]
		sb.WriteString(fmt.Sprintf("镜头%d (%.1fs-%.1fs): %s\n", fd.Index+1, shot.StartTime, shot.EndTime, fd.Description))
	}

	if len(dialogues) > 0 {
		sb.WriteString("\n## 语音转写（对白）\n")
		for _, d := range dialogues {
			sb.WriteString(fmt.Sprintf("[%.1fs-%.1fs] %s\n", d.StartTime, d.EndTime, d.Text))
		}
	}

	systemPrompt := `你是一个专业的影视分镜师和视频内容分析师，擅长从画面分析数据中还原可直接用于AI视频生成的分镜剧本。
请根据提供的镜头画面描述和对白信息，输出结构化 JSON，分镜格式需要可以直接用于 Seedance 视频生成。

{
  "title": "剧本标题（根据内容推断一个合适的标题）",
  "summary": "2-3句话概述剧情、风格和主题",
  "tags": ["从以下预设标签中选取2-5个最匹配的"],
  "characters": [
    {"name": "角色名-造型名", "description": "详细外貌描述（发型、服装颜色款式、配饰、气质等，便于AI绘画还原）", "profession": "角色职业（如花艺师、摄影师、大学生等）", "role": "主角/配角/群演"}
  ],
  "shots": [
    {
      "index": 0,
      "start_time": 0.0,
      "end_time": 5.0,
      "title": "镜头标题（简短概括这个镜头的内容）",
      "description": "完整的画面内容描述",
      "location": "具体场景地点",
      "time": "时间（如清晨、午后、傍晚、夜晚等）",
      "characters": ["角色名-造型名"],
      "dialogue": "该镜头的对白",
      "mood": "情感氛围",
      "shot_type": "景别",
      "angle": "镜头角度",
      "movement": "运镜方式",
      "first_frame_desc": "首帧静态画面描述",
      "middle_action_desc": "中间动态过程描述",
      "last_frame_desc": "尾帧画面变化描述",
      "video_prompt": "Seedance视频生成提示词",
      "bgm_prompt": "配乐风格描述",
      "sound_effect": "音效描述"
    }
  ]
}

可选的 tags 标签（从中挑选最符合的2-5个）：
都市情感、古装仙侠、悬疑推理、喜剧搞笑、科幻未来、恐怖惊悚、青春校园、职场商战、
家庭伦理、动作冒险、文艺清新、纪录写实、美食探店、旅行Vlog、知识科普、开箱测评、
穿搭时尚、美妆护肤、健身运动、宠物日常、游戏实况、音乐舞蹈、手工DIY、亲子育儿

【镜头属性可选值 —— 必须严格使用以下枚举值】
- shot_type 景别（5选1）：远景、全景、中景、近景、特写
- angle 镜头角度（7选1）：平视、俯视、仰视、高机位、低机位、过肩视角、主观视角
- movement 运镜（12选1）：固定镜头、推、拉、摇、移、跟、升、降、甩、环绕、旋转、变焦

【分析要求】

1. 角色设计（重要）：
   - ⚠️ 角色必须有正式的中文姓名（如「沈知薇」「江屿」「陈樱」），禁止使用「女孩」「男孩」「女人」「男人」等泛称
   - 根据角色的气质、穿着、行为场景，为每个角色推断一个合理的职业（profession 字段），如花艺师、摄影师、设计师、大学生等
   - 同一个人物如果在不同镜头中穿着/造型明显不同，必须拆分为多个角色条目
   - 命名格式：「角色名-造型名」，例如「沈知薇-居家装」「沈知薇-出行装」
   - 每个角色条目的 description 只描述该造型下的具体外貌，便于AI绘画精准还原
   - role 字段保持一致；如果全程只有一种造型，不加后缀
   - shots.characters 中引用对应造型的完整名称

2. 三段描述规范（⚠️ 这是最重要的部分，直接用于AI视频生成）：
   - first_frame_desc：镜头起始的静态画面。景别构图、角色位置姿态表情、环境状态。纯静态，30-80字
   - middle_action_desc：从首帧到尾帧的动态过程。包含所有动作和对话。对话格式：用穿着特征引出说话人+"台词内容"。描述说话时表情和肢体语言。根据时长调整长度
   - last_frame_desc：基于首帧同一构图，描述经过中间过程后的变化。只写变化了什么，30-80字
   - ⚠️ 三段描述中绝对禁止使用角色名字，只能用穿着特征指代（如"穿黑色西装的短发女性"）

3. video_prompt 视频生成提示词：
   - 公式：主体 + 运动 + 环境（可选） + 运镜（可选） + 美学描述（可选）
   - 用外貌特征指定角色，不用名字
   - 对话直接写出台词，用外貌特征引出说话人
   - 运镜融入描述，包含光线、色调、声音细节
   - 4-6秒约100-200字，7-10秒约200-350字

4. 场景设计：准确推断每个镜头的场景地点和时间
5. 对话语言：从画面和字幕推断对白，分配给具体角色
6. 剧情变化：通过shots的顺序和mood体现完整的故事弧线
7. 只输出 JSON，不要其他内容`

	result, err := s.aiService.GenerateTextWithLog(sb.String(), systemPrompt, "video_analysis_synthesis", nil, ai.WithMaxTokens(8000))
	if err != nil {
		return nil, fmt.Errorf("LLM synthesis: %w", err)
	}

	result = cleanJSONResponse(result)

	var analysisResult models.AnalysisResult
	if err := json.Unmarshal([]byte(result), &analysisResult); err != nil {
		return nil, fmt.Errorf("parse LLM output: %w (raw: %s)", err, result[:min(len(result), 500)])
	}

	// Merge dialogue segments
	if len(dialogues) > 0 {
		for _, d := range dialogues {
			analysisResult.Dialogues = append(analysisResult.Dialogues, models.AnalysisLine{
				StartTime: d.StartTime,
				EndTime:   d.EndTime,
				Text:      d.Text,
			})
		}
	}

	return &analysisResult, nil
}

// GetTask retrieves a video analysis task by its UUID.
func (s *VideoAnalysisService) GetTask(taskID string) (*models.VideoAnalysisTask, error) {
	var task models.VideoAnalysisTask
	err := s.db.Where("task_id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTaskByID retrieves a task by its numeric ID.
func (s *VideoAnalysisService) GetTaskByID(id uint) (*models.VideoAnalysisTask, error) {
	var task models.VideoAnalysisTask
	err := s.db.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// ResynthesizeScript re-runs only the LLM synthesis step using cached stageData.
func (s *VideoAnalysisService) ResynthesizeScript(taskID string, includeAudio bool) error {
	task, err := s.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("task not found")
	}

	if len(task.StageData) == 0 {
		return fmt.Errorf("该任务没有阶段数据，请先完成视频分析")
	}

	var stageData map[string]interface{}
	if err := json.Unmarshal(task.StageData, &stageData); err != nil {
		return fmt.Errorf("解析阶段数据失败: %w", err)
	}

	// Reconstruct frameDescs from stageData["analyze"]["frames"]
	var frameDescs []shotDescription
	if analyze, ok := stageData["analyze"].(map[string]interface{}); ok {
		if frames, ok := analyze["frames"].([]interface{}); ok {
			for _, f := range frames {
				fm, ok := f.(map[string]interface{})
				if !ok {
					continue
				}
				idx := 0
				if v, ok := fm["index"].(float64); ok {
					idx = int(v)
				}
				desc := ""
				if v, ok := fm["description"].(string); ok {
					desc = v
				}
				frameDescs = append(frameDescs, shotDescription{Index: idx, Description: desc})
			}
		}
	}
	if len(frameDescs) == 0 {
		return fmt.Errorf("无画面分析数据，请先完成视频分析")
	}

	// Reconstruct shot timestamps from stageData["detect"]["shots"]
	var shots []video.Shot
	if detect, ok := stageData["detect"].(map[string]interface{}); ok {
		if shotList, ok := detect["shots"].([]interface{}); ok {
			for _, sh := range shotList {
				shMap, ok := sh.(map[string]interface{})
				if !ok {
					continue
				}
				var s video.Shot
				if v, ok := shMap["start_time"].(float64); ok {
					s.StartTime = v
				}
				if v, ok := shMap["end_time"].(float64); ok {
					s.EndTime = v
				}
				shots = append(shots, s)
			}
		}
	}
	sceneResult := &video.SceneDetectResult{Shots: shots}

	// Reconstruct dialogues from stageData["transcribe"] if requested
	var dialogues []asr.TranscriptSegment
	if includeAudio {
		if transcribe, ok := stageData["transcribe"].(map[string]interface{}); ok {
			if transcribeShots, ok := transcribe["shots"].([]interface{}); ok {
				for _, ts := range transcribeShots {
					tsMap, ok := ts.(map[string]interface{})
					if !ok {
						continue
					}
					if segments, ok := tsMap["segments"].([]interface{}); ok {
						for _, seg := range segments {
							segMap, ok := seg.(map[string]interface{})
							if !ok {
								continue
							}
							var d asr.TranscriptSegment
							if v, ok := segMap["start"].(float64); ok {
								d.StartTime = v
							}
							if v, ok := segMap["end"].(float64); ok {
								d.EndTime = v
							}
							if v, ok := segMap["text"].(string); ok {
								d.Text = v
							}
							if d.Text != "" {
								dialogues = append(dialogues, d)
							}
						}
					}
				}
			}
			// Fallback: top-level segments
			if len(dialogues) == 0 {
				if segments, ok := transcribe["segments"].([]interface{}); ok {
					for _, seg := range segments {
						segMap, ok := seg.(map[string]interface{})
						if !ok {
							continue
						}
						var d asr.TranscriptSegment
						if v, ok := segMap["start"].(float64); ok {
							d.StartTime = v
						}
						if v, ok := segMap["end"].(float64); ok {
							d.EndTime = v
						}
						if v, ok := segMap["text"].(string); ok {
							d.Text = v
						}
						if d.Text != "" {
							dialogues = append(dialogues, d)
						}
					}
				}
			}
		}
	}

	// Update task status
	s.updateTaskProgress(task.ID, "synthesizing", 75, "正在重新生成剧本...")

	analysisResult, err := s.synthesizeScript(sceneResult, frameDescs, dialogues)
	if err != nil {
		s.updateTaskError(task.ID, fmt.Sprintf("剧本生成失败: %v", err))
		return err
	}

	resultJSON, _ := json.Marshal(analysisResult)
	stageData["synthesize"] = map[string]interface{}{"status": "done"}
	s.saveStageData(task.ID, stageData)

	updates := map[string]interface{}{
		"status":   "done",
		"progress": 100,
		"stage":    "complete",
		"result":   resultJSON,
	}
	if analysisResult.Title != "" {
		updates["title"] = analysisResult.Title
	}
	s.db.Model(&models.VideoAnalysisTask{}).Where("id = ?", task.ID).Updates(updates)

	s.log.Infow("Re-synthesized script", "taskID", taskID, "includeAudio", includeAudio)
	return nil
}

// ListTasks returns recent video analysis tasks filtered by team.
func (s *VideoAnalysisService) ListTasks(limit int, teamID uint) ([]models.VideoAnalysisTask, error) {
	if limit <= 0 {
		limit = 20
	}
	var tasks []models.VideoAnalysisTask
	tx := s.db.Order("created_at DESC").Limit(limit)
	if teamID > 0 {
		tx = tx.Where("team_id = ? OR team_id = 0", teamID)
	}
	err := tx.Find(&tasks).Error
	return tasks, err
}

// ImportToDrama creates a CineMaker drama from analysis results.
func (s *VideoAnalysisService) ImportToDrama(taskID string, dramaTitle string, teamID uint) (*models.Drama, error) {
	task, err := s.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	if task.Status != "done" {
		return nil, fmt.Errorf("task not complete (status: %s)", task.Status)
	}

	var result models.AnalysisResult
	if err := json.Unmarshal(task.Result, &result); err != nil {
		return nil, fmt.Errorf("parse analysis result: %w", err)
	}

	if dramaTitle == "" {
		dramaTitle = result.Title
	}
	if dramaTitle == "" {
		dramaTitle = task.Title
	}

	tx := s.db.Begin()

	drama := &models.Drama{
		Title:         dramaTitle,
		Status:        "draft",
		TotalEpisodes: 1,
		Style:         "realistic",
	}
	if teamID > 0 {
		drama.TeamID = &teamID
	}
	if result.Summary != "" {
		drama.Description = &result.Summary
	}
	if err := tx.Create(drama).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create drama: %w", err)
	}

	// Build shot_index → frame_url map from stageData
	shotFrameURLs := make(map[int]string)
	if len(task.StageData) > 0 {
		var sd map[string]interface{}
		if err := json.Unmarshal(task.StageData, &sd); err == nil {
			if detect, ok := sd["detect"].(map[string]interface{}); ok {
				if shots, ok := detect["shots"].([]interface{}); ok {
					for _, sh := range shots {
						if shMap, ok := sh.(map[string]interface{}); ok {
							idxFloat, ok := shMap["index"].(float64)
							if !ok {
								continue
							}
							idx := int(idxFloat)
							if url, ok := shMap["frame_url"].(string); ok && url != "" {
								shotFrameURLs[idx] = url
							}
						}
					}
				}
			}
		}
	}

	// Collect reference frame URLs per character name
	charFrameURLs := make(map[string][]string)
	for _, shot := range result.Shots {
		frameURL := shotFrameURLs[shot.Index]
		if frameURL == "" {
			continue
		}
		for _, charName := range shot.Characters {
			if !containsStr(charFrameURLs[charName], frameURL) {
				charFrameURLs[charName] = append(charFrameURLs[charName], frameURL)
			}
		}
	}

	// Group characters by base name: "角色名-造型名" → base="角色名", outfit="造型名"
	type charEntry struct {
		baseName   string
		outfitName string
		fullName   string
		role       string
		desc       string
		profession string
	}
	var entries []charEntry
	baseNameGroups := make(map[string][]int) // baseName → indices in entries
	for i, ch := range result.Characters {
		base, outfit := parseCharacterOutfit(ch.Name)
		entries = append(entries, charEntry{
			baseName:   base,
			outfitName: outfit,
			fullName:   ch.Name,
			role:       ch.Role,
			desc:       ch.Description,
			profession: ch.Profession,
		})
		baseNameGroups[base] = append(baseNameGroups[base], i)
		_ = i
	}

	charMap := make(map[string]uint) // full character name → character ID (child or standalone)
	for baseName, indices := range baseNameGroups {
		if len(indices) == 1 && entries[indices[0]].outfitName == "" {
			// Single character without outfit suffix → standalone (no parent/child)
			e := entries[indices[0]]
			role := e.role
			character := &models.Character{
				DramaID:    drama.ID,
				Name:       e.fullName,
				Role:       &role,
				Appearance: &e.desc,
			}
			if e.profession != "" {
				character.Description = &e.profession
			}
			if urls, ok := charFrameURLs[e.fullName]; ok && len(urls) > 0 {
				refJSON, _ := json.Marshal(urls)
				character.ReferenceImages = refJSON
			}
			if err := tx.Create(character).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("create character: %w", err)
			}
			charMap[e.fullName] = character.ID
		} else {
			// Multiple outfits or single with explicit outfit → create parent + children
			firstEntry := entries[indices[0]]
			parentRole := firstEntry.role
			parent := &models.Character{
				DramaID: drama.ID,
				Name:    baseName,
				Role:    &parentRole,
			}
			if firstEntry.profession != "" {
				parent.Description = &firstEntry.profession
			}
			// Merge all child reference images as parent references
			var allParentRefs []string
			for _, idx := range indices {
				if urls, ok := charFrameURLs[entries[idx].fullName]; ok {
					allParentRefs = append(allParentRefs, urls...)
				}
			}
			if len(allParentRefs) > 0 {
				refJSON, _ := json.Marshal(allParentRefs)
				parent.ReferenceImages = refJSON
			}
			if err := tx.Create(parent).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("create parent character: %w", err)
			}

			for _, idx := range indices {
				e := entries[idx]
				outfitName := e.outfitName
				if outfitName == "" {
					outfitName = "默认造型"
				}
				childRole := e.role
				child := &models.Character{
					DramaID:    drama.ID,
					ParentID:   &parent.ID,
					Name:       e.fullName,
					OutfitName: &outfitName,
					Role:       &childRole,
					Appearance: &e.desc,
				}
				if urls, ok := charFrameURLs[e.fullName]; ok && len(urls) > 0 {
					refJSON, _ := json.Marshal(urls)
					child.ReferenceImages = refJSON
				}
				if err := tx.Create(child).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("create child character: %w", err)
				}
				charMap[e.fullName] = child.ID
			}
		}
	}

	// Create episode
	episodeTitle := "第1集"
	episode := &models.Episode{
		DramaID:    drama.ID,
		EpisodeNum: 1,
		Title:      episodeTitle,
		Status:     "draft",
	}
	if err := tx.Create(episode).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create episode: %w", err)
	}

	// Associate all characters with the episode
	if len(result.Characters) > 0 {
		var allChars []models.Character
		for _, id := range charMap {
			allChars = append(allChars, models.Character{ID: id})
		}
		if err := tx.Model(episode).Association("Characters").Append(&allChars); err != nil {
			s.log.Warnw("Failed to associate characters with episode", "error", err)
		}
	}

	// Create scenes and storyboards from shots
	locationMap := make(map[string]uint) // normalized location → scene ID
	locationNames := make(map[string]string) // normalized location → display name

	for i, shot := range result.Shots {
		loc := shot.Location
		if loc == "" {
			loc = "未知场景"
		}

		// Find matching scene by substring containment (merge similar locations)
		matchedKey := ""
		var sceneID uint
		for existingLoc, id := range locationMap {
			shorter, longer := existingLoc, loc
			if len([]rune(shorter)) > len([]rune(longer)) {
				shorter, longer = longer, shorter
			}
			if len([]rune(shorter)) >= 3 && strings.Contains(longer, shorter) {
				matchedKey = existingLoc
				sceneID = id
				break
			}
		}

		if matchedKey == "" {
			scene := &models.Scene{
				DramaID:   drama.ID,
				EpisodeID: &episode.ID,
				Location:  loc,
				Prompt:    shot.Description,
				Status:    "pending",
			}
			sceneName := loc
			scene.Name = &sceneName
			// Set keyframe as scene reference image
			if frameURL := shotFrameURLs[shot.Index]; frameURL != "" {
				refJSON, _ := json.Marshal([]string{frameURL})
				scene.ReferenceImages = refJSON
			}
			if err := tx.Create(scene).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("create scene: %w", err)
			}
			sceneID = scene.ID
			locationMap[loc] = sceneID
			locationNames[loc] = loc
		} else {
			// Use the shorter (more general) name as canonical key
			shorter := matchedKey
			if len([]rune(loc)) < len([]rune(matchedKey)) {
				shorter = loc
				delete(locationMap, matchedKey)
				locationMap[shorter] = sceneID
				locationNames[shorter] = shorter
			}
			_ = shorter
		}

		duration := int(shot.EndTime - shot.StartTime)
		if duration < 4 {
			duration = 4
		}
		storyboard := &models.Storyboard{
			EpisodeID:        episode.ID,
			SceneID:          &sceneID,
			StoryboardNumber: i + 1,
			Description:      &shot.Description,
			Dialogue:         &shot.Dialogue,
			Duration:         duration,
			Status:           "pending",
		}
		if shot.Title != "" {
			storyboard.Title = &shot.Title
		}
		if shot.Mood != "" {
			storyboard.Atmosphere = &shot.Mood
		}
		if shot.Location != "" {
			storyboard.Location = &shot.Location
		}
		if shot.Time != "" {
			storyboard.Time = &shot.Time
		}
		if shot.ShotType != "" {
			storyboard.ShotType = &shot.ShotType
		}
		if shot.Angle != "" {
			storyboard.Angle = &shot.Angle
		}
		if shot.Movement != "" {
			storyboard.Movement = &shot.Movement
		}
		if shot.FirstFrameDesc != "" {
			storyboard.FirstFrameDesc = &shot.FirstFrameDesc
		}
		if shot.MiddleActionDesc != "" {
			storyboard.MiddleActionDesc = &shot.MiddleActionDesc
		}
		if shot.LastFrameDesc != "" {
			storyboard.LastFrameDesc = &shot.LastFrameDesc
		}
		if shot.VideoPrompt != "" {
			storyboard.VideoPrompt = &shot.VideoPrompt
		}
		if shot.BgmPrompt != "" {
			storyboard.BgmPrompt = &shot.BgmPrompt
		}
		if shot.SoundEffect != "" {
			storyboard.SoundEffect = &shot.SoundEffect
		}
		if err := tx.Create(storyboard).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("create storyboard: %w", err)
		}

		// Associate characters with this storyboard
		if len(shot.Characters) > 0 {
			var sbChars []models.Character
			for _, name := range shot.Characters {
				if cid, ok := charMap[name]; ok {
					sbChars = append(sbChars, models.Character{ID: cid})
				}
			}
			if len(sbChars) > 0 {
				if err := tx.Model(storyboard).Association("Characters").Append(&sbChars); err != nil {
					s.log.Warnw("Failed to associate characters with storyboard", "error", err, "storyboard", i+1)
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	s.db.Model(&models.VideoAnalysisTask{}).Where("id = ?", task.ID).
		Update("imported_drama_id", drama.ID)

	s.log.Infow("Imported video analysis as drama",
		"taskID", taskID,
		"dramaID", drama.ID,
		"characters", len(result.Characters),
		"shots", len(result.Shots))

	return drama, nil
}

// RetryTask resets a failed/done task and re-runs the processing pipeline.
func (s *VideoAnalysisService) RetryTask(taskID string) (*models.VideoAnalysisTask, error) {
	task, err := s.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	if task.VideoPath == "" {
		return nil, fmt.Errorf("no video file available for retry")
	}

	if _, err := os.Stat(task.VideoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("video file no longer exists: %s", task.VideoPath)
	}

	s.db.Model(task).Updates(map[string]interface{}{
		"status":     "processing",
		"progress":   10,
		"stage":      "retrying",
		"error_msg":  "",
		"result":     nil,
		"stage_data": nil,
	})

	go s.ProcessVideo(task.ID, task.VideoPath)

	task.Status = "processing"
	task.Progress = 10
	task.Stage = "retrying"
	return task, nil
}

// DeleteTask soft-deletes a task and removes its storage directory.
func (s *VideoAnalysisService) DeleteTask(taskID string) error {
	var task models.VideoAnalysisTask
	if err := s.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	storageDir := filepath.Join(s.workDir, "video_analysis", fmt.Sprintf("task_%d", task.ID))
	os.RemoveAll(storageDir)

	if task.VideoPath != "" {
		dir := filepath.Dir(task.VideoPath)
		if dir != "." && dir != "/" {
			os.RemoveAll(dir)
		}
	}

	if err := s.db.Delete(&task).Error; err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

func (s *VideoAnalysisService) updateTaskProgress(taskID uint, stage string, progress int, msg string) {
	s.db.Model(&models.VideoAnalysisTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"stage":    stage,
		"progress": progress,
	})
}

func (s *VideoAnalysisService) updateTaskError(taskID uint, errMsg string) {
	s.db.Model(&models.VideoAnalysisTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":    "failed",
		"error_msg": errMsg,
	})
	s.log.Errorw("Video analysis failed", "taskID", taskID, "error", errMsg)
}

// parseCharacterOutfit splits "角色名-造型名" into (baseName, outfitName).
// If no separator is found, returns (fullName, "").
func parseCharacterOutfit(name string) (string, string) {
	separators := []string{"-", "—", "·", "－"}
	for _, sep := range separators {
		if idx := strings.Index(name, sep); idx > 0 && idx < len(name)-len(sep) {
			return strings.TrimSpace(name[:idx]), strings.TrimSpace(name[idx+len(sep):])
		}
	}
	return name, ""
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func cleanJSONResponse(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	}
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

