package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AIMessageLog struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	RequestID   string         `gorm:"size:36;index" json:"request_id"`
	DramaID     *uint          `gorm:"index" json:"drama_id,omitempty"`
	ServiceType string         `gorm:"size:30;not null;index" json:"service_type"` // text, image, video
	Purpose     string         `gorm:"size:100;not null;index" json:"purpose"`     // translate_prompt, generate_character, generate_scene, generate_storyboard, extract_characters, etc.
	Provider    string         `gorm:"size:50" json:"provider"`                    // openai, volcengine, gemini, doubao
	Model       string         `gorm:"size:100" json:"model"`
	Endpoint    string         `gorm:"size:200" json:"endpoint,omitempty"`

	// 请求内容
	SystemPrompt string         `gorm:"type:text" json:"system_prompt,omitempty"`
	UserPrompt   string         `gorm:"type:text" json:"user_prompt"`
	FullRequest  datatypes.JSON `gorm:"type:text" json:"full_request,omitempty"` // 完整请求体 JSON（图片/视频时含 options 等）

	// 响应内容
	Response     string `gorm:"type:text" json:"response,omitempty"`
	Status       string `gorm:"size:20;not null;index;default:'pending'" json:"status"` // pending, success, failed
	ErrorMessage string `gorm:"type:text" json:"error_message,omitempty"`

	// 耗时与 token
	DurationMs   int64 `gorm:"default:0" json:"duration_ms"`
	PromptTokens int   `gorm:"default:0" json:"prompt_tokens,omitempty"`
	OutputTokens int   `gorm:"default:0" json:"output_tokens,omitempty"`
	TotalTokens  int   `gorm:"default:0" json:"total_tokens,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AIMessageLog) TableName() string {
	return "ai_message_logs"
}
