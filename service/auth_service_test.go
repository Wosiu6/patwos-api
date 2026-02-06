package service

import (
	"context"
	"testing"

	"github.com/Wosiu6/patwos-api/config"
	"github.com/Wosiu6/patwos-api/models"
)

func TestAuthService_RegisterAndLogin(t *testing.T) {
	ctx := context.Background()
	userRepo := &fakeUserRepo{byID: map[uint]*models.User{}}
	cfg := &config.Config{JWTSecret: "secret"}
	svc := NewAuthService(userRepo, cfg, nil)

	user, token, err := svc.Register(ctx, "user", "user@example.com", "pass1234")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if token == "" || user.ID == 0 {
		t.Fatalf("expected token and user id")
	}

	_, _, err = svc.Register(ctx, "user", "user@example.com", "pass1234")
	if err != ErrUserAlreadyExists {
		t.Fatalf("expected user exists error")
	}

	loggedIn, _, err := svc.Login(ctx, "user@example.com", "pass1234")
	if err != nil || loggedIn.Email != "user@example.com" {
		t.Fatalf("login failed")
	}

	_, _, err = svc.Login(ctx, "user@example.com", "bad")
	if err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials")
	}
}
