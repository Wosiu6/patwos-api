package repository

import (
	"context"

	"github.com/Wosiu6/patwos-api/models"
	"gorm.io/gorm"
)

type VoteRepository interface {
	Create(ctx context.Context, vote *models.ArticleVote) error
	Update(ctx context.Context, vote *models.ArticleVote) error
	Delete(ctx context.Context, articleID uint, userID uint) error
	FindByArticleAndUser(ctx context.Context, articleID uint, userID uint) (*models.ArticleVote, error)
	CountByArticleAndType(ctx context.Context, articleID uint, voteType models.VoteType) (int64, error)
	GetVoteCounts(ctx context.Context, articleID uint, userID *uint) (*models.VoteCounts, error)
}

type voteRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) VoteRepository {
	return &voteRepository{db: db}
}

func (r *voteRepository) Create(ctx context.Context, vote *models.ArticleVote) error {
	return r.db.WithContext(ctx).Create(vote).Error
}

func (r *voteRepository) Update(ctx context.Context, vote *models.ArticleVote) error {
	return r.db.WithContext(ctx).Save(vote).Error
}

func (r *voteRepository) Delete(ctx context.Context, articleID uint, userID uint) error {
	return r.db.WithContext(ctx).Where("article_id = ? AND user_id = ?", articleID, userID).
		Delete(&models.ArticleVote{}).Error
}

func (r *voteRepository) FindByArticleAndUser(ctx context.Context, articleID uint, userID uint) (*models.ArticleVote, error) {
	var vote models.ArticleVote
	err := r.db.WithContext(ctx).Where("article_id = ? AND user_id = ?", articleID, userID).
		First(&vote).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &vote, err
}

func (r *voteRepository) CountByArticleAndType(ctx context.Context, articleID uint, voteType models.VoteType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.ArticleVote{}).
		Where("article_id = ? AND vote_type = ?", articleID, voteType).
		Count(&count).Error
	return count, err
}

func (r *voteRepository) GetVoteCounts(ctx context.Context, articleID uint, userID *uint) (*models.VoteCounts, error) {
	counts := &models.VoteCounts{
		ArticleID: articleID,
	}

	likes, err := r.CountByArticleAndType(ctx, articleID, models.VoteLike)
	if err != nil {
		return nil, err
	}
	counts.Likes = likes

	dislikes, err := r.CountByArticleAndType(ctx, articleID, models.VoteDislike)
	if err != nil {
		return nil, err
	}
	counts.Dislikes = dislikes

	if userID == nil {
		return counts, nil
	}

	vote, err := r.FindByArticleAndUser(ctx, articleID, *userID)
	if err != nil {
		return nil, err
	}

	if vote != nil {
		counts.UserVote = string(vote.VoteType)
		counts.UserHasVoted = true
	}

	return counts, nil
}
