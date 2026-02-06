package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	"gorm.io/gorm"
)

type fakeCommentRepo struct {
	byID   map[uint]*models.Comment
	byArt  map[string][]*models.Comment
	nextID uint
}

func newFakeCommentRepo() *fakeCommentRepo {
	return &fakeCommentRepo{
		byID:   make(map[uint]*models.Comment),
		byArt:  make(map[string][]*models.Comment),
		nextID: 1,
	}
}

func (r *fakeCommentRepo) Create(_ context.Context, comment *models.Comment) error {
	comment.ID = r.nextID
	r.nextID++
	r.byID[comment.ID] = comment
	r.byArt[comment.ArticleID] = append(r.byArt[comment.ArticleID], comment)
	return nil
}

func (r *fakeCommentRepo) Update(_ context.Context, comment *models.Comment) error {
	r.byID[comment.ID] = comment
	return nil
}

func (r *fakeCommentRepo) Delete(_ context.Context, comment *models.Comment) error {
	delete(r.byID, comment.ID)
	return nil
}

func (r *fakeCommentRepo) FindByID(_ context.Context, id uint) (*models.Comment, error) {
	comment, ok := r.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return comment, nil
}

func (r *fakeCommentRepo) FindByArticleID(_ context.Context, articleID string) ([]models.Comment, error) {
	comments := r.byArt[articleID]
	var res []models.Comment
	for _, c := range comments {
		res = append(res, *c)
	}
	return res, nil
}

func TestCommentService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newFakeCommentRepo()
	svc := NewCommentService(repo)

	created, err := svc.CreateComment(ctx, "hi", "a1", 1)
	if err != nil {
		t.Fatalf("create comment failed: %v", err)
	}

	updated, err := svc.UpdateComment(ctx, created.ID, "updated", 1)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Content != "updated" {
		t.Fatalf("expected updated content")
	}

	if _, err := svc.GetComment(ctx, created.ID); err != nil {
		t.Fatalf("get comment failed: %v", err)
	}

	comments, err := svc.GetCommentsByArticle(ctx, "a1")
	if err != nil || len(comments) != 1 {
		t.Fatalf("expected 1 comment")
	}

	if err := svc.DeleteComment(ctx, created.ID, 1); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = svc.GetComment(ctx, created.ID)
	if !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("expected not found")
	}
}
