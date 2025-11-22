package controllers

import (
	"net/http"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/service"
	"github.com/gin-gonic/gin"
)

type VoteController struct {
	service service.VoteService
}

func NewVoteController(service service.VoteService) *VoteController {
	return &VoteController{service: service}
}

func (vc *VoteController) Vote(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := vc.service.Vote(req.ArticleID, userID.(uint), req.VoteType); err != nil {
		if err == service.ErrInvalidVoteType {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vote type. Use 'like' or 'dislike'"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process vote"})
		return
	}

	uid := userID.(uint)
	counts, err := vc.service.GetVoteCounts(req.ArticleID, &uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vote counts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vote recorded successfully",
		"counts":  counts,
	})
}

func (vc *VoteController) RemoveVote(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	articleID := c.Param("article_id")
	if articleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Article ID is required"})
		return
	}

	if err := vc.service.RemoveVote(articleID, userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove vote"})
		return
	}

	uid := userID.(uint)
	counts, err := vc.service.GetVoteCounts(articleID, &uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vote counts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vote removed successfully",
		"counts":  counts,
	})
}

func (vc *VoteController) GetVoteCounts(c *gin.Context) {
	articleID := c.Param("article_id")
	if articleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Article ID is required"})
		return
	}

	var userIDPtr *uint
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(uint)
		userIDPtr = &uid
	}

	counts, err := vc.service.GetVoteCounts(articleID, userIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vote counts"})
		return
	}

	c.JSON(http.StatusOK, counts)
}
