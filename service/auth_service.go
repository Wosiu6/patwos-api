package service

import (
	"context"
	"errors"
	"time"

	"github.com/Wosiu6/patwos-api/authcache"
	"github.com/Wosiu6/patwos-api/config"
	"github.com/Wosiu6/patwos-api/models"
	"github.com/Wosiu6/patwos-api/repository"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var (
	ErrUserAlreadyExists  = errors.New("user with this email or username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*models.User, string, error)
	Login(ctx context.Context, email, password string) (*models.User, string, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	Logout(ctx context.Context, token string, userID uint) error
	IsTokenRevoked(ctx context.Context, token string) bool
}

type authService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
	db       *gorm.DB
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config, db *gorm.DB) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
		db:       db,
	}
}

func (s *authService) Register(ctx context.Context, username, email, password string) (*models.User, string, error) {
	exists, err := s.userRepo.ExistsByEmailOrUsername(ctx, email, username)
	if err != nil {
		return nil, "", err
	}
	if exists {
		return nil, "", ErrUserAlreadyExists
	}

	user := &models.User{
		Username: username,
		Email:    email,
		State:    models.UserStatusActive,
		Role:     models.UserRoleUser,
	}

	if err := user.HashPassword(password); err != nil {
		return nil, "", err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.generateToken(user.ID, user.State, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if !user.CheckPassword(password) {
		return nil, "", ErrInvalidCredentials
	}

	if user.State != models.UserStatusActive {
		return nil, "", ErrUnauthorized
	}

	token, err := s.generateToken(user.ID, user.State, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *authService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *authService) Logout(ctx context.Context, token string, userID uint) error {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("invalid expiry claim")
	}

	revokedToken := &models.RevokedToken{
		Token:     token,
		UserID:    userID,
		RevokedAt: time.Now(),
		ExpiresAt: time.Unix(int64(exp), 0),
	}

	if err := s.db.WithContext(ctx).Create(revokedToken).Error; err != nil {
		return err
	}

	authcache.Add(token, revokedToken.ExpiresAt)
	return nil
}

func (s *authService) IsTokenRevoked(ctx context.Context, token string) bool {
	if authcache.IsRevoked(token) {
		return true
	}

	var count int64
	s.db.WithContext(ctx).Model(&models.RevokedToken{}).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		Count(&count)
	return count > 0
}

func (s *authService) generateToken(userID uint, userState models.UserState, userRole models.UserRole) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat":     time.Now().Unix(),
		"state":   userState,
		"role":    userRole,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
