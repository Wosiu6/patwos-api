package service

import (
	"context"
	"errors"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/repository"
)

var (
	ErrVoteNotFound    = errors.New("vote not found")
	ErrInvalidVoteType = errors.New("invalid vote type")
	ErrUnauthorized    = errors.New("unauthorized")
)

type VoteService interface {
	Vote(ctx context.Context, articleID uint, userID uint, voteType models.VoteType) error
	RemoveVote(ctx context.Context, articleID uint, userID uint) error
	GetVoteCounts(ctx context.Context, articleID uint, userID *uint) (*models.VoteCounts, error)
}

type voteService struct {
	repo repository.VoteRepository
}

func NewVoteService(repo repository.VoteRepository) VoteService {
	return &voteService{repo: repo}
}

func (s *voteService) Vote(ctx context.Context, articleID uint, userID uint, voteType models.VoteType) error {
	if !voteType.IsValid() {
		return ErrInvalidVoteType
	}

	existingVote, err := s.repo.FindByArticleAndUser(ctx, articleID, userID)
	if err != nil {
		return err
	}

	if existingVote != nil {
		if existingVote.VoteType != voteType {
			existingVote.VoteType = voteType
			return s.repo.Update(ctx, existingVote)
		}
		return nil
	}

	vote := &models.ArticleVote{
		ArticleID: articleID,
		UserID:    userID,
		VoteType:  voteType,
	}

	return s.repo.Create(ctx, vote)
}

func (s *voteService) RemoveVote(ctx context.Context, articleID uint, userID uint) error {
	return s.repo.Delete(ctx, articleID, userID)
}

func (s *voteService) GetVoteCounts(ctx context.Context, articleID uint, userID *uint) (*models.VoteCounts, error) {
	return s.repo.GetVoteCounts(ctx, articleID, userID)
}
