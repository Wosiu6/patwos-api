package models

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Title     string         `gorm:"not null;index" json:"title" binding:"required,min=3,max=200"`
	Slug      string         `gorm:"uniqueIndex;not null" json:"slug"`
	Content   string         `gorm:"type:text;not null" json:"content" binding:"required,min=10"`
	AuthorID  uint           `gorm:"not null;index" json:"author_id"`
	Author    User           `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Comments  []Comment      `gorm:"foreignKey:ArticleID" json:"comments,omitempty"`
	Votes     []ArticleVote  `gorm:"foreignKey:ArticleID" json:"votes,omitempty"`
}

type CreateArticleRequest struct {
	Title   string `json:"title" binding:"required,min=3,max=200"`
	Content string `json:"content" binding:"required,min=10"`
}

type UpdateArticleRequest struct {
	Title   string `json:"title" binding:"omitempty,min=3,max=200"`
	Content string `json:"content" binding:"omitempty,min=10"`
}

type ArticleResponse struct {
	ID        uint         `json:"id"`
	Title     string       `json:"title"`
	Slug      string       `json:"slug"`
	Content   string       `json:"content"`
	Author    UserResponse `json:"author"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func (a *Article) ToResponse() ArticleResponse {
	return ArticleResponse{
		ID:        a.ID,
		Title:     a.Title,
		Slug:      a.Slug,
		Content:   a.Content,
		Author:    a.Author.ToResponse(),
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}
