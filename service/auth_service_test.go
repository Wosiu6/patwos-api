package service

import (
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/repository"
	testhelpers "github.com/Wosiu6/patwos-api/testing"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_Register_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	user, token, err := authService.Register("testuser", "test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, models.UserRoleUser, user.Role)
	assert.Equal(t, models.UserStatusActive, user.State)
	assert.NotEmpty(t, user.Password)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	authService.Register("user1", "test@example.com", "password123")
	_, _, err := authService.Register("user2", "test@example.com", "different")

	assert.Error(t, err)
	assert.Equal(t, ErrUserAlreadyExists, err)
}

func TestAuthService_Register_DuplicateUsername(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	authService.Register("testuser", "test1@example.com", "password123")
	_, _, err := authService.Register("testuser", "test2@example.com", "different")

	assert.Error(t, err)
	assert.Equal(t, ErrUserAlreadyExists, err)
}

func TestAuthService_Register_PasswordHashing(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	plainPassword := "password123"
	user, _, err := authService.Register("testuser", "test@example.com", plainPassword)

	assert.NoError(t, err)
	assert.NotEqual(t, plainPassword, user.Password)
	assert.True(t, user.CheckPassword(plainPassword))
}

func TestAuthService_Login_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	authService.Register("testuser", "test@example.com", "password123")

	user, token, err := authService.Login("test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	authService.Register("testuser", "test@example.com", "password123")

	_, _, err := authService.Login("test@example.com", "wrongpassword")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
}

func TestAuthService_Login_NonExistentUser(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	_, _, err := authService.Login("nonexistent@example.com", "password123")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)
	user.State = models.UserStatusInactive
	db.Save(user)

	_, _, err := authService.Login("test@example.com", "password123")

	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestAuthService_GetUserByID_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	created := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	user, err := authService.GetUserByID(created.ID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, created.ID, user.ID)
	assert.Equal(t, created.Email, user.Email)
}

func TestAuthService_GetUserByID_NotFound(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	_, err := authService.GetUserByID(999)

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestAuthService_Logout_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	user, token, _ := authService.Register("testuser", "test@example.com", "password123")

	err := authService.Logout(token, user.ID)

	assert.NoError(t, err)
	assert.True(t, authService.IsTokenRevoked(token))
}

func TestAuthService_IsTokenRevoked_ValidToken(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	_, token, _ := authService.Register("testuser", "test@example.com", "password123")

	assert.False(t, authService.IsTokenRevoked(token))
}

func TestAuthService_IsTokenRevoked_RevokedToken(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	user, token, _ := authService.Register("testuser", "test@example.com", "password123")
	authService.Logout(token, user.ID)

	assert.True(t, authService.IsTokenRevoked(token))
}

func TestAuthService_GenerateToken_ValidClaims(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	_, token, err := authService.Register("testuser", "test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, len(token), 50)
}

func TestAuthService_Register_WeakPassword(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	user, token, err := authService.Register("testuser", "test@example.com", "123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
}

func TestAuthService_Login_EmptyPassword(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	authService.Register("testuser", "test@example.com", "password123")

	_, _, err := authService.Login("test@example.com", "")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
}

func TestAuthService_Register_SpecialCharactersUsername(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	user, _, err := authService.Register("test<script>", "test@example.com", "password123")

	assert.NoError(t, err)
	assert.Equal(t, "test<script>", user.Username)
}

func TestAuthService_Register_SQLInjectionAttempt(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	username := "admin'; DROP TABLE users;--"
	user, _, err := authService.Register(username, "test@example.com", "password123")

	assert.NoError(t, err)
	assert.Equal(t, username, user.Username)

	var count int64
	db.Model(&models.User{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestAuthService_Login_CaseSensitiveEmail(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, cfg, db)

	authService.Register("testuser", "Test@Example.com", "password123")

	_, _, err := authService.Login("test@example.com", "password123")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
}
