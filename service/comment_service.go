package service

import (
	"context"
	"errors"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/repository"
	"gorm.io/gorm"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrForbidden       = errors.New("forbidden: you can only modify your own comments")
)

type CommentService interface {
	CreateComment(ctx context.Context, content, articleID string, userID uint) (*models.Comment, error)
	UpdateComment(ctx context.Context, commentID uint, content string, userID uint) (*models.Comment, error)
	DeleteComment(ctx context.Context, commentID uint, userID uint) error
	GetComment(ctx context.Context, commentID uint) (*models.Comment, error)
	GetCommentsByArticle(ctx context.Context, articleID string) ([]models.CommentResponse, error)
}

type commentService struct {
	repo repository.CommentRepository
}

func NewCommentService(repo repository.CommentRepository) CommentService {
	return &commentService{repo: repo}
}

func (s *commentService) CreateComment(ctx context.Context, content, articleID string, userID uint) (*models.Comment, error) {
	comment := &models.Comment{
		Content:   content,
		ArticleID: articleID,
		UserID:    userID,
	}

	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, comment.ID)
}

func (s *commentService) UpdateComment(ctx context.Context, commentID uint, content string, userID uint) (*models.Comment, error) {
	comment, err := s.repo.FindByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	if comment.UserID != userID {
		return nil, ErrForbidden
	}

	comment.Content = content
	if err := s.repo.Update(ctx, comment); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, comment.ID)
}

func (s *commentService) DeleteComment(ctx context.Context, commentID uint, userID uint) error {
	comment, err := s.repo.FindByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommentNotFound
		}
		return err
	}

	if comment.UserID != userID {
		return ErrForbidden
	}

	return s.repo.Delete(ctx, comment)
}

func (s *commentService) GetComment(ctx context.Context, commentID uint) (*models.Comment, error) {
	comment, err := s.repo.FindByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return comment, nil
}

func (s *commentService) GetCommentsByArticle(ctx context.Context, articleID string) ([]models.CommentResponse, error) {
	comments, err := s.repo.FindByArticleID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	var response []models.CommentResponse
	for _, comment := range comments {
		response = append(response, comment.ToResponse())
	}

	return response, nil
}
