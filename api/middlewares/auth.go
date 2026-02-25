package middlewares

import (
	"strings"

	"github.com/cinemaker/backend/application/services"
	"github.com/cinemaker/backend/pkg/auth"
	"github.com/cinemaker/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "无效的认证格式")
			c.Abort()
			return
		}

		claims, err := authService.ParseToken(parts[1], services.TokenTypeAccess)
		if err != nil {
			response.Unauthorized(c, "登录已过期，请重新登录")
			c.Abort()
			return
		}

		c.Set(auth.KeyUserID, claims.UserID)
		c.Set(auth.KeyTeamID, claims.TeamID)
		c.Set(auth.KeyRole, claims.Role)

		c.Next()
	}
}
