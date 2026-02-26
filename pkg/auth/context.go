package auth

import "github.com/gin-gonic/gin"

const (
	KeyUserID = "userID"
	KeyTeamID = "teamID"
	KeyRole   = "role"
)

func GetUserID(c *gin.Context) uint {
	v, exists := c.Get(KeyUserID)
	if !exists {
		return 0
	}
	id, ok := v.(uint)
	if !ok {
		return 0
	}
	return id
}

func GetTeamID(c *gin.Context) uint {
	v, exists := c.Get(KeyTeamID)
	if !exists {
		return 0
	}
	id, ok := v.(uint)
	if !ok {
		return 0
	}
	return id
}

func GetRole(c *gin.Context) string {
	v, exists := c.Get(KeyRole)
	if !exists {
		return ""
	}
	role, ok := v.(string)
	if !ok {
		return ""
	}
	return role
}
