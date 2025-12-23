package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wosiu6/patwos-api/config"
	"github.com/Wosiu6/patwos-api/controllers"
	"github.com/Wosiu6/patwos-api/middleware"
	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/repository"
	"github.com/Wosiu6/patwos-api/service"
	testhelpers "github.com/Wosiu6/patwos-api/testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupIntegrationTest(t *testing.T) (*gin.Engine, *config.Config) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()

	router := gin.New()
	router.Use(middleware.SecurityHeaders())

	userRepo := repository.NewUserRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	authService := service.NewAuthService(userRepo, cfg, db)
	articleService := service.NewArticleService(articleRepo, userRepo)

	authController := controllers.NewAuthController(authService)
	articleController := controllers.NewArticleController(articleService)

	public := router.Group("/api")
	{
		public.POST("/register", authController.Register)
		public.POST("/login", authController.Login)
		public.GET("/articles", articleController.GetArticles)
		public.GET("/articles/:id", articleController.GetArticle)
	}

	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(db, cfg))
	{
		protected.GET("/me", authController.GetCurrentUser)
		protected.POST("/logout", authController.Logout)
		protected.POST("/articles", articleController.CreateArticle)
		protected.PUT("/articles/:id", articleController.UpdateArticle)
		protected.DELETE("/articles/:id", articleController.DeleteArticle)
	}

	return router, cfg
}

func TestIntegration_UserRegistrationAndLogin(t *testing.T) {
	router, _ := setupIntegrationTest(t)

	registerBody := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var registerResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResponse)
	assert.NotEmpty(t, registerResponse["token"])

	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ = json.Marshal(loginBody)
	req = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NotEmpty(t, loginResponse["token"])
}

func TestIntegration_CreateAndRetrieveArticle(t *testing.T) {
	router, cfg := setupIntegrationTest(t)

	registerBody := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var registerResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResponse)
	token := registerResponse["token"].(string)

	createBody := map[string]string{
		"title": "Test Article",
	}
	body, _ = json.Marshal(createBody)
	req = httptest.NewRequest("POST", "/api/articles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	article := createResponse["article"].(map[string]interface{})
	articleID := article["id"].(float64)

	req = httptest.NewRequest("GET", "/api/articles/"+string(rune(int(articleID))), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	testhelpers.AssertSecurityHeaders(t, w)
}

func TestIntegration_UnauthorizedAccessToProtectedEndpoint(t *testing.T) {
	router, _ := setupIntegrationTest(t)

	req := httptest.NewRequest("GET", "/api/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_UpdateOwnArticle(t *testing.T) {
	router, _ := setupIntegrationTest(t)

	registerBody := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var registerResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResponse)
	token := registerResponse["token"].(string)

	createBody := map[string]string{
		"title": "Original Title",
	}
	body, _ = json.Marshal(createBody)
	req = httptest.NewRequest("POST", "/api/articles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	article := createResponse["article"].(map[string]interface{})
	articleID := article["id"].(float64)

	updateBody := map[string]string{
		"title": "Updated Title",
	}
	body, _ = json.Marshal(updateBody)
	req = httptest.NewRequest("PUT", "/api/articles/"+string(rune(int(articleID))), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIntegration_CannotUpdateOtherUsersArticle(t *testing.T) {
	router, _ := setupIntegrationTest(t)

	user1Body := map[string]string{
		"username": "user1",
		"email":    "user1@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(user1Body)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var user1Response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &user1Response)
	token1 := user1Response["token"].(string)

	createBody := map[string]string{
		"title": "User1 Article",
	}
	body, _ = json.Marshal(createBody)
	req = httptest.NewRequest("POST", "/api/articles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token1)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	article := createResponse["article"].(map[string]interface{})
	articleID := article["id"].(float64)

	user2Body := map[string]string{
		"username": "user2",
		"email":    "user2@example.com",
		"password": "password123",
	}
	body, _ = json.Marshal(user2Body)
	req = httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var user2Response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &user2Response)
	token2 := user2Response["token"].(string)

	updateBody := map[string]string{
		"title": "Hacked Title",
	}
	body, _ = json.Marshal(updateBody)
	req = httptest.NewRequest("PUT", "/api/articles/"+string(rune(int(articleID))), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token2)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestIntegration_LogoutInvalidatesToken(t *testing.T) {
	router, _ := setupIntegrationTest(t)

	registerBody := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var registerResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResponse)
	token := registerResponse["token"].(string)

	req = httptest.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest("POST", "/api/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_SecurityHeadersOnAllEndpoints(t *testing.T) {
	router, _ := setupIntegrationTest(t)

	endpoints := []struct {
		method string
		path   string
		body   map[string]string
	}{
		{"GET", "/api/articles", nil},
		{"POST", "/api/register", map[string]string{"username": "test", "email": "test@example.com", "password": "password123"}},
		{"POST", "/api/login", map[string]string{"email": "test@example.com", "password": "password123"}},
	}

	for _, endpoint := range endpoints {
		var req *http.Request
		if endpoint.body != nil {
			body, _ := json.Marshal(endpoint.body)
			req = httptest.NewRequest(endpoint.method, endpoint.path, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(endpoint.method, endpoint.path, nil)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertSecurityHeaders(t, w)
	}
}

func TestIntegration_InputValidation(t *testing.T) {
	router, _ := setupIntegrationTest(t)

	tests := []struct {
		name     string
		endpoint string
		body     map[string]string
		expected int
	}{
		{"Short Username", "/api/register", map[string]string{"username": "ab", "email": "test@example.com", "password": "password123"}, http.StatusBadRequest},
		{"Invalid Email", "/api/register", map[string]string{"username": "testuser", "email": "invalid", "password": "password123"}, http.StatusBadRequest},
		{"Short Password", "/api/register", map[string]string{"username": "testuser", "email": "test@example.com", "password": "short"}, http.StatusBadRequest},
		{"Missing Fields", "/api/register", map[string]string{"username": "testuser"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", tt.endpoint, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expected, w.Code)
		})
	}
}

func TestIntegration_SQLInjectionPrevention(t *testing.T) {
	router, _ := setupIntegrationTest(t)

	registerBody := map[string]string{
		"username": "admin'; DROP TABLE users;--",
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ = json.Marshal(loginBody)
	req = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIntegration_XSSPrevention(t *testing.T) {
	router, _ := setupIntegrationTest(t)

	registerBody := map[string]string{
		"username": "<script>alert('xss')</script>",
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	user := response["user"].(map[string]interface{})

	assert.Equal(t, "<script>alert('xss')</script>", user["username"])
	assert.Contains(t, w.Header().Get("Content-Type"), "json")
}
