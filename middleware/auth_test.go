package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	testhelpers "github.com/Wosiu6/patwos-api/testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()

	router := gin.New()
	router.Use(AuthMiddleware(db, cfg))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()

	router := gin.New()
	router.Use(AuthMiddleware(db, cfg))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidToken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	expiredToken := testhelpers.GenerateExpiredTestJWT(user.ID, user.State, user.Role, cfg.JWTSecret)

	router := gin.New()
	router.Use(AuthMiddleware(db, cfg))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_RevokedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	token := testhelpers.GenerateTestJWT(user.ID, user.State, user.Role, cfg.JWTSecret)

	revokedToken := &models.RevokedToken{
		Token:  token,
		UserID: user.ID,
	}
	db.Create(revokedToken)

	router := gin.New()
	router.Use(AuthMiddleware(db, cfg))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "TOKEN_REVOKED")
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	token := testhelpers.GenerateTestJWT(user.ID, user.State, user.Role, cfg.JWTSecret)

	router := gin.New()
	router.Use(AuthMiddleware(db, cfg))
	router.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_InactiveUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	token := testhelpers.GenerateTestJWT(user.ID, models.UserStatusInactive, user.Role, cfg.JWTSecret)

	router := gin.New()
	router.Use(AuthMiddleware(db, cfg))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)

	token := testhelpers.GenerateTestJWT(user.ID, user.State, user.Role, "wrong-secret")

	router := gin.New()
	router.Use(AuthMiddleware(db, cfg))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_NonExistentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()

	token := testhelpers.GenerateTestJWT(999, models.UserStatusActive, models.UserRoleUser, cfg.JWTSecret)

	router := gin.New()
	router.Use(AuthMiddleware(db, cfg))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
