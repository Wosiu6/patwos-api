package routes

import (
	"github.com/Wosiu6/patwos-api/config"
	"github.com/Wosiu6/patwos-api/controllers"
	"github.com/Wosiu6/patwos-api/middleware"
	"github.com/Wosiu6/patwos-api/repository"
	"github.com/Wosiu6/patwos-api/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	userRepo := repository.NewUserRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	voteRepo := repository.NewVoteRepository(db)

	authService := service.NewAuthService(userRepo, cfg)
	commentService := service.NewCommentService(commentRepo)
	voteService := service.NewVoteService(voteRepo)

	authController := controllers.NewAuthController(authService)
	commentController := controllers.NewCommentController(commentService)
	voteController := controllers.NewVoteController(voteService)

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

		votes := v1.Group("/votes")
		{
			votes.GET("/:article_id", voteController.GetVoteCounts)

			votes.POST("", middleware.AuthMiddleware(db, cfg), voteController.Vote)
			votes.DELETE("/:article_id", middleware.AuthMiddleware(db, cfg), voteController.RemoveVote)
		}
	}
}
