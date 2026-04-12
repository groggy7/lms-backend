package domain

import (
	"context"
)

type Role string

const (
	RoleStudent    Role = "student"
	RoleInstructor Role = "instructor"
)

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	FullName string `json:"fullName"`
	Role     Role   `json:"role"`
}

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"fullName"`
}

type AuthResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"` // Mock JWT for now
}

type UserRepository interface {
	Create(ctx context.Context, user User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type AuthUsecase interface {
	Register(ctx context.Context, req AuthRequest) (*AuthResponse, error)
	Login(ctx context.Context, req AuthRequest) (*AuthResponse, error)
}
