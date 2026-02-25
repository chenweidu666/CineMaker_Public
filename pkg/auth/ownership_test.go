package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite DB and creates all tables needed for ownership tests.
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	raw := []string{
		`CREATE TABLE dramas (id INTEGER PRIMARY KEY, team_id INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE episodes (id INTEGER PRIMARY KEY, drama_id INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE characters (id INTEGER PRIMARY KEY, drama_id INTEGER)`,
		`CREATE TABLE scenes (id INTEGER PRIMARY KEY, drama_id INTEGER)`,
		`CREATE TABLE props (id INTEGER PRIMARY KEY, drama_id INTEGER)`,
		`CREATE TABLE image_generations (id INTEGER PRIMARY KEY, drama_id INTEGER)`,
		`CREATE TABLE storyboards (id INTEGER PRIMARY KEY, episode_id INTEGER)`,
		`CREATE TABLE video_generations (id INTEGER PRIMARY KEY, storyboard_id INTEGER)`,
		`CREATE TABLE video_merges (id INTEGER PRIMARY KEY, drama_id INTEGER)`,
		`CREATE TABLE ai_message_logs (id INTEGER PRIMARY KEY, drama_id INTEGER)`,
	}
	for _, s := range raw {
		assert.NoError(t, db.Exec(s).Error)
	}
	return db
}

func TestVerifyDramaTeam(t *testing.T) {
	db := setupTestDB(t)

	// Seed: drama 1 -> team 1, drama 2 -> team 2, drama 3 -> soft-deleted
	assert.NoError(t, db.Exec(`INSERT INTO dramas (id, team_id, deleted_at) VALUES (1, 1, NULL)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO dramas (id, team_id, deleted_at) VALUES (2, 2, NULL)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO dramas (id, team_id, deleted_at) VALUES (3, 1, '2020-01-01 00:00:00')`).Error)

	t.Run("Success_drama_exists_with_correct_team", func(t *testing.T) {
		id, err := VerifyDramaTeam(db, "1", 1)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), id)
	})

	t.Run("Fail_drama_does_not_exist", func(t *testing.T) {
		_, err := VerifyDramaTeam(db, "999", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "drama not found")
	})

	t.Run("Fail_drama_belongs_to_different_team", func(t *testing.T) {
		_, err := VerifyDramaTeam(db, "2", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "forbidden")
	})

	t.Run("Fail_drama_soft_deleted", func(t *testing.T) {
		_, err := VerifyDramaTeam(db, "3", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "drama not found")
	})
}

func TestVerifyEpisodeTeam(t *testing.T) {
	db := setupTestDB(t)

	assert.NoError(t, db.Exec(`INSERT INTO dramas (id, team_id, deleted_at) VALUES (1, 1, NULL)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO episodes (id, drama_id, deleted_at) VALUES (1, 1, NULL)`).Error)

	t.Run("Success_episode_drama_belongs_to_team", func(t *testing.T) {
		dramaID, err := VerifyEpisodeTeam(db, "1", 1)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), dramaID)
	})

	t.Run("Fail_episode_does_not_exist", func(t *testing.T) {
		_, err := VerifyEpisodeTeam(db, "999", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "episode not found")
	})
}

func TestVerifyCharacterTeam(t *testing.T) {
	db := setupTestDB(t)
	assert.NoError(t, db.Exec(`INSERT INTO dramas (id, team_id, deleted_at) VALUES (1, 1, NULL)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO characters (id, drama_id) VALUES (1, 1)`).Error)

	t.Run("Success", func(t *testing.T) {
		err := VerifyCharacterTeam(db, "1", 1)
		assert.NoError(t, err)
	})
	t.Run("Fail_character_not_found", func(t *testing.T) {
		err := VerifyCharacterTeam(db, "999", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "character not found")
	})
}

func TestVerifySceneTeam(t *testing.T) {
	db := setupTestDB(t)
	assert.NoError(t, db.Exec(`INSERT INTO dramas (id, team_id, deleted_at) VALUES (1, 1, NULL)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO scenes (id, drama_id) VALUES (1, 1)`).Error)

	t.Run("Success", func(t *testing.T) {
		err := VerifySceneTeam(db, "1", 1)
		assert.NoError(t, err)
	})
	t.Run("Fail_scene_not_found", func(t *testing.T) {
		err := VerifySceneTeam(db, "999", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scene not found")
	})
}

func TestVerifyPropTeam(t *testing.T) {
	db := setupTestDB(t)
	assert.NoError(t, db.Exec(`INSERT INTO dramas (id, team_id, deleted_at) VALUES (1, 1, NULL)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO props (id, drama_id) VALUES (1, 1)`).Error)

	t.Run("Success", func(t *testing.T) {
		err := VerifyPropTeam(db, "1", 1)
		assert.NoError(t, err)
	})
	t.Run("Fail_prop_not_found", func(t *testing.T) {
		err := VerifyPropTeam(db, "999", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "prop not found")
	})
}

func TestVerifyImageGenTeam(t *testing.T) {
	db := setupTestDB(t)
	assert.NoError(t, db.Exec(`INSERT INTO dramas (id, team_id, deleted_at) VALUES (1, 1, NULL)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO image_generations (id, drama_id) VALUES (1, 1)`).Error)

	t.Run("Success", func(t *testing.T) {
		err := VerifyImageGenTeam(db, 1, 1)
		assert.NoError(t, err)
	})
	t.Run("Fail_image_gen_not_found", func(t *testing.T) {
		err := VerifyImageGenTeam(db, 999, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "image generation not found")
	})
}

func TestVerifyStoryboardTeam(t *testing.T) {
	db := setupTestDB(t)
	assert.NoError(t, db.Exec(`INSERT INTO dramas (id, team_id, deleted_at) VALUES (1, 1, NULL)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO episodes (id, drama_id, deleted_at) VALUES (1, 1, NULL)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO storyboards (id, episode_id) VALUES (1, 1)`).Error)

	t.Run("Success", func(t *testing.T) {
		err := VerifyStoryboardTeam(db, "1", 1)
		assert.NoError(t, err)
	})
	t.Run("Fail_storyboard_not_found", func(t *testing.T) {
		err := VerifyStoryboardTeam(db, "999", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "storyboard not found")
	})
}

func TestVerifyVideoGenTeam(t *testing.T) {
	db := setupTestDB(t)
	assert.NoError(t, db.Exec(`INSERT INTO dramas (id, team_id, deleted_at) VALUES (1, 1, NULL)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO episodes (id, drama_id, deleted_at) VALUES (1, 1, NULL)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO storyboards (id, episode_id) VALUES (1, 1)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO video_generations (id, storyboard_id) VALUES (1, 1)`).Error)

	t.Run("Success", func(t *testing.T) {
		err := VerifyVideoGenTeam(db, 1, 1)
		assert.NoError(t, err)
	})
	t.Run("Fail_video_gen_not_found", func(t *testing.T) {
		err := VerifyVideoGenTeam(db, 999, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "video generation not found")
	})
}

func TestVerifyVideoMergeTeam(t *testing.T) {
	db := setupTestDB(t)
	assert.NoError(t, db.Exec(`INSERT INTO dramas (id, team_id, deleted_at) VALUES (1, 1, NULL)`).Error)
	assert.NoError(t, db.Exec(`INSERT INTO video_merges (id, drama_id) VALUES (1, 1)`).Error)

	t.Run("Success", func(t *testing.T) {
		err := VerifyVideoMergeTeam(db, 1, 1)
		assert.NoError(t, err)
	})
	t.Run("Fail_video_merge_not_found", func(t *testing.T) {
		err := VerifyVideoMergeTeam(db, 999, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "video merge not found")
	})
}

func TestVerifyAIMessageLogTeam(t *testing.T) {
	db := setupTestDB(t)
	assert.NoError(t, db.Exec(`INSERT INTO dramas (id, team_id, deleted_at) VALUES (1, 1, NULL)`).Error)
	// log with drama_id set
	assert.NoError(t, db.Exec(`INSERT INTO ai_message_logs (id, drama_id) VALUES (1, 1)`).Error)
	// log with drama_id NULL (global log) - should succeed for any team
	assert.NoError(t, db.Exec(`INSERT INTO ai_message_logs (id, drama_id) VALUES (2, NULL)`).Error)

	t.Run("Success_with_drama", func(t *testing.T) {
		err := VerifyAIMessageLogTeam(db, 1, 1)
		assert.NoError(t, err)
	})
	t.Run("Success_when_drama_id_NULL_returns_nil", func(t *testing.T) {
		err := VerifyAIMessageLogTeam(db, 2, 1)
		assert.NoError(t, err)
	})
	t.Run("Fail_log_not_found", func(t *testing.T) {
		err := VerifyAIMessageLogTeam(db, 999, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ai message log not found")
	})
}
