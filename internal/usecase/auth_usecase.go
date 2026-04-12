package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/serhatkilbas/lms-poc/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	userRepo domain.UserRepository
}

func NewAuthUsecase(repo domain.UserRepository) domain.AuthUsecase {
	return &authUsecase{userRepo: repo}
}

func (u *authUsecase) Register(ctx context.Context, req domain.AuthRequest) (*domain.AuthResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	user := domain.User{
		ID:       fmt.Sprintf("usr_%d", rand.Intn(1000000)),
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Role:     domain.RoleInstructor, // Default to instructor as requested
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := u.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

func (u *authUsecase) Login(ctx context.Context, req domain.AuthRequest) (*domain.AuthResponse, error) {
	user, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := u.generateJWT(*user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		User:  *user,
		Token: token,
	}, nil
}

func (u *authUsecase) generateJWT(user domain.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "lumina-secret-key-2026" // Fallback for POC
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	})

	return token.SignedString([]byte(secret))
}
