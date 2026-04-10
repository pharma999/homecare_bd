package middleware

import (
	"strings"

	"home_care_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	UserIDKey      = "user_id"
	PhoneNumberKey = "phone_number"
	UserRoleKey    = "user_role"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "Authorization header required")
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.UnauthorizedResponse(c, "Format: Bearer {token}")
			c.Abort()
			return
		}
		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			utils.UnauthorizedResponse(c, "Invalid or expired token")
			c.Abort()
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Set(PhoneNumberKey, claims.PhoneNumber)
		c.Set(UserRoleKey, claims.Role)
		c.Next()
	}
}

func RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, _ := c.Get(UserRoleKey)
		userRole, _ := roleVal.(string)
		for _, r := range roles {
			if strings.EqualFold(userRole, r) {
				c.Next()
				return
			}
		}
		utils.ForbiddenResponse(c, "Insufficient permissions")
		c.Abort()
	}
}

func GetUserID(c *gin.Context) string {
	v, _ := c.Get(UserIDKey)
	id, _ := v.(string)
	return id
}

func GetUserRole(c *gin.Context) string {
	v, _ := c.Get(UserRoleKey)
	r, _ := v.(string)
	return r
}
