package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cinemaker/backend/domain"
	models "github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/infrastructure/external/ffmpeg"
	"github.com/cinemaker/backend/infrastructure/storage"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/utils"
	"github.com/cinemaker/backend/pkg/video"
	"gorm.io/gorm"
)

type VideoGenerationService struct {
	db               *gorm.DB
	transferService  *ResourceTransferService
	log              *logger.Logger
	localStorage     storage.Storage
	aiService        *AIService
	ffmpeg           *ffmpeg.FFmpeg
	promptI18n       *PromptI18n
	promptTranslator *PromptTranslator
}

func NewVideoGenerationService(db *gorm.DB, transferService *ResourceTransferService, localStorage storage.Storage, aiService *AIService, log *logger.Logger, promptI18n *PromptI18n) *VideoGenerationService {
	service := &VideoGenerationService{
		db:               db,
		localStorage:     localStorage,
		transferService:  transferService,
		aiService:        aiService,
		log:              log,
		ffmpeg:           ffmpeg.NewFFmpeg(log),
		promptI18n:       promptI18n,
		promptTranslator: NewPromptTranslator(aiService, log),
	}

	go service.RecoverPendingTasks()

	return service
}

type GenerateVideoRequest struct {
	StoryboardID *uint  `json:"storyboard_id"`
	DramaID      string `json:"drama_id" binding:"required"`
	ImageGenID   *uint  `json:"image_gen_id"`

	// 参考图模式：single, first_last, multiple, none
	ReferenceMode string `json:"reference_mode"`

	// 单图模式
	ImageURL       string  `json:"image_url"`
	ImageLocalPath *string `json:"image_local_path"` // 单图模式的本地路径

	// 首尾帧模式
	FirstFrameURL       *string `json:"first_frame_url"`
	FirstFrameLocalPath *string `json:"first_frame_local_path"` // 首帧本地路径
	LastFrameURL        *string `json:"last_frame_url"`
	LastFrameLocalPath  *string `json:"last_frame_local_path"` // 尾帧本地路径

	// 多图模式
	ReferenceImageURLs []string `json:"reference_image_urls"`

	Prompt       string  `json:"prompt" binding:"required,min=5,max=2000"`
	Provider     string  `json:"provider"`
	Model        string  `json:"model"`
	Duration     *int    `json:"duration"`
	FPS          *int    `json:"fps"`
	AspectRatio  *string `json:"aspect_ratio"`
	Style        *string `json:"style"`
	MotionLevel  *int    `json:"motion_level"`
	CameraMotion *string `json:"camera_motion"`
	Seed         *int64  `json:"seed"`
	Resolution   *string `json:"resolution"` // 视频清晰度：480p 或 720p

	GenerateAudio  *bool `json:"generate_audio"`  // 是否生成音频（默认：true）
	EnableSubtitle *bool `json:"enable_subtitle"` // 是否生成字幕（默认：false，通过提示词控制）
}

func (s *VideoGenerationService) GenerateVideo(request *GenerateVideoRequest) (*models.VideoGeneration, error) {
	if request.StoryboardID != nil {
		var storyboard models.Storyboard
		if err := s.db.Preload("Episode").Where("id = ?", *request.StoryboardID).First(&storyboard).Error; err != nil {
			return nil, fmt.Errorf("storyboard not found")
		}
		if fmt.Sprintf("%d", storyboard.Episode.DramaID) != request.DramaID {
			return nil, fmt.Errorf("storyboard does not belong to drama")
		}
	}

	if request.ImageGenID != nil {
		var imageGen models.ImageGeneration
		if err := s.db.Where("id = ?", *request.ImageGenID).First(&imageGen).Error; err != nil {
			return nil, fmt.Errorf("image generation not found")
		}
	}

	provider := request.Provider
	if provider == "" {
		provider = "doubao"
	}

	dramaID, _ := strconv.ParseUint(request.DramaID, 10, 32)

	videoGen := &models.VideoGeneration{
		StoryboardID: request.StoryboardID,
		DramaID:      uint(dramaID),
		ImageGenID:   request.ImageGenID,
		Provider:     provider,
		Prompt:       request.Prompt,
		Model:        request.Model,
		Duration:     request.Duration,
		FPS:          request.FPS,
		AspectRatio:  request.AspectRatio,
		Style:        request.Style,
		MotionLevel:  request.MotionLevel,
		CameraMotion: request.CameraMotion,
		Seed:         request.Seed,
		Status:       models.VideoStatusPending,
	}

	generateAudio := false
	if request.GenerateAudio != nil {
		generateAudio = *request.GenerateAudio
	}
	videoGen.GenerateAudio = &generateAudio

	enableSubtitle := false
	if request.EnableSubtitle != nil {
		enableSubtitle = *request.EnableSubtitle
	}
	videoGen.EnableSubtitle = &enableSubtitle

	// 根据参考图模式处理不同的参数
	if request.ReferenceMode != "" {
		videoGen.ReferenceMode = &request.ReferenceMode
	}

	switch request.ReferenceMode {
	case "single":
		// 单图模式 - 优先使用 COS URL（远程 URL 更稳定，避免 base64 体积过大）
		if request.ImageURL != "" {
			videoGen.ImageURL = &request.ImageURL
		} else if request.ImageLocalPath != nil && *request.ImageLocalPath != "" {
			videoGen.ImageURL = request.ImageLocalPath
		}
	case "first_last":
		// 首尾帧模式 - 优先使用 COS URL
		if request.FirstFrameURL != nil && *request.FirstFrameURL != "" {
			videoGen.FirstFrameURL = request.FirstFrameURL
		} else if request.FirstFrameLocalPath != nil && *request.FirstFrameLocalPath != "" {
			videoGen.FirstFrameURL = request.FirstFrameLocalPath
		}
		if request.LastFrameURL != nil && *request.LastFrameURL != "" {
			videoGen.LastFrameURL = request.LastFrameURL
		} else if request.LastFrameLocalPath != nil && *request.LastFrameLocalPath != "" {
			videoGen.LastFrameURL = request.LastFrameLocalPath
		}
	case "multiple":
		// 多图模式
		if len(request.ReferenceImageURLs) > 0 {
			referenceImagesJSON, err := json.Marshal(request.ReferenceImageURLs)
			if err == nil {
				referenceImagesStr := string(referenceImagesJSON)
				videoGen.ReferenceImageURLs = &referenceImagesStr
			}
		}
	case "none":
		// 无参考图，纯文本生成
	default:
		// 向后兼容：如果没有指定模式，根据提供的参数自动判断
		if request.ImageURL != "" {
			videoGen.ImageURL = &request.ImageURL
			mode := "single"
			videoGen.ReferenceMode = &mode
		} else if request.FirstFrameURL != nil || request.LastFrameURL != nil {
			videoGen.FirstFrameURL = request.FirstFrameURL
			videoGen.LastFrameURL = request.LastFrameURL
			mode := "first_last"
			videoGen.ReferenceMode = &mode
		} else if len(request.ReferenceImageURLs) > 0 {
			referenceImagesJSON, err := json.Marshal(request.ReferenceImageURLs)
			if err == nil {
				referenceImagesStr := string(referenceImagesJSON)
				videoGen.ReferenceImageURLs = &referenceImagesStr
				mode := "multiple"
				videoGen.ReferenceMode = &mode
			}
		}
	}

	if err := s.db.Create(videoGen).Error; err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	// Start background goroutine to process video generation asynchronously
	// This allows the API to return immediately while video generation happens in background
	// CRITICAL: The goroutine will handle all video generation logic including API calls and polling
	go s.ProcessVideoGeneration(videoGen.ID)

	return videoGen, nil
}

func (s *VideoGenerationService) ProcessVideoGeneration(videoGenID uint) {
	var videoGen models.VideoGeneration
	if err := s.db.First(&videoGen, videoGenID).Error; err != nil {
		s.log.Errorw("Failed to load video generation", "error", err, "id", videoGenID)
		return
	}

	// 获取drama的style信息
	var drama models.Drama
	if err := s.db.First(&drama, videoGen.DramaID).Error; err != nil {
		s.log.Warnw("Failed to load drama for style", "error", err, "drama_id", videoGen.DramaID)
	}

	s.db.Model(&videoGen).Update("status", models.VideoStatusProcessing)

	client, vidClientInfo, err := s.getVideoClient(videoGen.Provider, videoGen.Model)
	if err != nil {
		s.log.Errorw("Failed to get video client", "error", err, "provider", videoGen.Provider, "model", videoGen.Model)
		s.updateVideoGenError(videoGenID, err.Error())
		return
	}

	s.log.Infow("Starting video generation", "id", videoGenID, "prompt", videoGen.Prompt, "provider", videoGen.Provider)

	var opts []video.VideoOption
	if videoGen.Model != "" {
		opts = append(opts, video.WithModel(videoGen.Model))
	}
	if videoGen.Duration != nil {
		opts = append(opts, video.WithDuration(*videoGen.Duration))
	}
	if videoGen.FPS != nil {
		opts = append(opts, video.WithFPS(*videoGen.FPS))
	}
	if videoGen.Resolution != nil {
		opts = append(opts, video.WithResolution(*videoGen.Resolution))
	}
	if videoGen.AspectRatio != nil {
		opts = append(opts, video.WithAspectRatio(*videoGen.AspectRatio))
	}
	if videoGen.Style != nil {
		opts = append(opts, video.WithStyle(*videoGen.Style))
	}
	if videoGen.MotionLevel != nil {
		opts = append(opts, video.WithMotionLevel(*videoGen.MotionLevel))
	}
	if videoGen.CameraMotion != nil {
		opts = append(opts, video.WithCameraMotion(*videoGen.CameraMotion))
	}
	if videoGen.Seed != nil {
		opts = append(opts, video.WithSeed(*videoGen.Seed))
	}

	// generate_audio: 优先用请求值，否则用 Drama 配置
	generateAudio := false
	if videoGen.GenerateAudio != nil {
		generateAudio = *videoGen.GenerateAudio
	} else if drama.GenerateAudio != nil {
		generateAudio = *drama.GenerateAudio
	}
	opts = append(opts, video.WithGenerateAudio(generateAudio))

	// 根据参考图模式添加相应的选项
	// 优先使用远程 URL（COS）直接传给 Seedance，仅本地路径才转 base64
	if videoGen.ReferenceMode != nil {
		switch *videoGen.ReferenceMode {
		case "first_last":
			if videoGen.FirstFrameURL != nil {
				firstFrameURL := s.resolveImageURL(*videoGen.FirstFrameURL)
				opts = append(opts, video.WithFirstFrame(firstFrameURL))
				s.log.Infow("First frame for video generation", "url_type", s.urlType(*videoGen.FirstFrameURL))
			}
			if videoGen.LastFrameURL != nil {
				lastFrameURL := s.resolveImageURL(*videoGen.LastFrameURL)
				opts = append(opts, video.WithLastFrame(lastFrameURL))
				s.log.Infow("Last frame for video generation", "url_type", s.urlType(*videoGen.LastFrameURL))
			}
		case "multiple":
			if videoGen.ReferenceImageURLs != nil {
				var imageURLs []string
				if err := json.Unmarshal([]byte(*videoGen.ReferenceImageURLs), &imageURLs); err == nil {
					var resolvedURLs []string
					for _, imgURL := range imageURLs {
						resolvedURLs = append(resolvedURLs, s.resolveImageURL(imgURL))
					}
					opts = append(opts, video.WithReferenceImages(resolvedURLs))
				}
			}
		}
	}

	// 构造imageURL参数（单图模式使用，其他模式传空字符串）
	imageURL := ""
	if videoGen.ImageURL != nil {
		imageURL = s.resolveImageURL(*videoGen.ImageURL)
	}

	// 构建完整的提示词：风格提示词 + 约束提示词 + 用户提示词
	prompt := videoGen.Prompt

	// 2. 添加视频约束提示词
	// 根据参考图模式选择对应的约束提示词
	referenceMode := "single" // 默认单图模式
	if videoGen.ReferenceMode != nil {
		referenceMode = *videoGen.ReferenceMode
	}

	constraintPrompt := s.promptI18n.GetVideoConstraintPrompt(referenceMode)

	constraintPrompt += "\n\n- 视频起始画面必须与传入的参考图片保持一致，人物位置和场景不要突变"
	constraintPrompt += "\n\n- 绝对不要在视频画面中生成任何文字、字幕、标题或水印"

	if constraintPrompt != "" {
		prompt = constraintPrompt + "\n\n" + prompt
		s.log.Infow("Added constraint prompt to video generation",
			"id", videoGenID,
			"reference_mode", referenceMode,
			"constraint_prompt_length", len(constraintPrompt))
	}

	s.log.Infow("Final prompt for video generation (Chinese, no translation for Seedance 1.5 Pro)",
		"id", videoGenID,
		"user_prompt", videoGen.Prompt,
		"constraint_prompt_length", len(constraintPrompt),
		"final_prompt_length", len(prompt),
		"prompt_preview", truncateString(prompt, 200))

	dramaID := videoGen.DramaID
	msgLog := s.aiService.GetMessageLogService()
	vLogID := msgLog.LogRequest(LogEntry{
		DramaID:     &dramaID,
		ServiceType: "video",
		Purpose:     "generate_video",
		Provider:    vidClientInfo.Provider,
		Model:       vidClientInfo.Model,
		UserPrompt:  prompt,
		FullRequest: map[string]interface{}{
			"video_gen_id":   videoGenID,
			"image_url":      truncateString(imageURL, 100),
			"duration":       videoGen.Duration,
			"fps":            videoGen.FPS,
			"resolution":     videoGen.Resolution,
			"reference_mode": videoGen.ReferenceMode,
		},
	})

	vidStart := time.Now()
	result, err := client.GenerateVideo(imageURL, prompt, opts...)
	vidElapsed := time.Since(vidStart).Milliseconds()
	if err != nil {
		msgLog.UpdateFailed(vLogID, err.Error(), vidElapsed)
		s.log.Errorw("Video generation API call failed", "error", err, "id", videoGenID)
		s.updateVideoGenError(videoGenID, err.Error())
		return
	}
	msgLog.UpdateSuccess(vLogID, fmt.Sprintf("task_id=%s", result.TaskID), vidElapsed, 0, 0)

	// CRITICAL FIX: Validate TaskID before starting polling goroutine
	// Empty TaskID would cause polling to fail silently or cause issues
	if result.TaskID != "" {
		s.db.Model(&videoGen).Updates(map[string]interface{}{
			"task_id": result.TaskID,
			"status":  models.VideoStatusProcessing,
		})
		// Start background goroutine to poll task status
		// This allows the API to return immediately while video generation continues asynchronously
		// The goroutine will poll until completion, failure, or timeout (max 300 attempts * 10s = 50 minutes)
		go s.pollTaskStatus(videoGenID, result.TaskID, videoGen.Provider, videoGen.Model)
		return
	}

	if result.VideoURL != "" {
		s.completeVideoGeneration(videoGenID, result.VideoURL, &result.Duration, &result.Width, &result.Height, nil)
		return
	}

	s.updateVideoGenError(videoGenID, "no task ID or video URL returned")
}

func (s *VideoGenerationService) pollTaskStatus(videoGenID uint, taskID string, provider string, model string) {
	// CRITICAL FIX: Validate taskID parameter to prevent invalid API calls
	// Empty taskID would cause unnecessary API calls and potential errors
	if taskID == "" {
		s.log.Errorw("Invalid empty taskID for polling", "video_gen_id", videoGenID)
		s.updateVideoGenError(videoGenID, "invalid task ID for polling")
		return
	}

	client, _, err := s.getVideoClient(provider, model)
	if err != nil {
		s.log.Errorw("Failed to get video client for polling", "error", err)
		s.updateVideoGenError(videoGenID, "failed to get video client")
		return
	}

	// Polling configuration: max 300 attempts with 10 second intervals
	// Total maximum polling time: 300 * 10s = 50 minutes
	// This prevents infinite polling if the task never completes
	maxAttempts := 300
	interval := 10 * time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Sleep before each poll attempt to avoid overwhelming the API
		// First iteration sleeps before the first check (after 0 attempts)
		time.Sleep(interval)

		var videoGen models.VideoGeneration
		if err := s.db.First(&videoGen, videoGenID).Error; err != nil {
			s.log.Errorw("Failed to load video generation", "error", err, "id", videoGenID)
			return
		}

		// CRITICAL FIX: Check if status was manually changed (e.g., cancelled by user)
		// If status is no longer "processing", stop polling to avoid unnecessary API calls
		// This prevents polling when the task has been cancelled or failed externally
		if videoGen.Status != models.VideoStatusProcessing {
			s.log.Infow("Video generation status changed, stopping poll", "id", videoGenID, "status", videoGen.Status)
			return
		}

		// Poll the video generation API for task status
		// Continue polling on transient errors (network issues, temporary API failures)
		// Only stop on permanent errors or task completion
		result, err := client.GetTaskStatus(taskID)
		if err != nil {
			s.log.Errorw("Failed to get task status", "error", err, "task_id", taskID, "attempt", attempt+1)
			// Continue polling on error - might be transient network issue
			// Will eventually timeout after maxAttempts if error persists
			continue
		}

		// Check if task completed successfully
		// CRITICAL FIX: Validate that video URL exists when task is marked as completed
		// Some APIs may mark task as completed but fail to provide the video URL
		if result.Completed {
			if result.VideoURL != "" {
				// Successfully completed with video URL - download and update database
				s.completeVideoGeneration(videoGenID, result.VideoURL, &result.Duration, &result.Width, &result.Height, nil)
				return
			}
			// Task marked as completed but no video URL - this is an error condition
			s.updateVideoGenError(videoGenID, "task completed but no video URL")
			return
		}

		// Check if task failed with an error message
		if result.Error != "" {
			s.updateVideoGenError(videoGenID, result.Error)
			return
		}

		// Task still in progress - log and continue polling
		s.log.Infow("Video generation in progress", "id", videoGenID, "attempt", attempt+1, "max_attempts", maxAttempts)
	}

	// CRITICAL FIX: Handle polling timeout gracefully
	// After maxAttempts (50 minutes), mark task as failed if still not completed
	// This prevents indefinite polling and resource waste
	s.updateVideoGenError(videoGenID, fmt.Sprintf("polling timeout after %d attempts (%.1f minutes)", maxAttempts, float64(maxAttempts*int(interval))/60.0))
}

func (s *VideoGenerationService) completeVideoGeneration(videoGenID uint, videoURL string, duration *int, width *int, height *int, firstFrameURL *string) {
	var localVideoPath *string

	// 下载视频到本地存储/COS并保存相对路径到数据库
	if s.localStorage != nil && videoURL != "" {
		downloadResult, err := s.localStorage.DownloadFromURLWithPath(videoURL, "videos")
		if err != nil {
			s.log.Warnw("Failed to download video to local storage",
				"error", err,
				"id", videoGenID,
				"original_url", videoURL)
		} else {
			localVideoPath = &downloadResult.RelativePath
			// 用持久化存储 URL (COS) 替换临时 URL（豆包 TOS 链接 24h 过期）
			videoURL = downloadResult.URL
			s.log.Infow("Video downloaded to storage",
				"id", videoGenID,
				"storage_url", videoURL,
				"local_path", downloadResult.RelativePath)
		}
	}

	// 如果视频已下载到本地，探测真实时长
	// 特别是当 AI 服务返回的 duration 为 0 或 nil 时，必须探测
	shouldProbe := localVideoPath != nil && s.ffmpeg != nil && (duration == nil || *duration == 0)
	if shouldProbe {
		absPath := s.localStorage.GetAbsolutePath(*localVideoPath)
		if probedDuration, err := s.ffmpeg.GetVideoDuration(absPath); err == nil {
			// 转换为整数秒（向上取整）
			durationInt := int(probedDuration + 0.5)
			duration = &durationInt
			s.log.Infow("Probed video duration (was 0 or nil)",
				"id", videoGenID,
				"duration_seconds", durationInt,
				"duration_float", probedDuration)
		} else {
			s.log.Errorw("Failed to probe video duration, duration will be 0",
				"error", err,
				"id", videoGenID,
				"local_path", *localVideoPath)
		}
	} else if localVideoPath != nil && s.ffmpeg != nil && duration != nil && *duration > 0 {
		// 即使有 duration，也验证一下（可选）
		absPath := s.localStorage.GetAbsolutePath(*localVideoPath)
		if probedDuration, err := s.ffmpeg.GetVideoDuration(absPath); err == nil {
			durationInt := int(probedDuration + 0.5)
			if durationInt != *duration {
				s.log.Warnw("Probed duration differs from provided duration",
					"id", videoGenID,
					"provided", *duration,
					"probed", durationInt)
				// 使用探测到的时长（更准确）
				duration = &durationInt
			}
		}
	}

	// 下载首帧图片到COS/本地存储，用持久化URL替换临时URL
	if firstFrameURL != nil && *firstFrameURL != "" && s.localStorage != nil {
		downloadResult, err := s.localStorage.DownloadFromURLWithPath(*firstFrameURL, "video_frames")
		if err != nil {
			s.log.Warnw("Failed to download first frame to storage",
				"error", err,
				"id", videoGenID,
				"original_url", *firstFrameURL)
		} else {
			*firstFrameURL = downloadResult.URL
			s.log.Infow("First frame downloaded to storage",
				"id", videoGenID,
				"storage_url", downloadResult.URL,
				"local_path", downloadResult.RelativePath)
		}
	}

	// 数据库中保存原始URL和本地路径
	updates := map[string]interface{}{
		"status":     models.VideoStatusCompleted,
		"video_url":  videoURL,
		"local_path": localVideoPath,
	}
	// 只有当 duration 大于 0 时才保存，避免保存无效的 0 值
	if duration != nil && *duration > 0 {
		updates["duration"] = *duration
	}
	if width != nil {
		updates["width"] = *width
	}
	if height != nil {
		updates["height"] = *height
	}
	if firstFrameURL != nil {
		updates["first_frame_url"] = *firstFrameURL
	}

	if err := s.db.Model(&models.VideoGeneration{}).Where("id = ?", videoGenID).Updates(updates).Error; err != nil {
		s.log.Errorw("Failed to update video generation", "error", err, "id", videoGenID)
		return
	}

	var videoGen models.VideoGeneration
	if err := s.db.First(&videoGen, videoGenID).Error; err == nil {
		if videoGen.StoryboardID != nil {
			// 更新 Storyboard 的 video_url 和 duration
			storyboardUpdates := map[string]interface{}{
				"video_url": videoURL,
			}
			// 只有当 duration 大于 0 时才更新，避免用无效的 0 值覆盖
			if duration != nil && *duration > 0 {
				storyboardUpdates["duration"] = *duration
			}
			if err := s.db.Model(&models.Storyboard{}).Where("id = ?", *videoGen.StoryboardID).Updates(storyboardUpdates).Error; err != nil {
				s.log.Warnw("Failed to update storyboard", "storyboard_id", *videoGen.StoryboardID, "error", err)
			} else {
				s.log.Infow("Updated storyboard with video info", "storyboard_id", *videoGen.StoryboardID, "duration", duration)
			}
		}
	}

	s.log.Infow("Video generation completed", "id", videoGenID, "url", videoURL, "duration", duration)

	// 自动创建 asset 记录，让视频出现在素材库中
	s.createVideoAsset(videoGenID)

	// 自动截取尾帧并保存到数据库
	s.autoExtractLastFrame(videoGenID)
}

// createVideoAsset 在视频生成完成后自动创建 asset 记录
func (s *VideoGenerationService) createVideoAsset(videoGenID uint) {
	var videoGen models.VideoGeneration
	if err := s.db.Preload("Storyboard").First(&videoGen, videoGenID).Error; err != nil {
		s.log.Warnw("Failed to load video generation for asset creation", "error", err, "id", videoGenID)
		return
	}

	if videoGen.VideoURL == nil || *videoGen.VideoURL == "" {
		return
	}

	// 检查是否已有该视频的 asset 记录（避免重复）
	var existingCount int64
	s.db.Model(&models.Asset{}).Where("video_gen_id = ?", videoGenID).Count(&existingCount)
	if existingCount > 0 {
		s.log.Infow("Asset already exists for video generation, skipping", "video_gen_id", videoGenID)
		return
	}

	dramaID := videoGen.DramaID

	var drama models.Drama
	var teamID *uint
	if err := s.db.Select("id, team_id").Where("id = ?", dramaID).First(&drama).Error; err == nil && drama.TeamID != nil {
		teamID = drama.TeamID
	}

	var episodeID *uint
	var storyboardNum *int
	if videoGen.Storyboard != nil {
		episodeID = &videoGen.Storyboard.EpisodeID
		storyboardNum = &videoGen.Storyboard.StoryboardNumber
	}

	asset := &models.Asset{
		Name:          fmt.Sprintf("Video_%d", videoGen.ID),
		Type:          models.AssetTypeVideo,
		URL:           *videoGen.VideoURL,
		LocalPath:     videoGen.LocalPath,
		DramaID:       &dramaID,
		TeamID:        teamID,
		EpisodeID:     episodeID,
		StoryboardID:  videoGen.StoryboardID,
		StoryboardNum: storyboardNum,
		VideoGenID:    &videoGenID,
		Duration:      videoGen.Duration,
		Width:         videoGen.Width,
		Height:        videoGen.Height,
	}

	if videoGen.FirstFrameURL != nil {
		asset.ThumbnailURL = videoGen.FirstFrameURL
	}

	if err := s.db.Create(asset).Error; err != nil {
		s.log.Errorw("Failed to create asset for video generation", "error", err, "video_gen_id", videoGenID)
	} else {
		s.log.Infow("Auto-created asset for completed video", "video_gen_id", videoGenID, "asset_id", asset.ID)
	}
}

func (s *VideoGenerationService) updateVideoGenError(videoGenID uint, errorMsg string) {
	// 对敏感内容错误进行友好提示
	userFriendlyMsg := s.formatUserFriendlyError(errorMsg)

	if err := s.db.Model(&models.VideoGeneration{}).Where("id = ?", videoGenID).Updates(map[string]interface{}{
		"status":    models.VideoStatusFailed,
		"error_msg": userFriendlyMsg,
	}).Error; err != nil {
		s.log.Errorw("Failed to update video generation error", "error", err, "id", videoGenID)
	}
}

// formatUserFriendlyError 将API错误转换为用户友好的提示
func (s *VideoGenerationService) formatUserFriendlyError(errorMsg string) string {
	// 检测敏感内容相关错误
	sensitiveContentKeywords := []string{
		"OutputVideoSensitiveContentDetected",
		"OutputImageSensitiveContentDetected",
		"sensitive content",
		"sensitive information",
		"敏感内容",
		"敏感信息",
		"content safety",
		"safety check",
	}

	lowerErrorMsg := strings.ToLower(errorMsg)
	for _, keyword := range sensitiveContentKeywords {
		if strings.Contains(lowerErrorMsg, strings.ToLower(keyword)) {
			return "视频生成失败：检测到敏感内容。请修改提示词或参考图片，避免包含不适当的内容后重试。"
		}
	}

	// 检测配额限制错误
	if strings.Contains(lowerErrorMsg, "setlimitexceeded") || strings.Contains(lowerErrorMsg, "inference limit") {
		return "视频生成失败：已达到模型推理配额限制。请访问火山引擎控制台的「模型激活」页面调整或关闭「安全体验模式」，或等待配额重置后重试。"
	}

	// 检测其他常见错误类型
	if strings.Contains(lowerErrorMsg, "rate limit") || strings.Contains(lowerErrorMsg, "too many requests") {
		return "视频生成失败：请求过于频繁，请稍后再试。"
	}

	if strings.Contains(lowerErrorMsg, "timeout") || strings.Contains(lowerErrorMsg, "timed out") {
		return "视频生成失败：请求超时，请检查网络连接后重试。"
	}

	if strings.Contains(lowerErrorMsg, "invalid") || strings.Contains(lowerErrorMsg, "invalid parameter") {
		return "视频生成失败：参数错误，请检查输入信息后重试。"
	}

	// 默认返回原始错误信息
	return errorMsg
}

func (s *VideoGenerationService) getVideoClient(provider string, modelName string) (video.VideoClient, *resolvedClientInfo, error) {
	var config *models.AIServiceConfig
	var err error

	if modelName != "" {
		config, err = s.aiService.GetConfigForModel("video", modelName)
		if err != nil {
			s.log.Warnw("Failed to get config for model, using default", "model", modelName, "error", err)
			config, err = s.aiService.GetDefaultConfig("video")
			if err != nil {
				return nil, nil, fmt.Errorf("no video AI config found: %w", err)
			}
		}
	} else {
		config, err = s.aiService.GetDefaultConfig("video")
		if err != nil {
			return nil, nil, fmt.Errorf("no video AI config found: %w", err)
		}
	}

	baseURL := config.BaseURL
	apiKey := config.APIKey
	model := modelName
	if model == "" && len(config.Model) > 0 {
		model = config.Model[0]
	}

	info := &resolvedClientInfo{Provider: config.Provider, Model: model}
	endpoint := config.Endpoint
	queryEndpoint := config.QueryEndpoint

	switch config.Provider {
	case "doubao", "volcengine", "volces":
		if endpoint == "" {
			endpoint = "/api/v3/contents/generations/tasks"
		}
		if queryEndpoint == "" {
			queryEndpoint = "/api/v3/contents/generations/tasks/{taskId}"
		}
		return video.NewVolcesArkClient(baseURL, apiKey, model, endpoint, queryEndpoint), info, nil
	case "openai":
		return video.NewOpenAISoraClient(baseURL, apiKey, model), info, nil
	case "runway":
		return video.NewRunwayClient(baseURL, apiKey, model), info, nil
	case "pika":
		return video.NewPikaClient(baseURL, apiKey, model), info, nil
	case "minimax":
		return video.NewMinimaxClient(baseURL, apiKey, model), info, nil
	default:
		return nil, info, fmt.Errorf("unsupported video provider: %s", provider)
	}
}

func (s *VideoGenerationService) RecoverPendingTasks() {
	var pendingVideos []models.VideoGeneration
	// Query for pending tasks with non-empty task_id
	// Note: Using IS NOT NULL and != '' to ensure we only get valid task IDs
	if err := s.db.Where("status = ? AND task_id IS NOT NULL AND task_id != ''", models.VideoStatusProcessing).Find(&pendingVideos).Error; err != nil {
		s.log.Errorw("Failed to load pending video tasks", "error", err)
		return
	}

	s.log.Infow("Recovering pending video generation tasks", "count", len(pendingVideos))

	for _, videoGen := range pendingVideos {
		// CRITICAL FIX: Check for nil TaskID before dereferencing to prevent panic
		// Even though we filter for non-empty task_id, GORM might still return nil pointers
		// This nil check prevents a potential runtime panic
		if videoGen.TaskID == nil || *videoGen.TaskID == "" {
			s.log.Warnw("Skipping video generation with nil or empty TaskID", "id", videoGen.ID)
			continue
		}

		// Start goroutine to poll task status for each pending video
		// Each goroutine will poll independently until completion or timeout
		go s.pollTaskStatus(videoGen.ID, *videoGen.TaskID, videoGen.Provider, videoGen.Model)
	}
}

func (s *VideoGenerationService) GetVideoGeneration(id uint) (*models.VideoGeneration, error) {
	var videoGen models.VideoGeneration
	if err := s.db.First(&videoGen, id).Error; err != nil {
		return nil, err
	}
	return &videoGen, nil
}

func (s *VideoGenerationService) ListVideoGenerations(dramaID *uint, storyboardID *uint, status string, limit int, offset int) ([]*models.VideoGeneration, int64, error) {
	var videos []*models.VideoGeneration
	var total int64

	query := s.db.Model(&models.VideoGeneration{})

	if dramaID != nil {
		query = query.Where("drama_id = ?", *dramaID)
	}
	if storyboardID != nil {
		query = query.Where("storyboard_id = ?", *storyboardID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

func (s *VideoGenerationService) GenerateVideoFromImage(imageGenID uint) (*models.VideoGeneration, error) {
	var imageGen models.ImageGeneration
	if err := s.db.First(&imageGen, imageGenID).Error; err != nil {
		return nil, fmt.Errorf("image generation not found")
	}

	if imageGen.Status != models.ImageStatusCompleted || imageGen.ImageURL == nil {
		return nil, fmt.Errorf("image is not ready")
	}

	// 获取关联的Storyboard以获取时长
	var duration *int
	if imageGen.StoryboardID != nil {
		var storyboard models.Storyboard
		if err := s.db.Where("id = ?", *imageGen.StoryboardID).First(&storyboard).Error; err == nil {
			duration = &storyboard.Duration
			s.log.Infow("Using storyboard duration for video generation",
				"storyboard_id", *imageGen.StoryboardID,
				"duration", storyboard.Duration)
		}
	}

	req := &GenerateVideoRequest{
		DramaID:      fmt.Sprintf("%d", imageGen.DramaID),
		StoryboardID: imageGen.StoryboardID,
		ImageGenID:   &imageGenID,
		ImageURL:     *imageGen.ImageURL,
		Prompt:       imageGen.Prompt,
		Provider:     "doubao",
		Duration:     duration,
	}

	return s.GenerateVideo(req)
}

func (s *VideoGenerationService) BatchGenerateVideosForEpisode(episodeID string, skipExisting bool) ([]*models.VideoGeneration, error) {
	var episode models.Episode
	if err := s.db.Preload("Storyboards").Where("id = ?", episodeID).First(&episode).Error; err != nil {
		return nil, domain.ErrEpisodeNotFound
	}

	var results []*models.VideoGeneration
	skippedCount := 0
	for _, storyboard := range episode.Storyboards {
		if storyboard.ImagePrompt == nil {
			continue
		}

		// 如果选择跳过已生成的，检查是否已有成功的视频
		if skipExisting {
			var existingCount int64
			s.db.Model(&models.VideoGeneration{}).
				Where("storyboard_id = ? AND status = ?", storyboard.ID, "completed").
				Count(&existingCount)
			if existingCount > 0 {
				s.log.Infow("Skipping storyboard with existing completed video",
					"storyboard_id", storyboard.ID, "existing_count", existingCount)
				skippedCount++
				continue
			}
		}

		var imageGen models.ImageGeneration
		if err := s.db.Where("storyboard_id = ? AND status = ?", storyboard.ID, models.ImageStatusCompleted).
			Order("created_at DESC").First(&imageGen).Error; err != nil {
			s.log.Warnw("No completed image for storyboard", "storyboard_id", storyboard.ID)
			continue
		}

		videoGen, err := s.GenerateVideoFromImage(imageGen.ID)
		if err != nil {
			s.log.Errorw("Failed to generate video", "storyboard_id", storyboard.ID, "error", err)
			continue
		}

		results = append(results, videoGen)
	}

	if skippedCount > 0 {
		s.log.Infow("Batch video generation skipped existing",
			"episode_id", episodeID, "skipped", skippedCount, "submitted", len(results))
	}

	return results, nil
}

func (s *VideoGenerationService) DeleteVideoGeneration(id uint) error {
	// 先删除关联的 asset 记录
	if err := s.db.Where("video_gen_id = ?", id).Delete(&models.Asset{}).Error; err != nil {
		s.log.Errorw("Failed to delete related assets", "error", err, "video_gen_id", id)
		return err
	}

	// 再删除视频生成记录
	if err := s.db.Delete(&models.VideoGeneration{}, id).Error; err != nil {
		s.log.Errorw("Failed to delete video generation", "error", err, "id", id)
		return err
	}

	s.log.Infow("Video generation deleted successfully", "id", id)
	return nil
}

// resolveImageURL 解析图片 URL：远程 URL 直接返回，本地路径转 base64
func (s *VideoGenerationService) resolveImageURL(imageRef string) string {
	if strings.HasPrefix(imageRef, "http://") || strings.HasPrefix(imageRef, "https://") {
		return imageRef
	}
	// 本地路径 fallback：转 base64
	base64Str, err := s.convertImageToBase64(imageRef)
	if err != nil {
		s.log.Warnw("Failed to convert local image to base64, using path as-is", "error", err, "path", imageRef)
		return imageRef
	}
	return base64Str
}

// urlType 返回 URL 类型用于日志
func (s *VideoGenerationService) urlType(imageRef string) string {
	if strings.HasPrefix(imageRef, "https://") {
		return "cos_url"
	}
	if strings.HasPrefix(imageRef, "http://") {
		return "http_url"
	}
	return "local_path"
}

// convertImageToBase64 将图片转换为base64格式
// 优先使用本地存储的图片，如果没有则使用URL
func (s *VideoGenerationService) convertImageToBase64(imageURL string) (string, error) {
	// 如果已经是base64格式，直接返回
	if strings.HasPrefix(imageURL, "data:") {
		return imageURL, nil
	}

	// 尝试从本地存储读取
	if s.localStorage != nil {
		var relativePath string

		// 1. 检查是否是本地URL（包含 /static/）
		if strings.Contains(imageURL, "/static/") {
			// 提取相对路径，例如从 "http://localhost:5678/static/images/xxx.jpg" 提取 "images/xxx.jpg"
			parts := strings.Split(imageURL, "/static/")
			if len(parts) == 2 {
				relativePath = parts[1]
			}
		} else if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
			// 2. 如果不是 HTTP/HTTPS URL，视为相对路径（如 "images/xxx.jpg"）
			relativePath = imageURL
		}

		// 如果识别出相对路径，尝试读取本地文件
		if relativePath != "" {
			absPath := s.localStorage.GetAbsolutePath(relativePath)

			// 使用工具函数转换为base64
			base64Str, err := utils.ImageToBase64(absPath)
			if err == nil {
				s.log.Infow("Converted local image to base64", "path", relativePath)
				return base64Str, nil
			}
			s.log.Warnw("Failed to convert local image to base64, will try URL", "error", err, "path", absPath)
		}
	}

	// 如果本地读取失败或不是本地路径，尝试从URL下载并转换
	base64Str, err := utils.ImageToBase64(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to convert image to base64: %w", err)
	}

	urlLen := len(imageURL)
	if urlLen > 50 {
		urlLen = 50
	}
	s.log.Infow("Converted remote image to base64", "url", imageURL[:urlLen])
	return base64Str, nil
}

// ============================================================
// V3: 链式视频生成 —— 上一个视频的尾帧自动作为下一个视频的首帧
// ============================================================

// ChainGenerateRequest 链式视频生成请求（已禁用，路由和 handler 已注释）
type ChainGenerateRequest struct {
	EpisodeID      string `json:"episode_id" binding:"required"`
	FirstFrameID   *uint  `json:"first_frame_id"`   // 第1个镜头的首帧图片 ID（可选，也可用 first_frame_path）
	FirstFramePath string `json:"first_frame_path"` // 第1个镜头的首帧图片本地路径
	Model          string `json:"model"`            // 视频模型
	Provider       string `json:"provider"`
	Resolution     string `json:"resolution"` // 视频清晰度：480p 或 720p
	SkipExisting   bool   `json:"skip_existing"`
	GenerateAudio  *bool  `json:"generate_audio"`  // 是否生成音频（默认：true）
	EnableSubtitle *bool  `json:"enable_subtitle"` // 是否生成字幕（默认：false，通过提示词控制）
}

// ChainGenerateResult 链式视频生成结果（已禁用）
type ChainGenerateResult struct {
	Total     int    `json:"total"`
	Completed int    `json:"completed"`
	Failed    int    `json:"failed"`
	Skipped   int    `json:"skipped"`
	Status    string `json:"status"` // running, completed, failed
	Message   string `json:"message"`
}

// ChainGenerateVideosForEpisode V3 链式视频生成（已禁用，路由和 handler 已注释）
// 保留代码以便后续恢复
func (s *VideoGenerationService) ChainGenerateVideosForEpisode(req *ChainGenerateRequest) (*ChainGenerateResult, error) {
	result := &ChainGenerateResult{Status: "running"}

	// 1. 加载章节和分镜
	var ep models.Episode
	if err := s.db.Preload("Drama").Where("id = ?", req.EpisodeID).First(&ep).Error; err != nil {
		return nil, domain.ErrEpisodeNotFound
	}

	var storyboards []models.Storyboard
	if err := s.db.Where("episode_id = ?", req.EpisodeID).
		Order("storyboard_number ASC, id ASC").Find(&storyboards).Error; err != nil {
		return nil, fmt.Errorf("failed to get storyboards: %w", err)
	}

	if len(storyboards) == 0 {
		return nil, fmt.Errorf("no storyboards found for episode")
	}

	result.Total = len(storyboards)
	s.log.Infow("V3 Chain video generation starting",
		"episode_id", req.EpisodeID,
		"total_storyboards", len(storyboards),
		"model", req.Model)

	// 2. 确定第1个镜头的首帧图片
	var currentFramePath string

	if req.FirstFramePath != "" {
		currentFramePath = req.FirstFramePath
	} else if req.FirstFrameID != nil {
		// 从图片生成记录获取
		var img models.ImageGeneration
		if err := s.db.First(&img, *req.FirstFrameID).Error; err != nil {
			return nil, fmt.Errorf("first frame image not found: %w", err)
		}
		if img.LocalPath != nil && *img.LocalPath != "" {
			currentFramePath = *img.LocalPath
		} else if img.ImageURL != nil {
			currentFramePath = *img.ImageURL
		}
	} else {
		// 自动查找第1个分镜的已完成首帧图片
		var img models.ImageGeneration
		err := s.db.Where("storyboard_id = ? AND frame_type = ? AND status = ?",
			storyboards[0].ID, "first", models.ImageStatusCompleted).
			Order("created_at DESC").First(&img).Error
		if err != nil {
			return nil, fmt.Errorf("no first frame image found for storyboard #1, please generate one first")
		}
		if img.LocalPath != nil && *img.LocalPath != "" {
			currentFramePath = *img.LocalPath
		} else if img.ImageURL != nil {
			currentFramePath = *img.ImageURL
		}
	}

	if currentFramePath == "" {
		return nil, fmt.Errorf("cannot determine first frame image path")
	}

	s.log.Infow("First frame for chain generation", "path", currentFramePath)

	// 3. 确定模型和 provider
	provider := req.Provider
	if provider == "" {
		provider = "doubao"
	}
	model := req.Model
	if model == "" {
		model = "doubao-seedance-1-5-pro-251215"
	}

	// 如果请求未指定 generate_audio，从 Drama 配置读取
	if req.GenerateAudio == nil && ep.Drama.GenerateAudio != nil {
		req.GenerateAudio = ep.Drama.GenerateAudio
	}

	// 4. 串行生成每个镜头的视频
	for i, sb := range storyboards {
		// 跳过已有视频的分镜
		if req.SkipExisting {
			var existingCount int64
			s.db.Model(&models.VideoGeneration{}).
				Where("storyboard_id = ? AND status = ?", sb.ID, models.VideoStatusCompleted).
				Count(&existingCount)
			if existingCount > 0 {
				s.log.Infow("V3 chain: skipping storyboard with existing video",
					"index", i+1, "storyboard_id", sb.ID)
				result.Skipped++

				// 但仍然需要获取这个视频的尾帧作为下一个的输入
				var existingVideo models.VideoGeneration
				s.db.Where("storyboard_id = ? AND status = ?", sb.ID, models.VideoStatusCompleted).
					Order("created_at DESC").First(&existingVideo)
				if existingVideo.LocalPath != nil && *existingVideo.LocalPath != "" {
					lastFramePath := s.extractLastFrameFromVideo(*existingVideo.LocalPath, sb.ID)
					if lastFramePath != "" {
						currentFramePath = lastFramePath
					}
				}
				continue
			}
		}

		// 构建视频提示词：使用分镜的 video_prompt 或 action 字段
		prompt := s.buildVideoPromptFromStoryboard(&sb)
		if prompt == "" {
			s.log.Warnw("V3 chain: no prompt for storyboard, skipping",
				"index", i+1, "storyboard_id", sb.ID)
			result.Failed++
			continue
		}

		s.log.Infow("V3 chain: generating video",
			"index", i+1,
			"storyboard_id", sb.ID,
			"input_frame", currentFramePath,
			"prompt_length", len(prompt))

		// 构建单图模式的视频生成请求
		storyboardID := sb.ID
		videoReq := &GenerateVideoRequest{
			DramaID:        fmt.Sprintf("%d", ep.DramaID),
			StoryboardID:   &storyboardID,
			Prompt:         prompt,
			Provider:       provider,
			Model:          model,
			ReferenceMode:  "single",
			ImageLocalPath: &currentFramePath,
			Duration:       &sb.Duration,
		}

		// 添加清晰度参数
		if req.Resolution != "" {
			videoReq.Resolution = &req.Resolution
		}

		// 添加音频和字幕参数
		if req.GenerateAudio != nil {
			videoReq.GenerateAudio = req.GenerateAudio
		}
		if req.EnableSubtitle != nil {
			videoReq.EnableSubtitle = req.EnableSubtitle
		}

		// 提交视频生成
		videoGen, err := s.GenerateVideo(videoReq)
		if err != nil {
			s.log.Errorw("V3 chain: failed to submit video generation",
				"index", i+1, "storyboard_id", sb.ID, "error", err)
			result.Failed++
			continue
		}

		// 同步等待视频生成完成（链式依赖，必须等前一个完成）
		s.log.Infow("V3 chain: waiting for video to complete...",
			"index", i+1, "storyboard_id", sb.ID, "video_gen_id", videoGen.ID)

		completedPath := s.waitForVideoCompletion(videoGen.ID, 5*time.Minute)
		if completedPath == "" {
			s.log.Errorw("V3 chain: video generation timed out or failed",
				"index", i+1, "storyboard_id", sb.ID, "video_gen_id", videoGen.ID)
			result.Failed++
			// 链断裂但继续尝试（用最后已知的帧）
			continue
		}

		result.Completed++
		s.log.Infow("V3 chain: video completed",
			"index", i+1, "storyboard_id", sb.ID,
			"video_path", completedPath)

		// 截取视频最后一帧，作为下一个镜头的输入
		lastFramePath := s.extractLastFrameFromVideo(completedPath, sb.ID)
		if lastFramePath != "" {
			currentFramePath = lastFramePath
			s.log.Infow("V3 chain: extracted last frame for next storyboard",
				"index", i+1, "last_frame_path", lastFramePath)
		} else {
			s.log.Warnw("V3 chain: failed to extract last frame, using previous frame",
				"index", i+1, "storyboard_id", sb.ID)
			// 不更新 currentFramePath，继续使用上一帧
		}
	}

	result.Status = "completed"
	result.Message = fmt.Sprintf("链式生成完成：共%d个镜头，成功%d，失败%d，跳过%d",
		result.Total, result.Completed, result.Failed, result.Skipped)

	s.log.Infow("V3 Chain video generation completed", "result", result)
	return result, nil
}

// buildVideoPromptFromStoryboard 从分镜数据构建视频提示词
func (s *VideoGenerationService) buildVideoPromptFromStoryboard(sb *models.Storyboard) string {
	// 优先使用 video_prompt 字段（前端生成的中文格式）
	if sb.VideoPrompt != nil && *sb.VideoPrompt != "" {
		return *sb.VideoPrompt
	}

	// 如果 video_prompt 为空，使用中文格式构建提示词（与前端逻辑保持一致）
	var parts []string

	// ========== 第一部分：参考图片说明 ==========
	// 注意：后端无法获取参考图片信息，所以这部分留空
	// 用户需要在前端点击"AI提取提示词"按钮来生成完整的提示词

	// ========== 第二部分：视频动作描述 ==========
	parts = append(parts, `【视频动作描述】`)

	// 1. 场景和地点
	if sb.Location != nil && *sb.Location != "" {
		parts = append(parts, fmt.Sprintf("场景：%s", *sb.Location))
	}

	// 2. 动作描述（核心内容）
	if sb.Action != nil && *sb.Action != "" {
		parts = append(parts, fmt.Sprintf("动作：%s", *sb.Action))
	}

	// 3. 运镜
	if sb.Movement != nil && *sb.Movement != "" {
		parts = append(parts, fmt.Sprintf("运镜：%s", *sb.Movement))
	}

	// 4. 景别
	if sb.ShotType != nil && *sb.ShotType != "" {
		parts = append(parts, fmt.Sprintf("景别：%s", *sb.ShotType))
	}

	// 5. 氛围
	if sb.Atmosphere != nil && *sb.Atmosphere != "" {
		parts = append(parts, fmt.Sprintf("氛围：%s", *sb.Atmosphere))
	}

	// 6. 对话
	if sb.Dialogue != nil && *sb.Dialogue != "" {
		parts = append(parts, fmt.Sprintf("对话：%s", *sb.Dialogue))
	}

	// 7. 配乐
	if sb.BgmPrompt != nil && *sb.BgmPrompt != "" {
		parts = append(parts, fmt.Sprintf("配乐：%s", *sb.BgmPrompt))
	}

	// 8. 音效
	if sb.SoundEffect != nil && *sb.SoundEffect != "" {
		parts = append(parts, fmt.Sprintf("音效：%s", *sb.SoundEffect))
	}

	// 9. 镜头结果
	if sb.Result != nil && *sb.Result != "" {
		parts = append(parts, fmt.Sprintf("画面效果：%s", *sb.Result))
	}

	if len(parts) == 0 {
		if sb.Description != nil && *sb.Description != "" {
			return *sb.Description
		}
		return ""
	}

	return strings.Join(parts, "\n")
}

// waitForVideoCompletion 同步等待视频生成完成，返回本地路径
func (s *VideoGenerationService) waitForVideoCompletion(videoGenID uint, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	pollInterval := 5 * time.Second

	for time.Now().Before(deadline) {
		time.Sleep(pollInterval)

		var vg models.VideoGeneration
		if err := s.db.First(&vg, videoGenID).Error; err != nil {
			s.log.Warnw("Failed to poll video status", "error", err, "id", videoGenID)
			return ""
		}

		switch vg.Status {
		case models.VideoStatusCompleted:
			if vg.LocalPath != nil && *vg.LocalPath != "" {
				return *vg.LocalPath
			}
			if vg.VideoURL != nil && *vg.VideoURL != "" {
				return *vg.VideoURL
			}
			return ""
		case models.VideoStatusFailed:
			s.log.Warnw("Video generation failed during chain", "id", videoGenID,
				"error", vg.ErrorMsg)
			return ""
		}
	}

	s.log.Warnw("Timeout waiting for video completion", "id", videoGenID, "timeout", timeout)
	return ""
}

// autoExtractLastFrame 在视频生成完成后自动截取尾帧
func (s *VideoGenerationService) autoExtractLastFrame(videoGenID uint) {
	var videoGen models.VideoGeneration
	if err := s.db.First(&videoGen, videoGenID).Error; err != nil {
		s.log.Warnw("Failed to load video generation for last frame extraction", "error", err, "id", videoGenID)
		return
	}

	// 检查视频是否有本地文件
	if videoGen.LocalPath == nil || *videoGen.LocalPath == "" {
		s.log.Infow("Video has no local path, skipping last frame extraction", "video_gen_id", videoGenID)
		return
	}

	// 截取尾帧
	framePath := s.extractLastFrameFromVideo(*videoGen.LocalPath, *videoGen.StoryboardID)
	if framePath == "" {
		s.log.Warnw("Failed to extract last frame from video", "video_gen_id", videoGenID)
		return
	}

	// 更新数据库，保存尾帧路径
	if err := s.db.Model(&models.VideoGeneration{}).Where("id = ?", videoGenID).Update("last_frame_url", framePath).Error; err != nil {
		s.log.Errorw("Failed to update last frame URL", "error", err, "video_gen_id", videoGenID)
	} else {
		s.log.Infow("Last frame extracted and saved", "video_gen_id", videoGenID, "frame_path", framePath)
	}
}

// GetLastFrameFromStoryboardVideo 从指定分镜的已完成视频截取尾帧（供前端API调用）
// videoID > 0 时指定具体视频，否则取最新的
func (s *VideoGenerationService) GetLastFrameFromStoryboardVideo(storyboardID uint, videoID uint) (string, *models.VideoGeneration, error) {
	var video models.VideoGeneration
	var err error
	if videoID > 0 {
		err = s.db.Where("id = ? AND storyboard_id = ? AND status = ?", videoID, storyboardID, "completed").
			First(&video).Error
	} else {
		err = s.db.Where("storyboard_id = ? AND status = ?", storyboardID, "completed").
			Order("created_at DESC").First(&video).Error
	}
	if err != nil {
		return "", nil, fmt.Errorf("未找到镜头 %d 的已完成视频", storyboardID)
	}

	// 优先使用已保存的尾帧路径
	if video.LastFrameURL != nil && *video.LastFrameURL != "" {
		s.log.Infow("Using cached last frame", "video_id", video.ID, "frame_path", *video.LastFrameURL)
		return *video.LastFrameURL, &video, nil
	}

	// 如果没有本地文件，无法截取
	if video.LocalPath == nil || *video.LocalPath == "" {
		return "", &video, fmt.Errorf("镜头 %d 的视频没有本地文件", storyboardID)
	}

	// 截取尾帧
	framePath := s.extractLastFrameFromVideo(*video.LocalPath, storyboardID)
	if framePath == "" {
		return "", &video, fmt.Errorf("从视频截取尾帧失败")
	}

	// 保存到数据库，下次直接使用
	if err := s.db.Model(&models.VideoGeneration{}).Where("id = ?", video.ID).Update("last_frame_url", framePath).Error; err != nil {
		s.log.Warnw("Failed to save last frame URL to database", "error", err)
	}

	return framePath, &video, nil
}

// extractLastFrameFromVideo 从视频中截取最后一帧，自动上传到COS
func (s *VideoGenerationService) extractLastFrameFromVideo(videoPath string, storyboardID uint) string {
	if s.localStorage == nil || s.ffmpeg == nil {
		s.log.Warnw("Cannot extract last frame: missing localStorage or ffmpeg")
		return ""
	}

	absVideoPath := s.localStorage.GetAbsolutePath(videoPath)

	outputRelPath := fmt.Sprintf("frames/lastframe_sb%d_%d.jpg", storyboardID, time.Now().Unix())
	absOutputPath := s.localStorage.GetAbsolutePath(outputRelPath)

	err := s.ffmpeg.ExtractLastFrame(absVideoPath, absOutputPath)
	if err != nil {
		s.log.Errorw("Failed to extract last frame from video",
			"error", err,
			"video_path", absVideoPath,
			"output_path", absOutputPath)
		return ""
	}

	// 将提取的帧上传到COS，获得永久URL
	f, err := os.Open(absOutputPath)
	if err != nil {
		s.log.Warnw("Failed to open extracted frame for upload", "error", err, "path", absOutputPath)
		return outputRelPath
	}
	defer f.Close()

	cosURL, err := s.localStorage.Upload(f, filepath.Base(absOutputPath), "frames")
	if err != nil {
		s.log.Warnw("Failed to upload extracted frame to storage", "error", err, "path", absOutputPath)
		return outputRelPath
	}

	s.log.Infow("Last frame extracted and uploaded",
		"video_path", videoPath,
		"storage_url", cosURL)

	return cosURL
}

// modelSupportsAudio 判断模型是否支持音频生成
// 当前只使用 Seedance 1.5 Pro，原生支持音频
func (s *VideoGenerationService) modelSupportsAudio(_ string) bool {
	return true
}

// BackfillVideosToCOS 将本地视频文件补传到COS，更新video_url为COS永久链接
func (s *VideoGenerationService) BackfillVideosToCOS() (int, int, error) {
	var videos []models.VideoGeneration
	if err := s.db.Where("local_path IS NOT NULL AND local_path != '' AND status = ?", models.VideoStatusCompleted).
		Find(&videos).Error; err != nil {
		return 0, 0, fmt.Errorf("failed to query videos: %w", err)
	}

	uploaded := 0
	skipped := 0

	for _, v := range videos {
		if v.LocalPath == nil || *v.LocalPath == "" {
			continue
		}

		// 检查video_url是否已是COS链接
		if v.VideoURL != nil && strings.Contains(*v.VideoURL, "cos.") {
			skipped++
			continue
		}

		absPath := s.localStorage.GetAbsolutePath(*v.LocalPath)
		f, err := os.Open(absPath)
		if err != nil {
			s.log.Warnw("Backfill: cannot open local video", "error", err, "path", absPath, "id", v.ID)
			continue
		}

		cosURL, err := s.localStorage.Upload(f, filepath.Base(absPath), "videos")
		f.Close()
		if err != nil {
			s.log.Warnw("Backfill: COS upload failed", "error", err, "id", v.ID)
			continue
		}

		// 更新 video_generations 表
		updates := map[string]interface{}{"video_url": cosURL}
		if err := s.db.Model(&models.VideoGeneration{}).Where("id = ?", v.ID).Updates(updates).Error; err != nil {
			s.log.Errorw("Backfill: DB update failed", "error", err, "id", v.ID)
			continue
		}

		// 同步更新 assets 表中对应的 url
		s.db.Model(&models.Asset{}).Where("video_gen_id = ?", v.ID).Update("url", cosURL)

		s.log.Infow("Backfill: video uploaded to COS", "id", v.ID, "cos_url", cosURL)
		uploaded++
	}

	// 同样处理 last_frame_url（本地相对路径 → COS）
	var videosWithLocalFrame []models.VideoGeneration
	s.db.Where("last_frame_url IS NOT NULL AND last_frame_url != '' AND last_frame_url NOT LIKE 'http%'").
		Find(&videosWithLocalFrame)

	for _, v := range videosWithLocalFrame {
		if v.LastFrameURL == nil {
			continue
		}
		absPath := s.localStorage.GetAbsolutePath(*v.LastFrameURL)
		f, err := os.Open(absPath)
		if err != nil {
			s.log.Warnw("Backfill: cannot open local last frame", "error", err, "path", absPath, "id", v.ID)
			continue
		}
		cosURL, err := s.localStorage.Upload(f, filepath.Base(absPath), "frames")
		f.Close()
		if err != nil {
			s.log.Warnw("Backfill: last frame COS upload failed", "error", err, "id", v.ID)
			continue
		}
		s.db.Model(&models.VideoGeneration{}).Where("id = ?", v.ID).Update("last_frame_url", cosURL)
		s.log.Infow("Backfill: last frame uploaded to COS", "id", v.ID, "cos_url", cosURL)
	}

	return uploaded, skipped, nil
}
