package repository

import (
	"context"

	"github.com/Wosiu6/patwos-api/models"
	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *models.Comment) error
	Update(ctx context.Context, comment *models.Comment) error
	Delete(ctx context.Context, comment *models.Comment) error
	FindByID(ctx context.Context, id uint) (*models.Comment, error)
	FindByArticleID(ctx context.Context, articleID string) ([]models.Comment, error)
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *models.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepository) Update(ctx context.Context, comment *models.Comment) error {
	return r.db.WithContext(ctx).Save(comment).Error
}

func (r *commentRepository) Delete(ctx context.Context, comment *models.Comment) error {
	return r.db.WithContext(ctx).Delete(comment).Error
}

func (r *commentRepository) FindByID(ctx context.Context, id uint) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.WithContext(ctx).Preload("User").First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) FindByArticleID(ctx context.Context, articleID string) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.WithContext(ctx).Preload("User").
		Where("article_id = ?", articleID).
		Order("created_at DESC").
		Find(&comments).Error
	return comments, err
}
