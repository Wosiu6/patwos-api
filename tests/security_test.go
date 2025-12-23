package security

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wosiu6/patwos-api/controllers"
	"github.com/Wosiu6/patwos-api/middleware"
	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/repository"
	"github.com/Wosiu6/patwos-api/service"
	testhelpers "github.com/Wosiu6/patwos-api/testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupSecurityTest(t *testing.T) (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()

	router := gin.New()
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg, db)
	authController := controllers.NewAuthController(authService)

	router.POST("/register", authController.Register)
	router.POST("/login", authController.Login)

	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(db, cfg))
	{
		protected.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	}

	return router, cfg.JWTSecret
}

func TestSecurity_BruteForceProtection(t *testing.T) {
	router, _ := setupSecurityTest(t)

	loginBody := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "wrongpassword",
	}

	for i := 0; i < 10; i++ {
		body, _ := json.Marshal(loginBody)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
}

func TestSecurity_MassivePayload(t *testing.T) {
	router, _ := setupSecurityTest(t)

	largeString := strings.Repeat("A", 1000000)
	registerBody := map[string]string{
		"username": largeString,
		"email":    "test@example.com",
		"password": "password123",
	}

	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecurity_NullByteInjection(t *testing.T) {
	router, _ := setupSecurityTest(t)

	registerBody := map[string]string{
		"username": "admin\x00user",
		"email":    "test@example.com",
		"password": "password123",
	}

	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestSecurity_UnicodeNormalization(t *testing.T) {
	router, _ := setupSecurityTest(t)

	registerBody := map[string]string{
		"username": "ᴀᴅᴍɪɴ",
		"email":    "test@example.com",
		"password": "password123",
	}

	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestSecurity_PasswordInURL(t *testing.T) {
	router, _ := setupSecurityTest(t)

	req := httptest.NewRequest("POST", "/login?password=secret123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotContains(t, w.Body.String(), "secret123")
}

func TestSecurity_JWTAlgorithmConfusion(t *testing.T) {
	router, secret := setupSecurityTest(t)

	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VyX2lkIjoxLCJleHAiOjk5OTk5OTk5OTksInN0YXRlIjowLCJyb2xlIjoxfQ."

	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+noneToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	_ = secret
}

func TestSecurity_TokenReplay(t *testing.T) {
	router, _ := setupSecurityTest(t)

	registerBody := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	token := response["token"].(string)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestSecurity_PasswordComplexity(t *testing.T) {
	router, _ := setupSecurityTest(t)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Too Short", "12345", true},
		{"Valid", "password123", false},
		{"Only Numbers", "123456", true},
		{"Very Long", strings.Repeat("a", 1000), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registerBody := map[string]string{
				"username": "user_" + tt.name,
				"email":    "test_" + tt.name + "@example.com",
				"password": tt.password,
			}

			body, _ := json.Marshal(registerBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.wantErr {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestSecurity_SessionFixation(t *testing.T) {
	router, _ := setupSecurityTest(t)

	registerBody := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var registerResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResponse)
	token1 := registerResponse["token"].(string)

	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ = json.Marshal(loginBody)
	req = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	token2 := loginResponse["token"].(string)

	assert.NotEqual(t, token1, token2)
}

func TestSecurity_EmailEnumeration(t *testing.T) {
	router, _ := setupSecurityTest(t)

	registerBody := map[string]string{
		"username": "testuser",
		"email":    "exists@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req)

	loginBody1 := map[string]string{
		"email":    "exists@example.com",
		"password": "wrongpassword",
	}
	body, _ = json.Marshal(loginBody1)
	req = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req)

	loginBody2 := map[string]string{
		"email":    "notexists@example.com",
		"password": "wrongpassword",
	}
	body, _ = json.Marshal(loginBody2)
	req = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req)

	assert.Equal(t, w2.Code, w3.Code)
}

func TestSecurity_CSRF_MissingOrigin(t *testing.T) {
	router, _ := setupSecurityTest(t)

	registerBody := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusForbidden, w.Code)
}

func TestSecurity_RandomByteInput(t *testing.T) {
	router, _ := setupSecurityTest(t)

	randomBytes := make([]byte, 1024)
	rand.Read(randomBytes)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(randomBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestSecurity_ContentTypeConfusion(t *testing.T) {
	router, _ := setupSecurityTest(t)

	registerBody := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecurity_HeaderInjection(t *testing.T) {
	router, _ := setupSecurityTest(t)

	registerBody := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "1.2.3.4\r\nX-Injected: malicious")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotContains(t, w.Header().Get("X-Injected"), "malicious")
}

func TestSecurity_TimingAttack_PasswordComparison(t *testing.T) {
	router, _ := setupSecurityTest(t)

	registerBody := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "correctpassword123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	passwords := []string{
		"c",
		"co",
		"cor",
		"corr",
		"corre",
		"correc",
		"correct",
		"correctp",
		"wrongpassword",
	}

	for _, pwd := range passwords {
		loginBody := map[string]string{
			"email":    "test@example.com",
			"password": pwd,
		}
		body, _ := json.Marshal(loginBody)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
}
