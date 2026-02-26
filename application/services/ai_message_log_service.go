package services

import (
	"encoding/json"
	"time"

	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIMessageLogService struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewAIMessageLogService(db *gorm.DB, log *logger.Logger) *AIMessageLogService {
	return &AIMessageLogService{db: db, log: log}
}

type LogEntry struct {
	RequestID    string
	DramaID      *uint
	ServiceType  string // text, image, video
	Purpose      string
	Provider     string
	Model        string
	Endpoint     string
	SystemPrompt string
	UserPrompt   string
	FullRequest  interface{}
}

// LogRequest creates a pending log entry before sending an AI request.
// Returns the log ID so the caller can update it after getting a response.
func (s *AIMessageLogService) LogRequest(entry LogEntry) uint {
	if entry.RequestID == "" {
		entry.RequestID = uuid.New().String()
	}

	var fullReqJSON []byte
	if entry.FullRequest != nil {
		fullReqJSON, _ = json.Marshal(entry.FullRequest)
	}

	log := models.AIMessageLog{
		RequestID:    entry.RequestID,
		DramaID:      entry.DramaID,
		ServiceType:  entry.ServiceType,
		Purpose:      entry.Purpose,
		Provider:     entry.Provider,
		Model:        entry.Model,
		Endpoint:     entry.Endpoint,
		SystemPrompt: entry.SystemPrompt,
		UserPrompt:   entry.UserPrompt,
		FullRequest:  fullReqJSON,
		Status:       "pending",
	}

	if err := s.db.Create(&log).Error; err != nil {
		s.log.Errorw("Failed to create AI message log", "error", err)
		return 0
	}
	return log.ID
}

// UpdateSuccess marks a log entry as successful with the response.
func (s *AIMessageLogService) UpdateSuccess(logID uint, response string, durationMs int64, promptTokens, outputTokens int) {
	if logID == 0 {
		return
	}
	s.db.Model(&models.AIMessageLog{}).Where("id = ?", logID).Updates(map[string]interface{}{
		"status":        "success",
		"response":      response,
		"duration_ms":   durationMs,
		"prompt_tokens": promptTokens,
		"output_tokens": outputTokens,
		"total_tokens":  promptTokens + outputTokens,
	})
}

// UpdateFailed marks a log entry as failed with error message.
func (s *AIMessageLogService) UpdateFailed(logID uint, errMsg string, durationMs int64) {
	if logID == 0 {
		return
	}
	s.db.Model(&models.AIMessageLog{}).Where("id = ?", logID).Updates(map[string]interface{}{
		"status":        "failed",
		"error_message": errMsg,
		"duration_ms":   durationMs,
	})
}

// LogComplete is a convenience method that logs a completed request in one call.
func (s *AIMessageLogService) LogComplete(entry LogEntry, response string, errMsg string, durationMs int64) {
	if entry.RequestID == "" {
		entry.RequestID = uuid.New().String()
	}

	var fullReqJSON []byte
	if entry.FullRequest != nil {
		fullReqJSON, _ = json.Marshal(entry.FullRequest)
	}

	status := "success"
	if errMsg != "" {
		status = "failed"
	}

	log := models.AIMessageLog{
		RequestID:    entry.RequestID,
		DramaID:      entry.DramaID,
		ServiceType:  entry.ServiceType,
		Purpose:      entry.Purpose,
		Provider:     entry.Provider,
		Model:        entry.Model,
		Endpoint:     entry.Endpoint,
		SystemPrompt: entry.SystemPrompt,
		UserPrompt:   entry.UserPrompt,
		FullRequest:  fullReqJSON,
		Response:     response,
		Status:       status,
		ErrorMessage: errMsg,
		DurationMs:   durationMs,
	}

	if err := s.db.Create(&log).Error; err != nil {
		s.log.Errorw("Failed to create AI message log", "error", err)
	}
}

type AIMessageLogQuery struct {
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
	ServiceType string `form:"service_type"`
	Purpose     string `form:"purpose"`
	Status      string `form:"status"`
	DramaID     *uint  `form:"drama_id"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	Keyword     string `form:"keyword"`
	TeamID      uint   `form:"-"`
}

type AIMessageLogListResult struct {
	Items      []models.AIMessageLog `json:"items"`
	Pagination Pagination            `json:"pagination"`
}

type Pagination struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func (s *AIMessageLogService) List(query AIMessageLogQuery) (*AIMessageLogListResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	tx := s.db.Model(&models.AIMessageLog{})

	if query.TeamID > 0 {
		tx = tx.Where("drama_id IN (SELECT id FROM dramas WHERE team_id = ? AND deleted_at IS NULL) OR drama_id IS NULL", query.TeamID)
	}

	if query.ServiceType != "" {
		tx = tx.Where("service_type = ?", query.ServiceType)
	}
	if query.Purpose != "" {
		tx = tx.Where("purpose = ?", query.Purpose)
	}
	if query.Status != "" {
		tx = tx.Where("status = ?", query.Status)
	}
	if query.DramaID != nil {
		tx = tx.Where("drama_id = ?", *query.DramaID)
	}
	if query.StartDate != "" {
		if t, err := time.Parse("2006-01-02", query.StartDate); err == nil {
			tx = tx.Where("created_at >= ?", t)
		}
	}
	if query.EndDate != "" {
		if t, err := time.Parse("2006-01-02", query.EndDate); err == nil {
			tx = tx.Where("created_at < ?", t.AddDate(0, 0, 1))
		}
	}
	if query.Keyword != "" {
		kw := "%" + query.Keyword + "%"
		tx = tx.Where("user_prompt LIKE ? OR response LIKE ? OR purpose LIKE ?", kw, kw, kw)
	}

	var total int64
	tx.Count(&total)

	var items []models.AIMessageLog
	offset := (query.Page - 1) * query.PageSize
	if err := tx.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}

	return &AIMessageLogListResult{
		Items: items,
		Pagination: Pagination{
			Total:    total,
			Page:     query.Page,
			PageSize: query.PageSize,
		},
	}, nil
}

func (s *AIMessageLogService) GetByID(id uint) (*models.AIMessageLog, error) {
	var log models.AIMessageLog
	if err := s.db.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (s *AIMessageLogService) GetStats(teamID uint) (map[string]interface{}, error) {
	teamScope := func(tx *gorm.DB) *gorm.DB {
		if teamID > 0 {
			return tx.Where("drama_id IN (SELECT id FROM dramas WHERE team_id = ? AND deleted_at IS NULL) OR drama_id IS NULL", teamID)
		}
		return tx
	}

	var totalCount int64
	teamScope(s.db.Model(&models.AIMessageLog{})).Count(&totalCount)

	var successCount int64
	teamScope(s.db.Model(&models.AIMessageLog{})).Where("status = 'success'").Count(&successCount)

	var failedCount int64
	teamScope(s.db.Model(&models.AIMessageLog{})).Where("status = 'failed'").Count(&failedCount)

	type TypeCount struct {
		ServiceType string `json:"service_type"`
		Count       int64  `json:"count"`
	}
	var typeCounts []TypeCount
	teamScope(s.db.Model(&models.AIMessageLog{})).Select("service_type, count(*) as count").Group("service_type").Scan(&typeCounts)

	type PurposeCount struct {
		Purpose string `json:"purpose"`
		Count   int64  `json:"count"`
	}
	var purposeCounts []PurposeCount
	teamScope(s.db.Model(&models.AIMessageLog{})).Select("purpose, count(*) as count").Group("purpose").Order("count desc").Limit(10).Scan(&purposeCounts)

	var todayCount int64
	today := time.Now().Format("2006-01-02")
	teamScope(s.db.Model(&models.AIMessageLog{})).Where("created_at >= ?", today).Count(&todayCount)

	return map[string]interface{}{
		"total":          totalCount,
		"success":        successCount,
		"failed":         failedCount,
		"today":          todayCount,
		"by_type":        typeCounts,
		"by_purpose_top": purposeCounts,
	}, nil
}
