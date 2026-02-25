package auth

import (
	"fmt"

	"gorm.io/gorm"
)

// VerifyDramaTeam checks that a drama belongs to the given team.
// Returns the drama ID (parsed uint) on success.
func VerifyDramaTeam(db *gorm.DB, dramaID string, teamID uint) (uint, error) {
	var result struct {
		ID     uint  `gorm:"column:id"`
		TeamID *uint `gorm:"column:team_id"`
	}
	if err := db.Table("dramas").Select("id, team_id").
		Where("id = ? AND deleted_at IS NULL", dramaID).First(&result).Error; err != nil {
		return 0, fmt.Errorf("drama not found")
	}
	if result.TeamID == nil || *result.TeamID != teamID {
		return 0, fmt.Errorf("forbidden: drama does not belong to your team")
	}
	return result.ID, nil
}

// VerifyEpisodeTeam checks that an episode's parent drama belongs to the given team.
// Returns the episode's drama_id on success.
func VerifyEpisodeTeam(db *gorm.DB, episodeID string, teamID uint) (uint, error) {
	var result struct {
		DramaID uint `gorm:"column:drama_id"`
	}
	if err := db.Table("episodes").Select("drama_id").
		Where("id = ? AND deleted_at IS NULL", episodeID).First(&result).Error; err != nil {
		return 0, fmt.Errorf("episode not found")
	}
	_, err := VerifyDramaTeam(db, fmt.Sprintf("%d", result.DramaID), teamID)
	if err != nil {
		return 0, err
	}
	return result.DramaID, nil
}

// VerifyCharacterTeam checks that a character's parent drama belongs to the given team.
func VerifyCharacterTeam(db *gorm.DB, characterID string, teamID uint) error {
	var result struct {
		DramaID uint `gorm:"column:drama_id"`
	}
	if err := db.Table("characters").Select("drama_id").
		Where("id = ?", characterID).First(&result).Error; err != nil {
		return fmt.Errorf("character not found")
	}
	_, err := VerifyDramaTeam(db, fmt.Sprintf("%d", result.DramaID), teamID)
	return err
}

// VerifySceneTeam checks that a scene's parent drama belongs to the given team.
func VerifySceneTeam(db *gorm.DB, sceneID string, teamID uint) error {
	var result struct {
		DramaID uint `gorm:"column:drama_id"`
	}
	if err := db.Table("scenes").Select("drama_id").
		Where("id = ?", sceneID).First(&result).Error; err != nil {
		return fmt.Errorf("scene not found")
	}
	_, err := VerifyDramaTeam(db, fmt.Sprintf("%d", result.DramaID), teamID)
	return err
}

// VerifyPropTeam checks that a prop's parent drama belongs to the given team.
func VerifyPropTeam(db *gorm.DB, propID string, teamID uint) error {
	var result struct {
		DramaID uint `gorm:"column:drama_id"`
	}
	if err := db.Table("props").Select("drama_id").
		Where("id = ?", propID).First(&result).Error; err != nil {
		return fmt.Errorf("prop not found")
	}
	_, err := VerifyDramaTeam(db, fmt.Sprintf("%d", result.DramaID), teamID)
	return err
}

// VerifyImageGenTeam checks that an image generation's drama belongs to the given team.
func VerifyImageGenTeam(db *gorm.DB, imageGenID uint, teamID uint) error {
	var result struct {
		DramaID uint `gorm:"column:drama_id"`
	}
	if err := db.Table("image_generations").Select("drama_id").
		Where("id = ?", imageGenID).First(&result).Error; err != nil {
		return fmt.Errorf("image generation not found")
	}
	_, err := VerifyDramaTeam(db, fmt.Sprintf("%d", result.DramaID), teamID)
	return err
}

// VerifyStoryboardTeam checks that a storyboard's episode's drama belongs to the given team.
func VerifyStoryboardTeam(db *gorm.DB, storyboardID string, teamID uint) error {
	var result struct {
		EpisodeID uint `gorm:"column:episode_id"`
	}
	if err := db.Table("storyboards").Select("episode_id").
		Where("id = ?", storyboardID).First(&result).Error; err != nil {
		return fmt.Errorf("storyboard not found")
	}
	_, err := VerifyEpisodeTeam(db, fmt.Sprintf("%d", result.EpisodeID), teamID)
	return err
}

// VerifyVideoGenTeam checks that a video generation's storyboard's drama belongs to the given team.
func VerifyVideoGenTeam(db *gorm.DB, videoGenID uint, teamID uint) error {
	var result struct {
		StoryboardID uint `gorm:"column:storyboard_id"`
	}
	if err := db.Table("video_generations").Select("storyboard_id").
		Where("id = ?", videoGenID).First(&result).Error; err != nil {
		return fmt.Errorf("video generation not found")
	}
	return VerifyStoryboardTeam(db, fmt.Sprintf("%d", result.StoryboardID), teamID)
}

// VerifyVideoMergeTeam checks that a video merge's drama belongs to the given team.
func VerifyVideoMergeTeam(db *gorm.DB, mergeID uint, teamID uint) error {
	var result struct {
		DramaID uint `gorm:"column:drama_id"`
	}
	if err := db.Table("video_merges").Select("drama_id").
		Where("id = ?", mergeID).First(&result).Error; err != nil {
		return fmt.Errorf("video merge not found")
	}
	_, err := VerifyDramaTeam(db, fmt.Sprintf("%d", result.DramaID), teamID)
	return err
}

// VerifyAIMessageLogTeam checks that an AI message log's drama belongs to the given team.
func VerifyAIMessageLogTeam(db *gorm.DB, logID uint, teamID uint) error {
	var result struct {
		DramaID *uint `gorm:"column:drama_id"`
	}
	if err := db.Table("ai_message_logs").Select("drama_id").
		Where("id = ?", logID).First(&result).Error; err != nil {
		return fmt.Errorf("ai message log not found")
	}
	if result.DramaID == nil {
		return nil
	}
	_, err := VerifyDramaTeam(db, fmt.Sprintf("%d", *result.DramaID), teamID)
	return err
}
