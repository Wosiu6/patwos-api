package repository

import (
	"github.com/Wosiu6/patwos-api/models"
	"gorm.io/gorm"
)

type VoteRepository interface {
	Create(vote *models.ArticleVote) error
	Update(vote *models.ArticleVote) error
	Delete(articleID string, userID uint) error
	FindByArticleAndUser(articleID string, userID uint) (*models.ArticleVote, error)
	CountByArticleAndType(articleID string, voteType models.VoteType) (int64, error)
	GetVoteCounts(articleID string, userID *uint) (*models.VoteCounts, error)
}

type voteRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) VoteRepository {
	return &voteRepository{db: db}
}

func (r *voteRepository) Create(vote *models.ArticleVote) error {
	return r.db.Create(vote).Error
}

func (r *voteRepository) Update(vote *models.ArticleVote) error {
	return r.db.Save(vote).Error
}

func (r *voteRepository) Delete(articleID string, userID uint) error {
	return r.db.Where("article_id = ? AND user_id = ?", articleID, userID).
		Delete(&models.ArticleVote{}).Error
}

func (r *voteRepository) FindByArticleAndUser(articleID string, userID uint) (*models.ArticleVote, error) {
	var vote models.ArticleVote
	err := r.db.Where("article_id = ? AND user_id = ?", articleID, userID).
		First(&vote).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &vote, err
}

func (r *voteRepository) CountByArticleAndType(articleID string, voteType models.VoteType) (int64, error) {
	var count int64
	err := r.db.Model(&models.ArticleVote{}).
		Where("article_id = ? AND vote_type = ?", articleID, voteType).
		Count(&count).Error
	return count, err
}

func (r *voteRepository) GetVoteCounts(articleID string, userID *uint) (*models.VoteCounts, error) {
	counts := &models.VoteCounts{
		ArticleID: articleID,
	}

	likes, err := r.CountByArticleAndType(articleID, models.VoteLike)
	if err != nil {
		return nil, err
	}
	counts.Likes = likes

	dislikes, err := r.CountByArticleAndType(articleID, models.VoteDislike)
	if err != nil {
		return nil, err
	}
	counts.Dislikes = dislikes

	if userID == nil {
		return counts, nil
	}

	vote, err := r.FindByArticleAndUser(articleID, *userID)
	if err != nil {
		return nil, err
	}

	if vote != nil {
		counts.UserVote = string(vote.VoteType)
		counts.UserHasVoted = true
	}

	return counts, nil
}
