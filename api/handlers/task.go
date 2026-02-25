package handlers

import (
	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskHandler struct {
	taskService *services.TaskService
	db          *gorm.DB
	log         *logger.Logger
}

func NewTaskHandler(db *gorm.DB, log *logger.Logger) *TaskHandler {
	return &TaskHandler{
		taskService: services.NewTaskService(db, log),
		db:          db,
		log:         log,
	}
}

// verifyTaskTeam checks whether the task's resource belongs to the caller's team.
// Task resource_id can be an episode_id, drama_id, storyboard_id, or prop_id depending on type.
func (h *TaskHandler) verifyTaskTeam(task *models.AsyncTask, teamID uint) error {
	switch task.Type {
	case "storyboard_generation", "batch_frame_prompt_generation", "background_extraction":
		_, err := auth.VerifyEpisodeTeam(h.db, task.ResourceID, teamID)
		return err
	case "character_generation", "character_extraction":
		_, err := auth.VerifyDramaTeam(h.db, task.ResourceID, teamID)
		return err
	case "frame_prompt_generation":
		return auth.VerifyStoryboardTeam(h.db, task.ResourceID, teamID)
	case "prop_extraction":
		_, err := auth.VerifyEpisodeTeam(h.db, task.ResourceID, teamID)
		return err
	case "prop_image_generation":
		return auth.VerifyPropTeam(h.db, task.ResourceID, teamID)
	case "script_rewrite":
		_, err := auth.VerifyEpisodeTeam(h.db, task.ResourceID, teamID)
		return err
	default:
		return nil
	}
}

// GetTaskStatus 获取任务状态
func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	teamID := auth.GetTeamID(c)
	taskID := c.Param("task_id")

	task, err := h.taskService.GetTask(taskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "任务不存在")
			return
		}
		h.log.Errorw("Failed to get task", "error", err, "task_id", taskID)
		response.InternalError(c, err.Error())
		return
	}

	if err := h.verifyTaskTeam(task, teamID); err != nil {
		response.Forbidden(c, "无权访问该任务")
		return
	}

	response.Success(c, task)
}

// GetResourceTasks 获取资源相关的所有任务
func (h *TaskHandler) GetResourceTasks(c *gin.Context) {
	resourceID := c.Query("resource_id")
	if resourceID == "" {
		response.BadRequest(c, "缺少resource_id参数")
		return
	}

	tasks, err := h.taskService.GetTasksByResource(resourceID)
	if err != nil {
		h.log.Errorw("Failed to get resource tasks", "error", err, "resource_id", resourceID)
		response.InternalError(c, err.Error())
		return
	}

	// Verify team ownership via the first task's resource
	teamID := auth.GetTeamID(c)
	if len(tasks) > 0 {
		if err := h.verifyTaskTeam(tasks[0], teamID); err != nil {
			response.Forbidden(c, "无权访问该任务")
			return
		}
	}

	response.Success(c, tasks)
}
