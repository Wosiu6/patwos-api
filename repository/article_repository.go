package repository

import (
	"context"

	"github.com/Wosiu6/patwos-api/models"
	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(ctx context.Context, article *models.Article) error
	Update(ctx context.Context, article *models.Article) error
	Delete(ctx context.Context, article *models.Article) error
	FindByID(ctx context.Context, id uint) (*models.Article, error)
	FindBySlug(ctx context.Context, slug string) (*models.Article, error)
	FindAll(ctx context.Context, limit, offset int) ([]models.Article, error)
	GetViews(ctx context.Context, id uint) (uint, error)
	IncrementViews(ctx context.Context, id uint) (uint, error)
}

type articleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &articleRepository{db: db}
}

func (r *articleRepository) Create(ctx context.Context, article *models.Article) error {
	return r.db.WithContext(ctx).Create(article).Error
}

func (r *articleRepository) Update(ctx context.Context, article *models.Article) error {
	return r.db.WithContext(ctx).Save(article).Error
}

func (r *articleRepository) Delete(ctx context.Context, article *models.Article) error {
	return r.db.WithContext(ctx).Delete(article).Error
}

func (r *articleRepository) FindByID(ctx context.Context, id uint) (*models.Article, error) {
	var article models.Article
	err := r.db.WithContext(ctx).Preload("Author").First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *articleRepository) FindBySlug(ctx context.Context, slug string) (*models.Article, error) {
	var article models.Article
	err := r.db.WithContext(ctx).Preload("Author").Where("slug = ?", slug).First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *articleRepository) FindAll(ctx context.Context, limit, offset int) ([]models.Article, error) {
	var articles []models.Article
	err := r.db.WithContext(ctx).Preload("Author").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&articles).Error
	return articles, err
}

func (r *articleRepository) GetViews(ctx context.Context, id uint) (uint, error) {
	var views uint
	err := r.db.WithContext(ctx).Model(&models.Article{}).Select("views").Where("id = ?", id).Scan(&views).Error
	if err != nil {
		return 0, err
	}
	return views, nil
}

func (r *articleRepository) IncrementViews(ctx context.Context, id uint) (uint, error) {
	if err := r.db.WithContext(ctx).Model(&models.Article{}).
		Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + ?", 1)).Error; err != nil {
		return 0, err
	}
	return r.GetViews(ctx, id)
}
