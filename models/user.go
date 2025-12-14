package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	State     UserState      `gorm:"not null;default:0" json:"state"`
	Role      UserRole       `gorm:"not null;default:0" json:"role"`
	Username  string         `gorm:"uniqueIndex;not null" json:"username" binding:"required,min=3,max=50"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email" binding:"required,email"`
	Password  string         `gorm:"not null" json:"-"`
	Comments  []Comment      `gorm:"foreignKey:UserID" json:"comments,omitempty"`
}

type UserState int

const (
	UserStatusActive UserState = iota
	UserStatusInactive
	UserStatusDeleted
)

type UserRole int

const (
	UserRoleUser UserRole = iota
	UserRoleAdmin
)

type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) ToResponse() UserResponse {
	role := "user"
	if u.Role == UserRoleAdmin {
		role = "admin"
	}
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      role,
		CreatedAt: u.CreatedAt,
	}
}

func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}
