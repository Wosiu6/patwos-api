package testing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wosiu6/patwos-api/config"
	"github.com/Wosiu6/patwos-api/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.Article{},
		&models.Comment{},
		&models.ArticleVote{},
		&models.CommentVote{},
		&models.RevokedToken{},
	)
	assert.NoError(t, err)

	return db
}

func CreateTestUser(t *testing.T, db *gorm.DB, username, email, password string, role models.UserRole) *models.User {
	user := &models.User{
		Username: username,
		Email:    email,
		State:    models.UserStatusActive,
		Role:     role,
	}
	err := user.HashPassword(password)
	assert.NoError(t, err)
	
	result := db.Create(user)
	assert.NoError(t, result.Error)
	
	return user
}

func CreateTestArticle(t *testing.T, db *gorm.DB, title string, authorID uint) *models.Article {
	article := &models.Article{
		Title:    title,
		Slug:     generateSlug(title),
		AuthorID: authorID,
	}
	
	result := db.Create(article)
	assert.NoError(t, result.Error)
	
	return article
}

func CreateTestComment(t *testing.T, db *gorm.DB, content string, articleID, userID uint) *models.Comment {
	comment := &models.Comment{
		Content:   content,
		ArticleID: articleID,
		UserID:    userID,
	}
	
	result := db.Create(comment)
	assert.NoError(t, result.Error)
	
	return comment
}

func GenerateTestJWT(userID uint, state models.UserState, role models.UserRole, secret string) string {
	claims := jwt.MapClaims{
		"user_id": float64(userID),
		"state":   float64(state),
		"role":    float64(role),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	
	return tokenString
}

func GenerateExpiredTestJWT(userID uint, state models.UserState, role models.UserRole, secret string) string {
	claims := jwt.MapClaims{
		"user_id": float64(userID),
		"state":   float64(state),
		"role":    float64(role),
		"exp":     time.Now().Add(-time.Hour).Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	
	return tokenString
}

func GetTestConfig() *config.Config {
	return &config.Config{
		JWTSecret:      "test-secret-key-for-testing-only",
		AllowedOrigins: []string{"http://localhost:3000"},
		GinMode:        "test",
		MaxRequestSize: 10485760,
	}
}

func MakeRequest(method, url string, body interface{}, token string) (*httptest.ResponseRecorder, *http.Request) {
	var req *http.Request
	
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	w := httptest.NewRecorder()
	return w, req
}

func ParseResponse(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), target)
	assert.NoError(t, err)
}

func AssertSecurityHeaders(t *testing.T, w *httptest.ResponseRecorder) {
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "geolocation=(), microphone=(), camera=()", w.Header().Get("Permissions-Policy"))
}

func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func generateSlug(title string) string {
	return title
}
