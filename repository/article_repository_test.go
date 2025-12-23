package repository

import (
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	testhelpers "github.com/Wosiu6/patwos-api/testing"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestArticleRepository_Create_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewArticleRepository(db)
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	article := &models.Article{
		Title:    "Test Article",
		Slug:     "test-article",
		AuthorID: user.ID,
	}

	err := repo.Create(article)

	assert.NoError(t, err)
	assert.NotZero(t, article.ID)
}

func TestArticleRepository_FindByID_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewArticleRepository(db)
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	created := testhelpers.CreateTestArticle(t, db, "Test Article", user.ID)

	found, err := repo.FindByID(created.ID)

	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Title, found.Title)
	assert.NotNil(t, found.Author)
}

func TestArticleRepository_FindByID_NotFound(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewArticleRepository(db)

	_, err := repo.FindByID(999)

	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestArticleRepository_FindBySlug_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewArticleRepository(db)
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	created := testhelpers.CreateTestArticle(t, db, "Test Article", user.ID)

	found, err := repo.FindBySlug(created.Slug)

	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestArticleRepository_FindBySlug_NotFound(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewArticleRepository(db)

	_, err := repo.FindBySlug("nonexistent-slug")

	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestArticleRepository_Update_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewArticleRepository(db)
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	article := testhelpers.CreateTestArticle(t, db, "Original Title", user.ID)
	article.Title = "Updated Title"

	err := repo.Update(article)

	assert.NoError(t, err)

	found, _ := repo.FindByID(article.ID)
	assert.Equal(t, "Updated Title", found.Title)
}

func TestArticleRepository_Delete_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewArticleRepository(db)
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	article := testhelpers.CreateTestArticle(t, db, "Test Article", user.ID)

	err := repo.Delete(article)

	assert.NoError(t, err)

	_, err = repo.FindByID(article.ID)
	assert.Error(t, err)
}

func TestArticleRepository_FindAll_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewArticleRepository(db)
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	for i := 0; i < 5; i++ {
		testhelpers.CreateTestArticle(t, db, "Article "+string(rune('A'+i)), user.ID)
	}

	articles, err := repo.FindAll(10, 0)

	assert.NoError(t, err)
	assert.Len(t, articles, 5)
}

func TestArticleRepository_FindAll_Pagination(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewArticleRepository(db)
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	for i := 0; i < 10; i++ {
		testhelpers.CreateTestArticle(t, db, "Article "+string(rune('A'+i)), user.ID)
	}

	page1, err := repo.FindAll(5, 0)
	assert.NoError(t, err)
	assert.Len(t, page1, 5)

	page2, err := repo.FindAll(5, 5)
	assert.NoError(t, err)
	assert.Len(t, page2, 5)

	assert.NotEqual(t, page1[0].ID, page2[0].ID)
}

func TestArticleRepository_FindAll_Empty(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewArticleRepository(db)

	articles, err := repo.FindAll(10, 0)

	assert.NoError(t, err)
	assert.Empty(t, articles)
}

func TestArticleRepository_Create_DuplicateSlug(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewArticleRepository(db)
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	article1 := &models.Article{
		Title:    "Test Article",
		Slug:     "test-article",
		AuthorID: user.ID,
	}
	repo.Create(article1)

	article2 := &models.Article{
		Title:    "Test Article 2",
		Slug:     "test-article",
		AuthorID: user.ID,
	}
	err := repo.Create(article2)

	assert.Error(t, err)
}

func TestArticleRepository_FindAll_OrderedByCreatedAt(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewArticleRepository(db)
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	first := testhelpers.CreateTestArticle(t, db, "First Article", user.ID)
	second := testhelpers.CreateTestArticle(t, db, "Second Article", user.ID)

	articles, err := repo.FindAll(10, 0)

	assert.NoError(t, err)
	assert.Equal(t, second.ID, articles[0].ID)
	assert.Equal(t, first.ID, articles[1].ID)
}
