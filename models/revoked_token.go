package models

import (
	"time"

	"gorm.io/gorm"
)

type RevokedToken struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Token     string         `gorm:"uniqueIndex;not null" json:"-"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	RevokedAt time.Time      `gorm:"not null" json:"revoked_at"`
	ExpiresAt time.Time      `gorm:"not null;index" json:"expires_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
