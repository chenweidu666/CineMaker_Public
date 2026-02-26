package models

const (
	DramaStatusDraft      = "draft"
	DramaStatusPlanning   = "planning"
	DramaStatusProduction = "production"
	DramaStatusCompleted  = "completed"
	DramaStatusArchived   = "archived"
)

const (
	EpisodeStatusDraft     = "draft"
	EpisodeStatusCreated   = "created"
	EpisodeStatusGenerating = "generating"
	EpisodeStatusCompleted = "completed"
)

const (
	StyleRealistic = "realistic"
	StyleComic     = "comic"
	DefaultStyle   = StyleRealistic
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
	MinPageSize     = 1
)
