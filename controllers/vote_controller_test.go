package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/service"
	"github.com/gin-gonic/gin"
)

type fakeVoteService struct {
	voteFn   func(ctx context.Context, articleID uint, userID uint, voteType models.VoteType) error
	removeFn func(ctx context.Context, articleID uint, userID uint) error
	countFn  func(ctx context.Context, articleID uint, userID *uint) (*models.VoteCounts, error)
}

func (f *fakeVoteService) Vote(ctx context.Context, articleID uint, userID uint, voteType models.VoteType) error {
	return f.voteFn(ctx, articleID, userID, voteType)
}
func (f *fakeVoteService) RemoveVote(ctx context.Context, articleID uint, userID uint) error {
	return f.removeFn(ctx, articleID, userID)
}
func (f *fakeVoteService) GetVoteCounts(ctx context.Context, articleID uint, userID *uint) (*models.VoteCounts, error) {
	return f.countFn(ctx, articleID, userID)
}

func TestVoteController_VoteAndRemove(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller := NewVoteController(&fakeVoteService{
		voteFn: func(context.Context, uint, uint, models.VoteType) error {
			return service.ErrInvalidVoteType
		},
		removeFn: func(context.Context, uint, uint) error {
			return nil
		},
		countFn: func(context.Context, uint, *uint) (*models.VoteCounts, error) {
			return &models.VoteCounts{ArticleID: 1, Likes: 1}, nil
		},
	})

	r := gin.New()
	r.POST("/votes", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		controller.Vote(c)
	})
	r.DELETE("/votes/:article_id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		controller.RemoveVote(c)
	})

	body, _ := json.Marshal(models.VoteRequest{ArticleID: 1, VoteType: "bad"})
	voteReq := httptest.NewRequest(http.MethodPost, "/votes", bytes.NewReader(body))
	voteReq.Header.Set("Content-Type", "application/json")
	voteW := httptest.NewRecorder()
	r.ServeHTTP(voteW, voteReq)
	if voteW.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", voteW.Code)
	}

	removeReq := httptest.NewRequest(http.MethodDelete, "/votes/1", nil)
	removeW := httptest.NewRecorder()
	r.ServeHTTP(removeW, removeReq)
	if removeW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", removeW.Code)
	}
}
