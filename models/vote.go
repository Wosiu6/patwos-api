package models

import (
	"time"

	"gorm.io/gorm"
)

type VoteType string

const (
	VoteLike    VoteType = "like"
	VoteDislike VoteType = "dislike"
)

type ArticleVote struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	ArticleID string         `gorm:"not null;index:idx_article_user,unique" json:"article_id"`
	UserID    uint           `gorm:"not null;index:idx_article_user,unique" json:"user_id"`
	VoteType  VoteType       `gorm:"type:varchar(10);not null" json:"vote_type"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type VoteRequest struct {
	ArticleID string   `json:"article_id" binding:"required"`
	VoteType  VoteType `json:"vote_type" binding:"required,oneof=like dislike"`
}

type VoteCounts struct {
	ArticleID    string `json:"article_id"`
	Likes        int64  `json:"likes"`
	Dislikes     int64  `json:"dislikes"`
	UserVote     string `json:"user_vote,omitempty"`
	UserHasVoted bool   `json:"user_has_voted"`
}

func (ArticleVote) TableName() string {
	return "article_votes"
}

func (v VoteType) IsValid() bool {
	return v == VoteLike || v == VoteDislike
}
