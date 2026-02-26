package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TeamScope filters queries by team_id. Returns empty result set when teamID is 0
// to prevent accidental data leakage across teams.
func TeamScope(teamID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if teamID == 0 {
			return db.Where("1 = 0")
		}
		return db.Where("team_id = ?", teamID)
	}
}

// MustGetTeamID extracts team_id from gin context; returns error if missing or zero.
func MustGetTeamID(c *gin.Context) (uint, error) {
	id := GetTeamID(c)
	if id == 0 {
		return 0, errors.New("missing team context")
	}
	return id, nil
}

// MustGetUserID extracts user_id from gin context; returns error if missing or zero.
func MustGetUserID(c *gin.Context) (uint, error) {
	id := GetUserID(c)
	if id == 0 {
		return 0, errors.New("missing user context")
	}
	return id, nil
}
