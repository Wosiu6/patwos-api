package service

import (
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/repository"
	testhelpers "github.com/Wosiu6/patwos-api/testing"
	"github.com/stretchr/testify/assert"
)

func TestArticleService_CreateArticle_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	article, err := articleService.CreateArticle("Test Article Title", user.ID)

	assert.NoError(t, err)
	assert.NotNil(t, article)
	assert.Equal(t, "Test Article Title", article.Title)
	assert.NotEmpty(t, article.Slug)
	assert.Equal(t, user.ID, article.AuthorID)
}

func TestArticleService_CreateArticle_GeneratesSlug(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	article, err := articleService.CreateArticle("My Great Article", user.ID)

	assert.NoError(t, err)
	assert.Equal(t, "my-great-article", article.Slug)
}

func TestArticleService_CreateArticle_DuplicateSlug(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	article1, err := articleService.CreateArticle("Test Article", user.ID)
	assert.NoError(t, err)

	article2, err := articleService.CreateArticle("Test Article", user.ID)
	assert.NoError(t, err)
	assert.NotEqual(t, article1.Slug, article2.Slug)
}

func TestArticleService_UpdateArticle_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)
	article := testhelpers.CreateTestArticle(t, db, "Original Title", user.ID)

	updated, err := articleService.UpdateArticle(article.ID, "Updated Title", user.ID)

	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
}

func TestArticleService_UpdateArticle_NotAuthor(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	author := testhelpers.CreateTestUser(t, db, "author", "author@example.com", "password123", models.UserRoleUser)
	otherUser := testhelpers.CreateTestUser(t, db, "other", "other@example.com", "password123", models.UserRoleUser)
	article := testhelpers.CreateTestArticle(t, db, "Original Title", author.ID)

	_, err := articleService.UpdateArticle(article.ID, "Updated Title", otherUser.ID)

	assert.Error(t, err)
	assert.Equal(t, ErrForbidden, err)
}

func TestArticleService_UpdateArticle_AdminCanUpdate(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	author := testhelpers.CreateTestUser(t, db, "author", "author@example.com", "password123", models.UserRoleUser)
	admin := testhelpers.CreateTestUser(t, db, "admin", "admin@example.com", "password123", models.UserRoleAdmin)
	article := testhelpers.CreateTestArticle(t, db, "Original Title", author.ID)

	updated, err := articleService.UpdateArticle(article.ID, "Admin Updated", admin.ID)

	assert.NoError(t, err)
	assert.Equal(t, "Admin Updated", updated.Title)
}

func TestArticleService_UpdateArticle_NotFound(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	_, err := articleService.UpdateArticle(999, "Updated Title", user.ID)

	assert.Error(t, err)
	assert.Equal(t, ErrArticleNotFound, err)
}

func TestArticleService_DeleteArticle_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)
	article := testhelpers.CreateTestArticle(t, db, "Test Article", user.ID)

	err := articleService.DeleteArticle(article.ID, user.ID)

	assert.NoError(t, err)

	_, err = articleService.GetArticle(article.ID)
	assert.Error(t, err)
}

func TestArticleService_DeleteArticle_NotAuthor(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	author := testhelpers.CreateTestUser(t, db, "author", "author@example.com", "password123", models.UserRoleUser)
	otherUser := testhelpers.CreateTestUser(t, db, "other", "other@example.com", "password123", models.UserRoleUser)
	article := testhelpers.CreateTestArticle(t, db, "Test Article", author.ID)

	err := articleService.DeleteArticle(article.ID, otherUser.ID)

	assert.Error(t, err)
	assert.Equal(t, ErrForbidden, err)
}

func TestArticleService_DeleteArticle_AdminCanDelete(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	author := testhelpers.CreateTestUser(t, db, "author", "author@example.com", "password123", models.UserRoleUser)
	admin := testhelpers.CreateTestUser(t, db, "admin", "admin@example.com", "password123", models.UserRoleAdmin)
	article := testhelpers.CreateTestArticle(t, db, "Test Article", author.ID)

	err := articleService.DeleteArticle(article.ID, admin.ID)

	assert.NoError(t, err)
}

func TestArticleService_GetArticle_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)
	created := testhelpers.CreateTestArticle(t, db, "Test Article", user.ID)

	article, err := articleService.GetArticle(created.ID)

	assert.NoError(t, err)
	assert.NotNil(t, article)
	assert.Equal(t, created.ID, article.ID)
}

func TestArticleService_GetArticle_NotFound(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	_, err := articleService.GetArticle(999)

	assert.Error(t, err)
	assert.Equal(t, ErrArticleNotFound, err)
}

func TestArticleService_GetArticleBySlug_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)
	created, _ := articleService.CreateArticle("Test Article", user.ID)

	article, err := articleService.GetArticleBySlug(created.Slug)

	assert.NoError(t, err)
	assert.NotNil(t, article)
	assert.Equal(t, created.ID, article.ID)
}

func TestArticleService_GetAllArticles_Pagination(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)
	
	for i := 0; i < 5; i++ {
		testhelpers.CreateTestArticle(t, db, "Article "+string(rune('A'+i)), user.ID)
	}

	articles, err := articleService.GetAllArticles(3, 0)

	assert.NoError(t, err)
	assert.Len(t, articles, 3)
}

func TestArticleService_CreateArticle_SpecialCharacters(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	article, err := articleService.CreateArticle("Test & Article <script>", user.ID)

	assert.NoError(t, err)
	assert.Equal(t, "Test & Article <script>", article.Title)
	assert.NotContains(t, article.Slug, "<")
	assert.NotContains(t, article.Slug, ">")
}

func TestArticleService_UpdateArticle_EmptyTitle(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	articleRepo := repository.NewArticleRepository(db)
	userRepo := repository.NewUserRepository(db)
	articleService := NewArticleService(articleRepo, userRepo)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)
	article := testhelpers.CreateTestArticle(t, db, "Original Title", user.ID)

	updated, err := articleService.UpdateArticle(article.ID, "", user.ID)

	assert.NoError(t, err)
	assert.Equal(t, "Original Title", updated.Title)
}
