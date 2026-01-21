package repository

import (
	"github.com/Wosiu6/patwos-api/models"
	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(article *models.Article) error
	Update(article *models.Article) error
	Delete(article *models.Article) error
	FindByID(id uint) (*models.Article, error)
	FindBySlug(slug string) (*models.Article, error)
	FindAll(limit, offset int) ([]models.Article, error)
	GetViews(id uint) (uint, error)
	IncrementViews(id uint) (uint, error)
}

type articleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &articleRepository{db: db}
}

func (r *articleRepository) Create(article *models.Article) error {
	return r.db.Create(article).Error
}

func (r *articleRepository) Update(article *models.Article) error {
	return r.db.Save(article).Error
}

func (r *articleRepository) Delete(article *models.Article) error {
	return r.db.Delete(article).Error
}

func (r *articleRepository) FindByID(id uint) (*models.Article, error) {
	var article models.Article
	err := r.db.Preload("Author").First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *articleRepository) FindBySlug(slug string) (*models.Article, error) {
	var article models.Article
	err := r.db.Preload("Author").Where("slug = ?", slug).First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *articleRepository) FindAll(limit, offset int) ([]models.Article, error) {
	var articles []models.Article
	err := r.db.Preload("Author").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&articles).Error
	return articles, err
}

func (r *articleRepository) GetViews(id uint) (uint, error) {
	var views uint
	err := r.db.Model(&models.Article{}).Select("views").Where("id = ?", id).Scan(&views).Error
	if err != nil {
		return 0, err
	}
	return views, nil
}

func (r *articleRepository) IncrementViews(id uint) (uint, error) {
	if err := r.db.Model(&models.Article{}).
		Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + ?", 1)).Error; err != nil {
		return 0, err
	}
	return r.GetViews(id)
}
