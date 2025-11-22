package routes

import (
	"github.com/Wosiu6/patwos-api/config"
	"github.com/Wosiu6/patwos-api/controllers"
	"github.com/Wosiu6/patwos-api/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authController := controllers.NewAuthController(db, cfg)
	commentController := controllers.NewCommentController(db)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", middleware.StrictRateLimitMiddleware(), authController.Register)
			auth.POST("/login", middleware.StrictRateLimitMiddleware(), authController.Login)
			auth.GET("/me", middleware.AuthMiddleware(db, cfg), authController.GetCurrentUser)
		}

		comments := v1.Group("/comments")
		{
			comments.GET("/article/:article_id", commentController.GetCommentsByArticle)
			comments.GET("/:id", commentController.GetComment)

			comments.POST("", middleware.AuthMiddleware(db, cfg), commentController.CreateComment)
			comments.PUT("/:id", middleware.AuthMiddleware(db, cfg), commentController.UpdateComment)
			comments.PATCH("/:id", middleware.AuthMiddleware(db, cfg), commentController.UpdateComment)
			comments.DELETE("/:id", middleware.AuthMiddleware(db, cfg), commentController.DeleteComment)
		}
	}
}
