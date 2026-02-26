package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DramaExportImportHandler struct {
	db            *gorm.DB
	exportService *services.DramaExportService
	importService *services.DramaImportService
	log           *logger.Logger
	storagePath   string
}

func NewDramaExportImportHandler(
	db *gorm.DB,
	exportService *services.DramaExportService,
	importService *services.DramaImportService,
	log *logger.Logger,
	storagePath string,
) *DramaExportImportHandler {
	return &DramaExportImportHandler{
		db:            db,
		exportService: exportService,
		importService: importService,
		log:           log,
		storagePath:   storagePath,
	}
}

// ExportDrama exports a drama as a ZIP file download.
func (h *DramaExportImportHandler) ExportDrama(c *gin.Context) {
	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	idStr := c.Param("id")
	if _, err := auth.VerifyDramaTeam(h.db, idStr, teamID); err != nil {
		response.Forbidden(c, "无权访问该剧本")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的剧本 ID")
		return
	}

	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("drama_export_%s.zip", uuid.New().String()))
	defer os.Remove(tmpFile)

	if err := h.exportService.ExportDramaToZip(uint(id), tmpFile); err != nil {
		h.log.Errorw("export drama failed", "error", err, "drama_id", id)
		response.InternalError(c, fmt.Sprintf("导出失败: %v", err))
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="drama_%d.zip"`, id))
	c.Header("Content-Type", "application/zip")
	c.File(tmpFile)
}

// ExportAnalysis exports video analysis results directly as a drama ZIP.
// Video analysis tasks do not have team_id yet; no team check.
func (h *DramaExportImportHandler) ExportAnalysis(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		response.BadRequest(c, "缺少 taskId")
		return
	}

	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("analysis_export_%s.zip", uuid.New().String()))
	defer os.Remove(tmpFile)

	if err := h.exportService.ExportAnalysisToZip(taskID, tmpFile); err != nil {
		h.log.Errorw("export analysis failed", "error", err, "task_id", taskID)
		response.InternalError(c, fmt.Sprintf("导出失败: %v", err))
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="analysis_%s.zip"`, taskID))
	c.Header("Content-Type", "application/zip")
	c.File(tmpFile)
}

// ImportDrama imports a drama from an uploaded ZIP file.
func (h *DramaExportImportHandler) ImportDrama(c *gin.Context) {
	teamID, err := auth.MustGetTeamID(c)
	if err != nil {
		response.Unauthorized(c, "请先登录")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传 ZIP 文件")
		return
	}

	if file.Size > 500*1024*1024 {
		response.BadRequest(c, "文件不能超过 500MB")
		return
	}

	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("drama_import_%s.zip", uuid.New().String()))
	if err := c.SaveUploadedFile(file, tmpFile); err != nil {
		response.InternalError(c, "保存上传文件失败")
		return
	}
	defer os.Remove(tmpFile)

	dramaID, err := h.importService.ImportDramaFromZip(tmpFile)
	if err != nil {
		h.log.Errorw("import drama failed", "error", err)
		response.InternalError(c, fmt.Sprintf("导入失败: %v", err))
		return
	}

	if err := h.db.Table("dramas").Where("id = ?", dramaID).Update("team_id", teamID).Error; err != nil {
		h.log.Errorw("set drama team_id after import failed", "error", err, "drama_id", dramaID)
		response.InternalError(c, "导入成功但设置团队失败")
		return
	}

	response.Success(c, gin.H{
		"message":  "导入成功",
		"drama_id": dramaID,
	})
}
