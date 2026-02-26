package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/cinemaker/backend/pkg/video"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VideoAnalysisHandler struct {
	service *services.VideoAnalysisService
	log     *logger.Logger
	workDir string
}

func NewVideoAnalysisHandler(service *services.VideoAnalysisService, log *logger.Logger, workDir string) *VideoAnalysisHandler {
	return &VideoAnalysisHandler{
		service: service,
		log:     log,
		workDir: workDir,
	}
}

// UploadAndAnalyze handles video file uploads and starts analysis.
func (h *VideoAnalysisHandler) UploadAndAnalyze(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		response.BadRequest(c, "请上传视频文件")
		return
	}

	if file.Size > 500*1024*1024 {
		response.BadRequest(c, "视频文件不能超过 500MB")
		return
	}

	taskUUID := uuid.New().String()
	uploadDir := filepath.Join(h.workDir, "video_analysis", taskUUID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.InternalError(c, "创建目录失败")
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".mp4"
	}
	videoPath := filepath.Join(uploadDir, "input"+ext)
	if err := c.SaveUploadedFile(file, videoPath); err != nil {
		response.InternalError(c, "保存视频失败")
		return
	}

	teamID := auth.GetTeamID(c)
	task, err := h.service.CreateTask(videoPath, "", file.Filename, teamID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("创建任务失败: %v", err))
		return
	}

	h.service.StartProcessing(task.ID)

	response.Success(c, gin.H{
		"task_id": task.TaskID,
		"status":  task.Status,
		"message": "视频上传成功，开始分析",
	})
}

// AnalyzeFromURL handles video analysis from a URL (yt-dlp download).
func (h *VideoAnalysisHandler) AnalyzeFromURL(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供视频链接")
		return
	}

	if !video.IsValidVideoURL(req.URL) {
		response.BadRequest(c, "不支持的视频链接格式")
		return
	}

	teamID := auth.GetTeamID(c)
	task, err := h.service.DownloadAndCreateTask(req.URL, teamID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("创建任务失败: %v", err))
		return
	}

	response.Success(c, gin.H{
		"task_id": task.TaskID,
		"status":  task.Status,
		"message": "开始下载视频并分析",
	})
}

// GetTaskStatus returns the current status of a video analysis task.
func (h *VideoAnalysisHandler) GetTaskStatus(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	taskID := c.Param("taskId")
	if taskID == "" {
		response.BadRequest(c, "缺少 taskId")
		return
	}

	task, err := h.service.GetTask(taskID)
	if err != nil {
		response.NotFound(c, "任务不存在")
		return
	}

	if task.TeamID != 0 && task.TeamID != teamID {
		response.Forbidden(c, "无权访问该任务")
		return
	}

	response.Success(c, task)
}

// ListTasks returns recent video analysis tasks.
func (h *VideoAnalysisHandler) ListTasks(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	tasks, err := h.service.ListTasks(50, teamID)
	if err != nil {
		response.InternalError(c, "查询任务列表失败")
		return
	}
	response.Success(c, gin.H{"items": tasks})
}

// RetryTask retries a failed or completed analysis task.
func (h *VideoAnalysisHandler) RetryTask(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	taskID := c.Param("taskId")
	if taskID == "" {
		response.BadRequest(c, "缺少 taskId")
		return
	}

	existing, err := h.service.GetTask(taskID)
	if err != nil {
		response.NotFound(c, "任务不存在")
		return
	}
	if existing.TeamID != 0 && existing.TeamID != teamID {
		response.Forbidden(c, "无权操作该任务")
		return
	}

	task, err := h.service.RetryTask(taskID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("重试失败: %v", err))
		return
	}

	response.Success(c, gin.H{
		"task_id": task.TaskID,
		"status":  task.Status,
		"message": "已重新开始分析",
	})
}

// DeleteTask deletes a failed or completed analysis task.
func (h *VideoAnalysisHandler) DeleteTask(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	taskID := c.Param("taskId")
	if taskID == "" {
		response.BadRequest(c, "缺少 taskId")
		return
	}

	existing, err := h.service.GetTask(taskID)
	if err != nil {
		response.NotFound(c, "任务不存在")
		return
	}
	if existing.TeamID != 0 && existing.TeamID != teamID {
		response.Forbidden(c, "无权操作该任务")
		return
	}

	if err := h.service.DeleteTask(taskID); err != nil {
		response.InternalError(c, fmt.Sprintf("删除失败: %v", err))
		return
	}

	response.Success(c, gin.H{"message": "已删除"})
}

// ResynthesizeScript re-runs only the LLM synthesis step.
func (h *VideoAnalysisHandler) ResynthesizeScript(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	taskID := c.Param("taskId")
	if taskID == "" {
		response.BadRequest(c, "缺少 taskId")
		return
	}

	existing, err := h.service.GetTask(taskID)
	if err != nil {
		response.NotFound(c, "任务不存在")
		return
	}
	if existing.TeamID != 0 && existing.TeamID != teamID {
		response.Forbidden(c, "无权操作该任务")
		return
	}

	var req struct {
		IncludeAudio *bool `json:"include_audio"`
	}
	c.ShouldBindJSON(&req)
	includeAudio := true
	if req.IncludeAudio != nil {
		includeAudio = *req.IncludeAudio
	}

	go func() {
		if err := h.service.ResynthesizeScript(taskID, includeAudio); err != nil {
			h.log.Errorw("Failed to resynthesize", "error", err, "taskId", taskID)
		}
	}()

	response.Success(c, gin.H{
		"task_id": existing.TaskID,
		"status":  "processing",
		"message": "正在重新合成剧本",
	})
}

// ImportToDrama imports analysis results as a CineMaker drama.
func (h *VideoAnalysisHandler) ImportToDrama(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	taskID := c.Param("taskId")
	if taskID == "" {
		response.BadRequest(c, "缺少 taskId")
		return
	}

	existing, err := h.service.GetTask(taskID)
	if err != nil {
		response.NotFound(c, "任务不存在")
		return
	}
	if existing.TeamID != 0 && existing.TeamID != teamID {
		response.Forbidden(c, "无权操作该任务")
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	c.ShouldBindJSON(&req)

	drama, err := h.service.ImportToDrama(taskID, req.Title, teamID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("导入失败: %v", err))
		return
	}

	response.Success(c, gin.H{
		"message":  "导入成功",
		"drama_id": drama.ID,
		"title":    drama.Title,
	})
}
