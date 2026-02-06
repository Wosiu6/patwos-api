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

type fakeCommentService struct {
	createFn func(ctx context.Context, content, articleID string, userID uint) (*models.Comment, error)
	updateFn func(ctx context.Context, commentID uint, content string, userID uint) (*models.Comment, error)
	deleteFn func(ctx context.Context, commentID uint, userID uint) error
	getFn    func(ctx context.Context, commentID uint) (*models.Comment, error)
	listFn   func(ctx context.Context, articleID string) ([]models.CommentResponse, error)
}

func (f *fakeCommentService) CreateComment(ctx context.Context, content, articleID string, userID uint) (*models.Comment, error) {
	return f.createFn(ctx, content, articleID, userID)
}
func (f *fakeCommentService) UpdateComment(ctx context.Context, commentID uint, content string, userID uint) (*models.Comment, error) {
	return f.updateFn(ctx, commentID, content, userID)
}
func (f *fakeCommentService) DeleteComment(ctx context.Context, commentID uint, userID uint) error {
	return f.deleteFn(ctx, commentID, userID)
}
func (f *fakeCommentService) GetComment(ctx context.Context, commentID uint) (*models.Comment, error) {
	return f.getFn(ctx, commentID)
}
func (f *fakeCommentService) GetCommentsByArticle(ctx context.Context, articleID string) ([]models.CommentResponse, error) {
	return f.listFn(ctx, articleID)
}

func TestCommentController_CreateAndUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller := NewCommentController(&fakeCommentService{
		createFn: func(context.Context, string, string, uint) (*models.Comment, error) {
			return &models.Comment{ID: 1, Content: "hi", ArticleID: "a1", UserID: 1}, nil
		},
		updateFn: func(context.Context, uint, string, uint) (*models.Comment, error) {
			return nil, service.ErrCommentNotFound
		},
	})

	r := gin.New()
	r.POST("/comments", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		controller.CreateComment(c)
	})
	r.PUT("/comments/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		controller.UpdateComment(c)
	})

	createBody, _ := json.Marshal(models.CreateCommentRequest{Content: "hi", ArticleID: "a1"})
	createReq := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createW.Code)
	}

	updateBody, _ := json.Marshal(models.UpdateCommentRequest{Content: "x"})
	updateReq := httptest.NewRequest(http.MethodPut, "/comments/1", bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateW := httptest.NewRecorder()
	r.ServeHTTP(updateW, updateReq)
	if updateW.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", updateW.Code)
	}
}

func TestCommentController_GetAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller := NewCommentController(&fakeCommentService{
		getFn: func(context.Context, uint) (*models.Comment, error) {
			return &models.Comment{ID: 1, Content: "hi", ArticleID: "a1", UserID: 1}, nil
		},
		deleteFn: func(context.Context, uint, uint) error {
			return nil
		},
	})

	r := gin.New()
	r.GET("/comments/:id", controller.GetComment)
	r.DELETE("/comments/:id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		controller.DeleteComment(c)
	})

	getReq := httptest.NewRequest(http.MethodGet, "/comments/1", nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getW.Code)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/comments/1", nil)
	deleteW := httptest.NewRecorder()
	r.ServeHTTP(deleteW, deleteReq)
	if deleteW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", deleteW.Code)
	}
}
