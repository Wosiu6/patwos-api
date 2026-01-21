package middleware

import (
	"net/http"
	"strings"

	"github.com/Wosiu6/patwos-api/config"
	"github.com/Wosiu6/patwos-api/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type ErrorMessage string

const (
	ErrUnauthorized   ErrorMessage = "unauthorized"
	ErrTokenExpired   ErrorMessage = "token_expired"
	ErrTokenRevoked   ErrorMessage = "token_revoked"
	ErrTokenInvalid   ErrorMessage = "token_invalid"
	ErrSessionExpired ErrorMessage = "session_expired"
)

func AuthMiddleware(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized})
			c.Abort()
			return
		}

		var revokedToken models.RevokedToken
		if err := db.Where("token = ?", tokenString).First(&revokedToken).Error; err == nil {
			gin.DefaultWriter.Write([]byte("[AUTH-FAILED] Revoked token | IP: " + c.ClientIP() + " | Path: " + c.Request.URL.Path + " | Status: 401\n"))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   string(ErrTokenRevoked),
				"message": "Your session has been logged out. Please log in again.",
				"code":    "TOKEN_REVOKED",
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || token == nil || !token.Valid {
			gin.DefaultWriter.Write([]byte("[AUTH-FAILED] Invalid token | IP: " + c.ClientIP() + " | Path: " + c.Request.URL.Path + " | Error: " + func() string {
				if err != nil {
					return err.Error()
				}
				return "invalid"
			}() + " | Status: 401\n"))

			if err == jwt.ErrTokenExpired {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   string(ErrTokenExpired),
					"message": "Your session has expired. Please log in again.",
					"code":    "TOKEN_EXPIRED",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   string(ErrTokenInvalid),
					"message": "Invalid authentication token. Please log in again.",
					"code":    "TOKEN_INVALID",
				})
			}
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized})
			c.Abort()
			return
		}

		userState, ok := claims["state"].(float64)
		if !ok || userState != float64(models.UserStatusActive) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized})
			c.Abort()
			return
		}

		userRole, ok := claims["role"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized})
			c.Abort()
			return
		}

		var user models.User
		if err := db.First(&user, uint(userID)).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", models.UserRole(userRole))
		c.Next()
	}
}
