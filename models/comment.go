package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Content   string         `gorm:"type:text;not null" json:"content" binding:"required,min=1,max=5000"`
	ArticleID string         `gorm:"not null;index" json:"article_id" binding:"required"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type CreateCommentRequest struct {
	Content   string `json:"content" binding:"required,min=1,max=5000"`
	ArticleID string `json:"article_id" binding:"required"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

type CommentResponse struct {
	ID        uint         `json:"id"`
	Content   string       `json:"content"`
	ArticleID string       `json:"article_id"`
	UserID    uint         `json:"user_id"`
	User      UserResponse `json:"user"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func (c *Comment) ToResponse() CommentResponse {
	return CommentResponse{
		ID:        c.ID,
		Content:   c.Content,
		ArticleID: c.ArticleID,
		UserID:    c.UserID,
		User:      c.User.ToResponse(),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
