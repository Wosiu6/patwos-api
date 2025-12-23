package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/repository"
	"github.com/Wosiu6/patwos-api/service"
	testhelpers "github.com/Wosiu6/patwos-api/testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAuthController(t *testing.T) (*AuthController, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg, db)
	controller := NewAuthController(authService)

	router := gin.New()
	return controller, router
}

func TestAuthController_Register_Success(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/register", controller.Register)

	reqBody := models.UserRegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["user"])
	assert.NotEmpty(t, response["token"])
}

func TestAuthController_Register_InvalidJSON(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/register", controller.Register)

	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_Register_MissingFields(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/register", controller.Register)

	reqBody := map[string]string{
		"username": "testuser",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_Register_ShortUsername(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/register", controller.Register)

	reqBody := models.UserRegisterRequest{
		Username: "ab",
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_Register_ShortPassword(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/register", controller.Register)

	reqBody := models.UserRegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "short",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_Register_InvalidEmail(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/register", controller.Register)

	reqBody := models.UserRegisterRequest{
		Username: "testuser",
		Email:    "invalid-email",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_Register_DuplicateUser(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/register", controller.Register)

	reqBody := models.UserRegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req1 := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusConflict, w2.Code)
}

func TestAuthController_Login_Success(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/register", controller.Register)
	router.POST("/login", controller.Login)

	registerBody := models.UserRegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginBody := models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ = json.Marshal(loginBody)
	req = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["user"])
	assert.NotEmpty(t, response["token"])
}

func TestAuthController_Login_InvalidCredentials(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/register", controller.Register)
	router.POST("/login", controller.Login)

	registerBody := models.UserRegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginBody := models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	body, _ = json.Marshal(loginBody)
	req = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthController_Login_NonExistentUser(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/login", controller.Login)

	loginBody := models.UserLoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(loginBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthController_Login_InvalidJSON(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/login", controller.Login)

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_GetCurrentUser_Success(t *testing.T) {
	controller, router := setupAuthController(t)

	router.GET("/me", func(c *gin.Context) {
		user := models.User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     models.UserRoleUser,
		}
		c.Set("user", user)
		controller.GetCurrentUser(c)
	})

	req := httptest.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["user"])
}

func TestAuthController_GetCurrentUser_NoUser(t *testing.T) {
	controller, router := setupAuthController(t)
	router.GET("/me", controller.GetCurrentUser)

	req := httptest.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthController_Logout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testhelpers.SetupTestDB(t)
	cfg := testhelpers.GetTestConfig()
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg, db)
	controller := NewAuthController(authService)

	user := testhelpers.CreateTestUser(t, db, "testuser", "test@example.com", "password123", models.UserRoleUser)
	token := testhelpers.GenerateTestJWT(user.ID, user.State, user.Role, cfg.JWTSecret)

	router := gin.New()
	router.POST("/logout", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		controller.Logout(c)
	})

	req := httptest.NewRequest("POST", "/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_Logout_NoUserID(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/logout", controller.Logout)

	req := httptest.NewRequest("POST", "/logout", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthController_Register_XSSAttempt(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/register", controller.Register)

	reqBody := models.UserRegisterRequest{
		Username: "<script>alert('xss')</script>",
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAuthController_Register_LongUsername(t *testing.T) {
	controller, router := setupAuthController(t)
	router.POST("/register", controller.Register)

	reqBody := models.UserRegisterRequest{
		Username: string(make([]byte, 51)),
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
