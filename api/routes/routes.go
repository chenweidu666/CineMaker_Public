package routes

import (
	handlers2 "github.com/cinemaker/backend/api/handlers"
	middlewares2 "github.com/cinemaker/backend/api/middlewares"
	services2 "github.com/cinemaker/backend/application/services"
	storage2 "github.com/cinemaker/backend/infrastructure/storage"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(cfg *config.Config, db *gorm.DB, log *logger.Logger, store storage2.Storage) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middlewares2.LoggerMiddleware(log))
	r.Use(middlewares2.CORSMiddleware(cfg.Server.CORSOrigins))

	// 静态文件服务（用户上传的文件）
	r.Static("/static", cfg.Storage.LocalPath)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"app":     cfg.App.Name,
			"version": cfg.App.Version,
		})
	})

	// Auth service & handlers
	authService := services2.NewAuthService(db, cfg, log)
	authHandler := handlers2.NewAuthHandler(authService, log)
	teamHandler := handlers2.NewTeamHandler(authService, log)

	aiService := services2.NewAIService(db, log)
	transferService := services2.NewResourceTransferService(db, log)
	promptI18n := services2.NewPromptI18n(cfg)
	dramaHandler := handlers2.NewDramaHandler(db, cfg, log, nil)
	aiConfigHandler := handlers2.NewAIConfigHandler(db, cfg, log)
	scriptGenHandler := handlers2.NewScriptGenerationHandler(db, cfg, log)
	imageGenService := services2.NewImageGenerationService(db, cfg, transferService, store, log)
	imageGenHandler := handlers2.NewImageGenerationHandler(db, cfg, log, transferService, store)
	videoGenHandler := handlers2.NewVideoGenerationHandler(db, transferService, store, aiService, log, promptI18n)
	videoMergeHandler := handlers2.NewVideoMergeHandler(db, nil, cfg.Storage.LocalPath, cfg.Storage.BaseURL, log)
	assetHandler := handlers2.NewAssetHandler(db, cfg, log)
	characterLibraryService := services2.NewCharacterLibraryService(db, log, cfg, store)
	characterLibraryHandler := handlers2.NewCharacterLibraryHandler(db, cfg, log, transferService, store)
	uploadHandler := handlers2.NewUploadHandler(store, log, characterLibraryService)
	storyboardHandler := handlers2.NewStoryboardHandler(db, cfg, log)
	sceneHandler := handlers2.NewSceneHandler(db, log, imageGenService, aiService, cfg)
	taskHandler := handlers2.NewTaskHandler(db, log)
	framePromptService := services2.NewFramePromptService(db, cfg, log)
	framePromptHandler := handlers2.NewFramePromptHandler(framePromptService, log)
	audioExtractionHandler := handlers2.NewAudioExtractionHandler(log, cfg.Storage.LocalPath)
	propHandler := handlers2.NewPropHandler(db, cfg, log, aiService, imageGenService)
	aiMsgLogService := services2.NewAIMessageLogService(db, log)
	aiMsgLogHandler := handlers2.NewAIMessageLogHandler(db, aiMsgLogService, log)
	videoAnalysisService := services2.NewVideoAnalysisService(db, log, aiService, cfg.Storage.LocalPath)
	videoAnalysisHandler := handlers2.NewVideoAnalysisHandler(videoAnalysisService, log, cfg.Storage.LocalPath)
	dramaExportService := services2.NewDramaExportService(db, log, cfg.Storage.LocalPath)
	dramaImportService := services2.NewDramaImportService(db, log, cfg.Storage.LocalPath)
	dramaExportImportHandler := handlers2.NewDramaExportImportHandler(db, dramaExportService, dramaImportService, log, cfg.Storage.LocalPath)

	api := r.Group("/api/v1")
	{
		api.Use(middlewares2.RateLimitMiddleware())

		// Public auth routes (no auth required)
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/refresh", authHandler.RefreshToken)
		}

		// Admin/management routes (internal only, no auth)
		admin := api.Group("/admin")
		{
			admin.POST("/backfill-videos-cos", videoGenHandler.BackfillVideosToCOS)
		}

		// All routes below require authentication
		api.Use(middlewares2.AuthMiddleware(authService))

		// Authenticated auth routes
		authMe := api.Group("/auth")
		{
			authMe.GET("/me", authHandler.GetMe)
			authMe.PUT("/me", authHandler.UpdateMe)
		}

		// Team routes
		teamRoutes := api.Group("/team")
		{
			teamRoutes.GET("", teamHandler.GetTeam)
			teamRoutes.PUT("", teamHandler.UpdateTeam)
			teamRoutes.POST("/invite", teamHandler.InviteMember)
			teamRoutes.POST("/invite/accept", teamHandler.AcceptInvitation)
			teamRoutes.DELETE("/members/:id", teamHandler.RemoveMember)
		}

		dramas := api.Group("/dramas")
		{
			dramas.GET("", dramaHandler.ListDramas)
			dramas.POST("", dramaHandler.CreateDrama)
			dramas.GET("/stats", dramaHandler.GetDramaStats) // 统计接口放在/:id之前
			dramas.GET("/:id", dramaHandler.GetDrama)
			dramas.PUT("/:id", dramaHandler.UpdateDrama)
			dramas.DELETE("/:id", dramaHandler.DeleteDrama)

			dramas.PUT("/:id/outline", dramaHandler.SaveOutline)
			dramas.GET("/:id/characters", dramaHandler.GetCharacters)
			dramas.PUT("/:id/characters", dramaHandler.SaveCharacters)
			dramas.PUT("/:id/episodes", dramaHandler.SaveEpisodes)
			dramas.PUT("/:id/episodes/title", dramaHandler.UpdateEpisodeTitle)
			dramas.PUT("/:id/progress", dramaHandler.SaveProgress)
			dramas.GET("/:id/props", propHandler.ListProps) // Added prop list route
			dramas.GET("/:id/export", dramaExportImportHandler.ExportDrama)
		}

		dramas.POST("/import", dramaExportImportHandler.ImportDrama)

		// 角色路由
		characters := api.Group("/characters")
		{
			characters.POST("", dramaHandler.CreateCharacter)
			characters.PUT("/:id", characterLibraryHandler.UpdateCharacter)
			characters.DELETE("/:id", characterLibraryHandler.DeleteCharacter)
			characters.POST("/:id/generate-image", characterLibraryHandler.GenerateCharacterImage)
			characters.POST("/batch-generate-images", characterLibraryHandler.BatchGenerateCharacterImages)
			characters.POST("/:id/upload-image", uploadHandler.UploadCharacterImage)
			characters.PUT("/:id/image", characterLibraryHandler.UploadCharacterImage)
			characters.PUT("/:id/image-from-library", characterLibraryHandler.ApplyLibraryItemToCharacter)
			characters.POST("/:id/add-to-library", characterLibraryHandler.AddCharacterToLibrary)
		}

		aiConfigs := api.Group("/ai-configs")
		{
			aiConfigs.GET("", aiConfigHandler.ListConfigs)
			aiConfigs.POST("", aiConfigHandler.CreateConfig)
			aiConfigs.POST("/test", aiConfigHandler.TestConnection)
			aiConfigs.POST("/test-all", aiConfigHandler.TestConnectionAll)
			aiConfigs.GET("/:id", aiConfigHandler.GetConfig)
			aiConfigs.PUT("/:id", aiConfigHandler.UpdateConfig)
			aiConfigs.DELETE("/:id", aiConfigHandler.DeleteConfig)
		}

		generation := api.Group("/generation")
		{
			generation.POST("/characters", scriptGenHandler.GenerateCharacters)
			generation.POST("/rewrite-script", scriptGenHandler.RewriteScript)
			generation.POST("/revert-rewrite", scriptGenHandler.RevertScriptRewrite)
			generation.POST("/parse-extract", scriptGenHandler.ParseExtract) // V3: 程序解析提取角色和场景
		}

		// 角色库路由
		characterLibrary := api.Group("/character-library")
		{
			characterLibrary.GET("", characterLibraryHandler.ListLibraryItems)
			characterLibrary.POST("", characterLibraryHandler.CreateLibraryItem)
			characterLibrary.GET("/:id", characterLibraryHandler.GetLibraryItem)
			characterLibrary.DELETE("/:id", characterLibraryHandler.DeleteLibraryItem)
		}

		props := api.Group("/props")
		{
			props.POST("", propHandler.CreateProp)
			props.PUT("/:id", propHandler.UpdateProp)
			props.DELETE("/:id", propHandler.DeleteProp)
			props.POST("/:id/generate", propHandler.GenerateImage)
		}

		// 文件上传路由
		upload := api.Group("/upload")
		{
			upload.POST("/image", uploadHandler.UploadImage)
		}

		// 分镜头路由
		episodes := api.Group("/episodes")
		{
			episodes.PUT("/:episode_id/script", dramaHandler.SaveEpisodeScript)
			episodes.PUT("/:episode_id/description", dramaHandler.SaveEpisodeDescription)
			// 分镜头
			episodes.POST("/:episode_id/storyboards", storyboardHandler.GenerateStoryboard)
			episodes.POST("/:episode_id/props/extract", propHandler.ExtractProps)
			episodes.POST("/:episode_id/characters/extract", characterLibraryHandler.ExtractCharacters)
			episodes.GET("/:episode_id/storyboards", sceneHandler.GetStoryboardsForEpisode)
			episodes.POST("/:episode_id/finalize", dramaHandler.FinalizeEpisode)
			episodes.GET("/:episode_id/download", dramaHandler.DownloadEpisodeVideo)
			episodes.POST("/:episode_id/batch-frame-prompts", framePromptHandler.BatchGenerateFramePrompts)
			episodes.POST("/:episode_id/characters", dramaHandler.AddCharacterToEpisode)
			episodes.POST("/:episode_id/scenes", dramaHandler.AddSceneToEpisode)
			episodes.POST("/:episode_id/props", dramaHandler.AddPropToEpisode)
			episodes.DELETE("/:episode_id/characters/:character_id", dramaHandler.RemoveCharacterFromEpisode)
			episodes.DELETE("/:episode_id/scenes/:scene_id", dramaHandler.RemoveSceneFromEpisode)
			episodes.DELETE("/:episode_id/props/:prop_id", dramaHandler.RemovePropFromEpisode)
		}

		// 任务路由
		tasks := api.Group("/tasks")
		{
			tasks.GET("/:task_id", taskHandler.GetTaskStatus)
			tasks.GET("", taskHandler.GetResourceTasks)
		}

		// 场景路由
		scenes := api.Group("/scenes")
		{
			scenes.GET("", sceneHandler.ListScenes)
			scenes.PUT("/:scene_id", sceneHandler.UpdateScene)
			scenes.PUT("/:scene_id/prompt", sceneHandler.UpdateScenePrompt)
			scenes.DELETE("/:scene_id", sceneHandler.DeleteScene)
			scenes.POST("/generate-image", sceneHandler.GenerateSceneImage)
			scenes.POST("", sceneHandler.CreateScene)
			scenes.POST("/polish-prompt", sceneHandler.PolishPrompt)
		}

		images := api.Group("/images")
		{
			images.GET("", imageGenHandler.ListImageGenerations)
			images.POST("", imageGenHandler.GenerateImage)
			images.POST("/preview", imageGenHandler.PreviewImagePrompt)
			images.GET("/:id", imageGenHandler.GetImageGeneration)
			images.DELETE("/:id", imageGenHandler.DeleteImageGeneration)
			// DEPRECATED: 注释掉场景图片生成路由 - 目前只使用首帧图片生成
			// images.POST("/scene/:scene_id", imageGenHandler.GenerateImagesForScene)
			images.POST("/edit", imageGenHandler.EditImage)
			images.GET("/:id/edits", imageGenHandler.ListEditsBySource)
			images.POST("/:id/replace", imageGenHandler.ReplaceWithEdit)
			images.POST("/upload", imageGenHandler.UploadImage)
			images.GET("/episode/:episode_id/backgrounds", imageGenHandler.GetBackgroundsForEpisode)
			images.POST("/episode/:episode_id/backgrounds/extract", imageGenHandler.ExtractBackgroundsForEpisode)
			images.POST("/episode/:episode_id/batch", imageGenHandler.BatchGenerateForEpisode)
		}

		videos := api.Group("/videos")
		{
			videos.GET("", videoGenHandler.ListVideoGenerations)
			videos.POST("", videoGenHandler.GenerateVideo)
			videos.GET("/:id", videoGenHandler.GetVideoGeneration)
			videos.DELETE("/:id", videoGenHandler.DeleteVideoGeneration)
			videos.POST("/image/:image_gen_id", videoGenHandler.GenerateVideoFromImage)
			videos.POST("/episode/:episode_id/batch", videoGenHandler.BatchGenerateForEpisode)
			// videos.POST("/episode/:episode_id/chain", videoGenHandler.ChainGenerateForEpisode)
			videos.POST("/storyboard/:storyboard_id/extract-last-frame", videoGenHandler.ExtractLastFrame)
		}

		videoMerges := api.Group("/video-merges")
		{
			videoMerges.GET("", videoMergeHandler.ListMerges)
			videoMerges.POST("", videoMergeHandler.MergeVideos)
			videoMerges.GET("/:merge_id", videoMergeHandler.GetMerge)
			videoMerges.DELETE("/:merge_id", videoMergeHandler.DeleteMerge)
		}

		assets := api.Group("/assets")
		{
			assets.GET("", assetHandler.ListAssets)
			assets.POST("", assetHandler.CreateAsset)
			assets.GET("/:id", assetHandler.GetAsset)
			assets.PUT("/:id", assetHandler.UpdateAsset)
			assets.DELETE("/:id", assetHandler.DeleteAsset)
			assets.POST("/import/image/:image_gen_id", assetHandler.ImportFromImageGen)
			assets.POST("/import/video/:video_gen_id", assetHandler.ImportFromVideoGen)
		}

		storyboards := api.Group("/storyboards")
		{
			storyboards.GET("/episode/:episode_id/generate", storyboardHandler.GenerateStoryboard)
			storyboards.POST("", storyboardHandler.CreateStoryboard)
			storyboards.PUT("/:id", storyboardHandler.UpdateStoryboard)
			storyboards.DELETE("/:id", storyboardHandler.DeleteStoryboard)
			storyboards.POST("/:id/props", propHandler.AssociateProps)
			storyboards.POST("/:id/frame-prompt", framePromptHandler.GenerateFramePrompt)
			storyboards.GET("/:id/frame-prompts", handlers2.GetStoryboardFramePrompts(db, log))
			storyboards.POST("/:id/video-prompt", storyboardHandler.GenerateVideoPrompt)
		}

		audio := api.Group("/audio")
		{
			audio.POST("/extract", audioExtractionHandler.ExtractAudio)
			audio.POST("/extract/batch", audioExtractionHandler.BatchExtractAudio)
		}


		aiLogs := api.Group("/ai-logs")
		{
			aiLogs.GET("", aiMsgLogHandler.List)
			aiLogs.GET("/stats", aiMsgLogHandler.GetStats)
			aiLogs.GET("/:id", aiMsgLogHandler.GetByID)
		}

		videoAnalysis := api.Group("/video-analysis")
		{
			videoAnalysis.GET("", videoAnalysisHandler.ListTasks)
			videoAnalysis.POST("/upload", videoAnalysisHandler.UploadAndAnalyze)
			videoAnalysis.POST("/from-url", videoAnalysisHandler.AnalyzeFromURL)
			videoAnalysis.GET("/:taskId/status", videoAnalysisHandler.GetTaskStatus)
			videoAnalysis.POST("/:taskId/retry", videoAnalysisHandler.RetryTask)
			videoAnalysis.POST("/:taskId/resynthesize", videoAnalysisHandler.ResynthesizeScript)
			videoAnalysis.DELETE("/:taskId", videoAnalysisHandler.DeleteTask)
			videoAnalysis.POST("/:taskId/import", videoAnalysisHandler.ImportToDrama)
			videoAnalysis.GET("/:taskId/export", dramaExportImportHandler.ExportAnalysis)
		}
	}

	// 前端静态文件服务（放在API路由之后，避免冲突）
	// 服务前端构建产物
	r.Static("/assets", "./web/dist/assets")
	r.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

	// NoRoute处理：对于所有未匹配的路由
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是API路径，返回404
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}

		// SPA fallback - 返回index.html
		c.File("./web/dist/index.html")
	})

	return r
}
