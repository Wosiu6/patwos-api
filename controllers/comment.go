package controllers

import (
	"net/http"
	"strconv"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/service"
	"github.com/gin-gonic/gin"
)

type CommentController struct {
	service service.CommentService
}

func NewCommentController(commentService service.CommentService) *CommentController {
	return &CommentController{service: commentService}
}

func (cc *CommentController) GetCommentsByArticle(c *gin.Context) {
	articleID := c.Param("article_id")

	comments, err := cc.service.GetCommentsByArticle(articleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

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

	comment, err := cc.service.CreateComment(req.Content, req.ArticleID, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"comment": comment.ToResponse()})
}

func (cc *CommentController) UpdateComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := cc.service.UpdateComment(uint(commentID), req.Content, userID.(uint))
	if err != nil {
		if err == service.ErrCommentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own comments"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment.ToResponse()})
}

func (cc *CommentController) DeleteComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err = cc.service.DeleteComment(uint(commentID), userID.(uint))
	if err != nil {
		if err == service.ErrCommentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}
		if err == service.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own comments"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

func (cc *CommentController) GetComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	comment, err := cc.service.GetComment(uint(commentID))
	if err != nil {
		if err == service.ErrCommentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment.ToResponse()})
}
