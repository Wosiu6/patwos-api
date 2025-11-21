package controllers

import (
	"net/http"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentController struct {
	DB *gorm.DB
}

func NewCommentController(db *gorm.DB) *CommentController {
	return &CommentController{DB: db}
}

// GetCommentsByArticle gets all comments for a specific article
func (cc *CommentController) GetCommentsByArticle(c *gin.Context) {
	articleID := c.Param("article_id")

	var comments []models.Comment
	if err := cc.DB.Preload("User").Where("article_id = ?", articleID).Order("created_at DESC").Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	var response []models.CommentResponse
	for _, comment := range comments {
		response = append(response, comment.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{"comments": response})
}

// CreateComment creates a new comment (requires authentication)
func (cc *CommentController) CreateComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment := models.Comment{
		Content:   req.Content,
		ArticleID: req.ArticleID,
		UserID:    userID.(uint),
	}

	if err := cc.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	// Reload comment with user data
	if err := cc.DB.Preload("User").First(&comment, comment.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"comment": comment.ToResponse()})
}

// UpdateComment updates an existing comment (requires authentication and ownership)
func (cc *CommentController) UpdateComment(c *gin.Context) {
	commentID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var comment models.Comment
	if err := cc.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	// Check if user owns the comment
	if comment.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own comments"})
		return
	}

	var req models.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment.Content = req.Content
	if err := cc.DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	// Reload comment with user data
	if err := cc.DB.Preload("User").First(&comment, comment.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment.ToResponse()})
}

// DeleteComment deletes a comment (requires authentication and ownership)
func (cc *CommentController) DeleteComment(c *gin.Context) {
	commentID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var comment models.Comment
	if err := cc.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	// Check if user owns the comment
	if comment.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own comments"})
		return
	}

	if err := cc.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

// GetComment gets a single comment by ID
func (cc *CommentController) GetComment(c *gin.Context) {
	commentID := c.Param("id")

	var comment models.Comment
	if err := cc.DB.Preload("User").First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment.ToResponse()})
}
