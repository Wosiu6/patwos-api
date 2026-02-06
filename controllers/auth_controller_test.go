package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/service"
	"github.com/gin-gonic/gin"
)

type fakeAuthService struct {
	registerFn func(ctx context.Context, username, email, password string) (*models.User, string, error)
	loginFn    func(ctx context.Context, email, password string) (*models.User, string, error)
	logoutFn   func(ctx context.Context, token string, userID uint) error
}

func (f *fakeAuthService) Register(ctx context.Context, username, email, password string) (*models.User, string, error) {
	return f.registerFn(ctx, username, email, password)
}

func (f *fakeAuthService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	return f.loginFn(ctx, email, password)
}

func (f *fakeAuthService) GetUserByID(context.Context, uint) (*models.User, error) {
	return nil, nil
}

func (f *fakeAuthService) Logout(ctx context.Context, token string, userID uint) error {
	if f.logoutFn == nil {
		return nil
	}
	return f.logoutFn(ctx, token, userID)
}

func (f *fakeAuthService) IsTokenRevoked(context.Context, string) bool {
	return false
}

func TestAuthController_RegisterAndLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller := NewAuthController(&fakeAuthService{
		registerFn: func(context.Context, string, string, string) (*models.User, string, error) {
			return &models.User{ID: 1, Username: "user", Email: "user@example.com"}, "token", nil
		},
		loginFn: func(context.Context, string, string) (*models.User, string, error) {
			return nil, "", service.ErrInvalidCredentials
		},
	})

	r := gin.New()
	r.POST("/register", controller.Register)
	r.POST("/login", controller.Login)

	regBody, _ := json.Marshal(models.UserRegisterRequest{Username: "user", Email: "user@example.com", Password: "pass123"})
	regReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	r.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", regW.Code)
	}

	loginBody, _ := json.Marshal(models.UserLoginRequest{Email: "user@example.com", Password: "bad"})
	loginReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)
	if loginW.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", loginW.Code)
	}
}

func TestAuthController_LogoutMissingAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller := NewAuthController(&fakeAuthService{})
	r := gin.New()
	r.POST("/logout", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		controller.Logout(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
