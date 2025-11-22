package repository

import (
	"github.com/Wosiu6/patwos-api/models"
	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(comment *models.Comment) error
	Update(comment *models.Comment) error
	Delete(comment *models.Comment) error
	FindByID(id uint) (*models.Comment, error)
	FindByArticleID(articleID string) ([]models.Comment, error)
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *commentRepository) Update(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

func (r *commentRepository) Delete(comment *models.Comment) error {
	return r.db.Delete(comment).Error
}

func (r *commentRepository) FindByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.Preload("User").First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) FindByArticleID(articleID string) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.Preload("User").
		Where("article_id = ?", articleID).
		Order("created_at DESC").
		Find(&comments).Error
	return comments, err
}
