package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type VideoAnalysisTask struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID    string         `gorm:"size:36;uniqueIndex" json:"task_id"`
	TeamID    uint           `gorm:"index;default:0" json:"team_id"`
	VideoURL  string         `gorm:"type:text" json:"video_url,omitempty"`
	VideoPath string         `gorm:"type:text" json:"video_path"`
	Title     string         `gorm:"size:200" json:"title"`
	Duration  float64        `gorm:"default:0" json:"duration"`

	Status   string `gorm:"size:30;not null;default:'pending';index" json:"status"`
	Progress int    `gorm:"default:0" json:"progress"`
	Stage    string `gorm:"size:50" json:"stage"`

	ShotCount  int            `gorm:"default:0" json:"shot_count"`
	Result     datatypes.JSON `gorm:"type:text" json:"result,omitempty"`
	StageData  datatypes.JSON `gorm:"type:text" json:"stage_data,omitempty"`
	ErrorMsg   string         `gorm:"type:text" json:"error_msg,omitempty"`

	ImportedDramaID *uint `gorm:"index" json:"imported_drama_id,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (VideoAnalysisTask) TableName() string {
	return "video_analysis_tasks"
}

// AnalysisResult is the structured output from video analysis.
type AnalysisResult struct {
	Title      string          `json:"title"`
	Summary    string          `json:"summary"`
	Tags       []string        `json:"tags,omitempty"`
	Characters []AnalysisChar  `json:"characters"`
	Shots      []AnalysisShot  `json:"shots"`
	Dialogues  []AnalysisLine  `json:"dialogues"`
}

type AnalysisChar struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Profession  string `json:"profession"`
	Role        string `json:"role"`
}

type AnalysisShot struct {
	Index            int      `json:"index"`
	StartTime        float64  `json:"start_time"`
	EndTime          float64  `json:"end_time"`
	Title            string   `json:"title,omitempty"`
	Description      string   `json:"description"`
	Location         string   `json:"location"`
	Time             string   `json:"time,omitempty"`
	Characters       []string `json:"characters"`
	Dialogue         string   `json:"dialogue"`
	Mood             string   `json:"mood"`
	ShotType         string   `json:"shot_type,omitempty"`
	Angle            string   `json:"angle,omitempty"`
	Movement         string   `json:"movement,omitempty"`
	FirstFrameDesc   string   `json:"first_frame_desc,omitempty"`
	MiddleActionDesc string   `json:"middle_action_desc,omitempty"`
	LastFrameDesc    string   `json:"last_frame_desc,omitempty"`
	VideoPrompt      string   `json:"video_prompt,omitempty"`
	BgmPrompt        string   `json:"bgm_prompt,omitempty"`
	SoundEffect      string   `json:"sound_effect,omitempty"`
	FramePath        string   `json:"frame_path,omitempty"`
}

type AnalysisLine struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Speaker   string  `json:"speaker"`
	Text      string  `json:"text"`
}
