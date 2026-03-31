package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/repository"
	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

var (
	ErrArticleNotFound    = errors.New("article not found")
	ErrSlugExists         = errors.New("article with this slug already exists")
	ErrViewAlreadyCounted = errors.New("view already counted")
)

type ArticleService interface {
	CreateArticle(ctx context.Context, title string, authorID uint) (*models.Article, error)
	UpdateArticle(ctx context.Context, articleID uint, title string, userID uint) (*models.Article, error)
	DeleteArticle(ctx context.Context, articleID uint, userID uint) error
	GetArticle(ctx context.Context, articleID uint) (*models.Article, error)
	GetArticleBySlug(ctx context.Context, slug string) (*models.Article, error)
	GetAllArticles(ctx context.Context, limit, offset int) ([]models.ArticleResponse, error)
	GetArticleViews(ctx context.Context, articleID uint) (uint, error)
	IncrementArticleViews(ctx context.Context, articleID uint, ip string) (uint, error)
}

type viewDedup struct {
	mu   sync.Mutex
	seen map[string]time.Time
}

func newViewDedup(ttl time.Duration) *viewDedup {
	vd := &viewDedup{seen: make(map[string]time.Time)}
	go vd.cleanupLoop(ttl)
	return vd
}

func (vd *viewDedup) Allow(articleID uint, ip string) bool {
	key := fmt.Sprintf("%d:%s", articleID, ip)
	vd.mu.Lock()
	defer vd.mu.Unlock()
	if _, exists := vd.seen[key]; exists {
		return false
	}
	vd.seen[key] = time.Now()
	return true
}

func (vd *viewDedup) cleanupLoop(ttl time.Duration) {
	ticker := time.NewTicker(ttl / 2)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-ttl)
		vd.mu.Lock()
		for k, t := range vd.seen {
			if t.Before(cutoff) {
				delete(vd.seen, k)
			}
		}
		vd.mu.Unlock()
	}
}

type articleService struct {
	repo      repository.ArticleRepository
	userRepo  repository.UserRepository
	viewDedup *viewDedup
}

func NewArticleService(repo repository.ArticleRepository, userRepo repository.UserRepository) ArticleService {
	return &articleService{
		repo:      repo,
		userRepo:  userRepo,
		viewDedup: newViewDedup(24 * time.Hour),
	}
}

func (s *articleService) CreateArticle(ctx context.Context, title string, authorID uint) (*models.Article, error) {
	articleSlug := slug.Make(title)

	existing, _ := s.repo.FindBySlug(ctx, articleSlug)
	if existing != nil {
		articleSlug = articleSlug + "-" + slug.Make(strings.Split(title, " ")[0])
	}

	article := &models.Article{
		Title:    title,
		Slug:     articleSlug,
		AuthorID: authorID,
	}

	if err := s.repo.Create(ctx, article); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, article.ID)
}

func (s *articleService) UpdateArticle(ctx context.Context, articleID uint, title string, userID uint) (*models.Article, error) {
	article, err := s.repo.FindByID(ctx, articleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if article.AuthorID != userID && !user.IsAdmin() {
		return nil, ErrForbidden
	}

	if title != "" {
		article.Title = title
		article.Slug = slug.Make(title)
	}

	if err := s.repo.Update(ctx, article); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, article.ID)
}

func (s *articleService) DeleteArticle(ctx context.Context, articleID uint, userID uint) error {
	article, err := s.repo.FindByID(ctx, articleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrArticleNotFound
		}
		return err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if article.AuthorID != userID && !user.IsAdmin() {
		return ErrForbidden
	}

	return s.repo.Delete(ctx, article)
}

func (s *articleService) GetArticle(ctx context.Context, articleID uint) (*models.Article, error) {
	article, err := s.repo.FindByID(ctx, articleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}
	return article, nil
}

func (s *articleService) GetArticleBySlug(ctx context.Context, slug string) (*models.Article, error) {
	article, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}
	return article, nil
}

func (s *articleService) GetAllArticles(ctx context.Context, limit, offset int) ([]models.ArticleResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	articles, err := s.repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var response []models.ArticleResponse
	for _, article := range articles {
		response = append(response, article.ToResponse())
	}

	return response, nil
}

func (s *articleService) GetArticleViews(ctx context.Context, articleID uint) (uint, error) {
	_, err := s.repo.FindByID(ctx, articleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrArticleNotFound
		}
		return 0, err
	}
	return s.repo.GetViews(ctx, articleID)
}

func (s *articleService) IncrementArticleViews(ctx context.Context, articleID uint, ip string) (uint, error) {
	_, err := s.repo.FindByID(ctx, articleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrArticleNotFound
		}
		return 0, err
	}
	if !s.viewDedup.Allow(articleID, ip) {
		return s.repo.GetViews(ctx, articleID)
	}
	return s.repo.IncrementViews(ctx, articleID)
}
