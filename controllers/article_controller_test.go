package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/service"
	"github.com/gin-gonic/gin"
)

type fakeArticleService struct {
	listFn    func(ctx context.Context, limit, offset int) ([]models.ArticleResponse, error)
	getFn     func(ctx context.Context, id uint) (*models.Article, error)
	getSlugFn func(ctx context.Context, slug string) (*models.Article, error)
}

func (f *fakeArticleService) CreateArticle(context.Context, string, uint) (*models.Article, error) {
	return nil, nil
}
func (f *fakeArticleService) UpdateArticle(context.Context, uint, string, uint) (*models.Article, error) {
	return nil, nil
}
func (f *fakeArticleService) DeleteArticle(context.Context, uint, uint) error {
	return nil
}
func (f *fakeArticleService) GetArticle(ctx context.Context, id uint) (*models.Article, error) {
	return f.getFn(ctx, id)
}
func (f *fakeArticleService) GetArticleBySlug(ctx context.Context, slug string) (*models.Article, error) {
	return f.getSlugFn(ctx, slug)
}
func (f *fakeArticleService) GetAllArticles(ctx context.Context, limit, offset int) ([]models.ArticleResponse, error) {
	return f.listFn(ctx, limit, offset)
}
func (f *fakeArticleService) GetArticleViews(context.Context, uint) (uint, error) {
	return 0, nil
}
func (f *fakeArticleService) IncrementArticleViews(context.Context, uint) (uint, error) {
	return 0, nil
}

func TestArticleController_GetArticlesAndGetArticle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller := NewArticleController(&fakeArticleService{
		listFn: func(context.Context, int, int) ([]models.ArticleResponse, error) {
			return []models.ArticleResponse{{
				ID:        1,
				Title:     "t",
				Slug:      "s",
				Author:    models.UserResponse{ID: 1, Username: "u"},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Views:     0,
			}}, nil
		},
		getFn: func(context.Context, uint) (*models.Article, error) {
			return &models.Article{ID: 1, Title: "t", Slug: "s", Author: models.User{ID: 1}}, nil
		},
		getSlugFn: func(context.Context, string) (*models.Article, error) {
			return nil, service.ErrArticleNotFound
		},
	})

	r := gin.New()
	r.GET("/articles", controller.GetArticles)
	r.GET("/articles/:id", controller.GetArticle)

	listReq := httptest.NewRequest(http.MethodGet, "/articles", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listW.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/articles/1", nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getW.Code)
	}
}
