package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arifkurniawan200/platform-blog/services/auth/internal/domain"
	"github.com/arifkurniawan200/platform-blog/services/auth/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrUsernameTaken      = errors.New("username already taken")
)

// AuthUsecase handles authentication business logic
type AuthUsecase struct {
	repo      repository.UserRepository
	jwtSecret []byte
}

// NewAuthUsecase creates a new auth usecase
func NewAuthUsecase(repo repository.UserRepository, jwtSecret string) *AuthUsecase {
	return &AuthUsecase{repo: repo, jwtSecret: []byte(jwtSecret)}
}

// Register creates a new user account
func (uc *AuthUsecase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Check existing
	if _, err := uc.repo.FindByEmail(ctx, req.Email); err == nil {
		return nil, ErrEmailTaken
	}
	if _, err := uc.repo.FindByUsername(ctx, req.Username); err == nil {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		DisplayName:  req.Username,
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	access, refresh, err := uc.generateTokens(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         user,
	}, nil
}

// Login authenticates a user
func (uc *AuthUsecase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := uc.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	access, refresh, err := uc.generateTokens(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         user,
	}, nil
}

func (uc *AuthUsecase) generateTokens(user *domain.User) (access, refresh string, err error) {
	now := time.Now()

	// Access token: 15 minutes
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  "user",
		"exp":   now.Add(15 * time.Minute).Unix(),
		"iat":   now.Unix(),
	})
	access, err = accessToken.SignedString(uc.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token: 7 days
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": now.Add(7 * 24 * time.Hour).Unix(),
		"iat": now.Unix(),
	})
	refresh, err = refreshToken.SignedString(uc.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("sign refresh token: %w", err)
	}

	return access, refresh, nil
}
