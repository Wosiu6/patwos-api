package controllers

import (
	"net/http"
	"strconv"

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

	if err := vc.service.Vote(c.Request.Context(), req.ArticleID, userID.(uint), req.VoteType); err != nil {
		if err == service.ErrInvalidVoteType {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vote type. Use 'like' or 'dislike'"})
			return
		}
		gin.DefaultWriter.Write([]byte("[VOTE-ERROR] ArticleID: " + strconv.Itoa(int(req.ArticleID)) + " | UserID: " + strconv.Itoa(int(userID.(uint))) + " | Error: " + err.Error() + "\n"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process vote", "details": err.Error()})
		return
	}

	uid := userID.(uint)
	counts, err := vc.service.GetVoteCounts(c.Request.Context(), req.ArticleID, &uid)
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

	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	if err := vc.service.RemoveVote(c.Request.Context(), uint(articleID), userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove vote"})
		return
	}

	uid := userID.(uint)
	counts, err := vc.service.GetVoteCounts(c.Request.Context(), uint(articleID), &uid)
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
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	var userIDPtr *uint
	if userID, exists := c.Get("user_id"); exists {
		uid := userID.(uint)
		userIDPtr = &uid
	}

	counts, err := vc.service.GetVoteCounts(c.Request.Context(), uint(articleID), userIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vote counts"})
		return
	}

	c.JSON(http.StatusOK, counts)
}
