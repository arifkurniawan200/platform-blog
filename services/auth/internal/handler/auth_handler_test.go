package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arifkurniawan200/platform-blog/services/auth/internal/domain"
	"github.com/arifkurniawan200/platform-blog/services/auth/internal/usecase"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap/zaptest"
)

// stubUsecase implements AuthUsecaseInterface for testing
type stubUsecase struct {
	registerFn func(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error)
	loginFn    func(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error)
}

func (s *stubUsecase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	if s.registerFn != nil {
		return s.registerFn(ctx, req)
	}
	return nil, nil
}
func (s *stubUsecase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	if s.loginFn != nil {
		return s.loginFn(ctx, req)
	}
	return nil, nil
}

func TestRegister_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		body any
		want int
	}{
		{"missing username", domain.RegisterRequest{Email: "a@b.com", Password: "pwd12345"}, 422},
		{"short username", domain.RegisterRequest{Username: "ab", Email: "a@b.com", Password: "pwd12345"}, 422},
		{"empty email", domain.RegisterRequest{Username: "johndoe", Password: "pwd12345"}, 422},
		{"invalid email", domain.RegisterRequest{Username: "johndoe", Email: "not-email", Password: "pwd12345"}, 422},
		{"short password", domain.RegisterRequest{Username: "johndoe", Email: "a@b.com", Password: "short"}, 422},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &stubUsecase{
				registerFn: func(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
					return nil, nil
				},
			}
			h := &AuthHandler{uc: uc, log: zaptest.NewLogger(nil), valid: validator.New()}

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Register(rec, req)

			if rec.Code != tt.want {
				t.Errorf("expected %d, got %d", tt.want, rec.Code)
			}
		})
	}
}

func TestRegister_Success(t *testing.T) {
	uc := &stubUsecase{
		registerFn: func(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
			return &domain.AuthResponse{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				User:         &domain.User{ID: "1", Username: req.Username},
			}, nil
		},
	}
	h := &AuthHandler{uc: uc, log: zaptest.NewLogger(nil), valid: validator.New()}

	body := domain.RegisterRequest{Username: "johndoe", Email: "john@example.com", Password: "password123"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var resp domain.AuthResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.AccessToken == "" {
		t.Error("expected access_token in response")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	uc := &stubUsecase{
		registerFn: func(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
			return nil, usecase.ErrEmailTaken
		},
	}
	h := &AuthHandler{uc: uc, log: zaptest.NewLogger(nil), valid: validator.New()}

	body := domain.RegisterRequest{Username: "johndoe", Email: "john@example.com", Password: "password123"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %d", rec.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	uc := &stubUsecase{
		loginFn: func(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
			return &domain.AuthResponse{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				User:         &domain.User{ID: "1", Email: req.Email},
			}, nil
		},
	}
	h := &AuthHandler{uc: uc, log: zaptest.NewLogger(nil), valid: validator.New()}

	body := domain.LoginRequest{Email: "john@example.com", Password: "password123"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestLogin_BadRequest(t *testing.T) {
	h := &AuthHandler{uc: &stubUsecase{}, log: zaptest.NewLogger(nil), valid: validator.New()}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("bad json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	uc := &stubUsecase{
		loginFn: func(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
			return nil, usecase.ErrInvalidCredentials
		},
	}
	h := &AuthHandler{uc: uc, log: zaptest.NewLogger(nil), valid: validator.New()}

	body := domain.LoginRequest{Email: "john@example.com", Password: "wrongpass"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
