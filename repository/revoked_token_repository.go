package repository

import (
	"context"
	"time"

	"github.com/Wosiu6/patwos-api/models"
	"gorm.io/gorm"
)

type RevokedTokenRepository interface {
	Create(ctx context.Context, token *models.RevokedToken) error
	ExistsByToken(ctx context.Context, token string) (bool, error)
	FindByToken(ctx context.Context, token string) (*models.RevokedToken, error)
}

type revokedTokenRepository struct {
	db *gorm.DB
}

func NewRevokedTokenRepository(db *gorm.DB) RevokedTokenRepository {
	return &revokedTokenRepository{db: db}
}

func (r *revokedTokenRepository) Create(ctx context.Context, token *models.RevokedToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *revokedTokenRepository) ExistsByToken(ctx context.Context, token string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.RevokedToken{}).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		Count(&count).Error
	return count > 0, err
}

func (r *revokedTokenRepository) FindByToken(ctx context.Context, token string) (*models.RevokedToken, error) {
	var revoked models.RevokedToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&revoked).Error
	if err != nil {
		return nil, err
	}
	return &revoked, nil
}
