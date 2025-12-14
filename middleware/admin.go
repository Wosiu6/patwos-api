package middleware

import (
	"net/http"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		role, ok := userRole.(models.UserRole)
		if !ok || role != models.UserRoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
