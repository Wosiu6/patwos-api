package service

import (
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
	Vote(articleID string, userID uint, voteType models.VoteType) error
	RemoveVote(articleID string, userID uint) error
	GetVoteCounts(articleID string, userID *uint) (*models.VoteCounts, error)
}

type voteService struct {
	repo repository.VoteRepository
}

func NewVoteService(repo repository.VoteRepository) VoteService {
	return &voteService{repo: repo}
}

func (s *voteService) Vote(articleID string, userID uint, voteType models.VoteType) error {
	if !voteType.IsValid() {
		return ErrInvalidVoteType
	}

	existingVote, err := s.repo.FindByArticleAndUser(articleID, userID)
	if err != nil {
		return err
	}

	if existingVote != nil {
		if existingVote.VoteType != voteType {
			existingVote.VoteType = voteType
			return s.repo.Update(existingVote)
		}
		return nil
	}

	vote := &models.ArticleVote{
		ArticleID: articleID,
		UserID:    userID,
		VoteType:  voteType,
	}

	return s.repo.Create(vote)
}

func (s *voteService) RemoveVote(articleID string, userID uint) error {
	return s.repo.Delete(articleID, userID)
}

func (s *voteService) GetVoteCounts(articleID string, userID *uint) (*models.VoteCounts, error) {
	return s.repo.GetVoteCounts(articleID, userID)
}
