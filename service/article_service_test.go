package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	"gorm.io/gorm"
)

type fakeArticleRepo struct {
	byID   map[uint]*models.Article
	bySlug map[string]*models.Article
	views  map[uint]uint
	nextID uint
}

func newFakeArticleRepo() *fakeArticleRepo {
	return &fakeArticleRepo{
		byID:   make(map[uint]*models.Article),
		bySlug: make(map[string]*models.Article),
		views:  make(map[uint]uint),
		nextID: 1,
	}
}

func (r *fakeArticleRepo) Create(_ context.Context, article *models.Article) error {
	if article.ID == 0 {
		article.ID = r.nextID
		r.nextID++
	}
	r.byID[article.ID] = article
	r.bySlug[article.Slug] = article
	r.views[article.ID] = article.Views
	return nil
}

func (r *fakeArticleRepo) Update(_ context.Context, article *models.Article) error {
	r.byID[article.ID] = article
	r.bySlug[article.Slug] = article
	return nil
}

func (r *fakeArticleRepo) Delete(_ context.Context, article *models.Article) error {
	delete(r.byID, article.ID)
	delete(r.bySlug, article.Slug)
	delete(r.views, article.ID)
	return nil
}

func (r *fakeArticleRepo) FindByID(_ context.Context, id uint) (*models.Article, error) {
	article, ok := r.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return article, nil
}

func (r *fakeArticleRepo) FindBySlug(_ context.Context, slug string) (*models.Article, error) {
	article, ok := r.bySlug[slug]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return article, nil
}

func (r *fakeArticleRepo) FindAll(_ context.Context, limit, offset int) ([]models.Article, error) {
	var items []models.Article
	for _, article := range r.byID {
		items = append(items, *article)
	}
	if offset >= len(items) {
		return []models.Article{}, nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], nil
}

func (r *fakeArticleRepo) GetViews(_ context.Context, id uint) (uint, error) {
	views, ok := r.views[id]
	if !ok {
		return 0, gorm.ErrRecordNotFound
	}
	return views, nil
}

func (r *fakeArticleRepo) IncrementViews(_ context.Context, id uint) (uint, error) {
	_, ok := r.views[id]
	if !ok {
		return 0, gorm.ErrRecordNotFound
	}
	r.views[id]++
	return r.views[id], nil
}

type fakeUserRepo struct {
	byID   map[uint]*models.User
	nextID uint
}

func (r *fakeUserRepo) Create(_ context.Context, user *models.User) error {
	if r.byID == nil {
		r.byID = make(map[uint]*models.User)
	}
	if user.ID == 0 {
		if r.nextID == 0 {
			r.nextID = 1
		}
		user.ID = r.nextID
		r.nextID++
	}
	r.byID[user.ID] = user
	return nil
}

func (r *fakeUserRepo) FindByEmail(_ context.Context, email string) (*models.User, error) {
	for _, u := range r.byID {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeUserRepo) FindByID(_ context.Context, id uint) (*models.User, error) {
	u, ok := r.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) FindByUsername(_ context.Context, username string) (*models.User, error) {
	for _, u := range r.byID {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeUserRepo) ExistsByEmailOrUsername(_ context.Context, email, username string) (bool, error) {
	for _, u := range r.byID {
		if u.Email == email || u.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func TestArticleService_CRUDAndViews(t *testing.T) {
	ctx := context.Background()
	repo := newFakeArticleRepo()
	userRepo := &fakeUserRepo{byID: map[uint]*models.User{1: {ID: 1, Role: models.UserRoleAdmin}}}
	svc := NewArticleService(repo, userRepo)

	article, err := svc.CreateArticle(ctx, "Hello World", 1)
	if err != nil {
		t.Fatalf("create article failed: %v", err)
	}
	if article.Slug == "" {
		t.Fatalf("expected slug to be set")
	}

	updated, err := svc.UpdateArticle(ctx, article.ID, "Updated", 1)
	if err != nil {
		t.Fatalf("update article failed: %v", err)
	}
	if updated.Title != "Updated" {
		t.Fatalf("expected updated title")
	}

	views, err := svc.GetArticleViews(ctx, article.ID)
	if err != nil {
		t.Fatalf("get views failed: %v", err)
	}
	if views != 0 {
		t.Fatalf("expected 0 views")
	}

	inc, err := svc.IncrementArticleViews(ctx, article.ID)
	if err != nil {
		t.Fatalf("increment views failed: %v", err)
	}
	if inc != 1 {
		t.Fatalf("expected 1 view")
	}

	list, err := svc.GetAllArticles(ctx, 10, 0)
	if err != nil || len(list) == 0 {
		t.Fatalf("expected articles list")
	}

	if err := svc.DeleteArticle(ctx, article.ID, 1); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = svc.GetArticle(ctx, article.ID)
	if !errors.Is(err, ErrArticleNotFound) {
		t.Fatalf("expected not found error")
	}
}
