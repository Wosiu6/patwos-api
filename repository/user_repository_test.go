package repository

import (
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	testhelpers "github.com/Wosiu6/patwos-api/testing"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUserRepository_Create_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		State:    models.UserStatusActive,
		Role:     models.UserRoleUser,
	}

	err := repo.Create(user)

	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestUserRepository_FindByEmail_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewUserRepository(db)

	created := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	found, err := repo.FindByEmail("test@example.com")

	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Email, found.Email)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.FindByEmail("nonexistent@example.com")

	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestUserRepository_FindByID_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewUserRepository(db)

	created := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	found, err := repo.FindByID(created.ID)

	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.FindByID(999)

	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestUserRepository_FindByUsername_Success(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewUserRepository(db)

	created := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	found, err := repo.FindByUsername("testuser")

	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Username, found.Username)
}

func TestUserRepository_FindByUsername_NotFound(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.FindByUsername("nonexistent")

	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestUserRepository_ExistsByEmailOrUsername_Email(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewUserRepository(db)

	testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	exists, err := repo.ExistsByEmailOrUsername("test@example.com", "otheruser")

	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_ExistsByEmailOrUsername_Username(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewUserRepository(db)

	testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	exists, err := repo.ExistsByEmailOrUsername("other@example.com", "testuser")

	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_ExistsByEmailOrUsername_Neither(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewUserRepository(db)

	testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	exists, err := repo.ExistsByEmailOrUsername("other@example.com", "otheruser")

	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewUserRepository(db)

	user1 := &models.User{
		Username: "user1",
		Email:    "test@example.com",
		Password: "password",
		State:    models.UserStatusActive,
		Role:     models.UserRoleUser,
	}
	repo.Create(user1)

	user2 := &models.User{
		Username: "user2",
		Email:    "test@example.com",
		Password: "password",
		State:    models.UserStatusActive,
		Role:     models.UserRoleUser,
	}
	err := repo.Create(user2)

	assert.Error(t, err)
}

func TestUserRepository_Create_DuplicateUsername(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	repo := NewUserRepository(db)

	user1 := &models.User{
		Username: "testuser",
		Email:    "test1@example.com",
		Password: "password",
		State:    models.UserStatusActive,
		Role:     models.UserRoleUser,
	}
	repo.Create(user1)

	user2 := &models.User{
		Username: "testuser",
		Email:    "test2@example.com",
		Password: "password",
		State:    models.UserStatusActive,
		Role:     models.UserRoleUser,
	}
	err := repo.Create(user2)

	assert.Error(t, err)
}
