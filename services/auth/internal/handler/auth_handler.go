package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/arifkurniawan200/platform-blog/pkg/middleware"
	"github.com/arifkurniawan200/platform-blog/pkg/response"
	"github.com/arifkurniawan200/platform-blog/services/auth/internal/domain"
	"github.com/arifkurniawan200/platform-blog/services/auth/internal/repository"
	"github.com/arifkurniawan200/platform-blog/services/auth/internal/usecase"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// AuthUsecaseInterface defines the auth business logic (for testability)
type AuthUsecaseInterface interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error)
	GetUserRepo() repository.UserRepository
}

// AuthHandler handles auth HTTP endpoints
type AuthHandler struct {
	uc    AuthUsecaseInterface
	log   *zap.Logger
	valid *validator.Validate
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(uc AuthUsecaseInterface, log *zap.Logger) *AuthHandler {
	return &AuthHandler{uc: uc, log: log, valid: validator.New()}
}

// RegisterRoutes registers auth routes
func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("GET /api/v1/users/{username}", h.GetProfile)
	mux.HandleFunc("PATCH /api/v1/users/me", h.UpdateProfile)
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.valid.Struct(req); err != nil {
		response.WriteError(w, http.StatusUnprocessableEntity, formatValidationError(err))
		return
	}

	authResp, err := h.uc.Register(ctx, &req)
	if err != nil {
		switch err {
		case usecase.ErrEmailTaken:
			response.WriteError(w, http.StatusConflict, "Email already registered")
		case usecase.ErrUsernameTaken:
			response.WriteError(w, http.StatusConflict, "Username already taken")
		default:
			h.log.Error("Register failed", zap.Error(err))
			response.WriteError(w, http.StatusInternalServerError, "Registration failed")
		}
		return
	}

	response.WriteJSON(w, http.StatusCreated, map[string]interface{}{"data": authResp})
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.valid.Struct(req); err != nil {
		response.WriteError(w, http.StatusUnprocessableEntity, formatValidationError(err))
		return
	}

	authResp, err := h.uc.Login(ctx, &req)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			response.WriteError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		h.log.Error("Login failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Login failed")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": authResp})
}

func formatValidationError(err error) string {
	var fieldErrors []string
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			fieldErrors = append(fieldErrors, fmt.Sprintf("field %s failed on %s", fe.Field(), fe.Tag()))
		}
		return strings.Join(fieldErrors, "; ")
	}
	return err.Error()
}

// GetProfile handles GET /api/v1/users/{username}
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	username := r.PathValue("username")
	if username == "" {
		response.WriteError(w, http.StatusBadRequest, "Missing username")
		return
	}

	repo := h.uc.GetUserRepo()
	user, err := repo.FindByUsername(ctx, username)
	if err != nil {
		if err == repository.ErrUserNotFound {
			response.WriteError(w, http.StatusNotFound, "User not found")
			return
		}
		h.log.Error("Get profile failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to get profile")
		return
	}

	articleCount, _ := repo.GetArticleCount(ctx, user.ID)

	resp := domain.ProfileResponse{
		User:         *user,
		ArticleCount: articleCount,
	}
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": resp})
}

// UpdateProfile handles PATCH /api/v1/users/me
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	claims, ok := middleware.GetClaims(ctx)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	repo := h.uc.GetUserRepo()
	user, err := repo.FindByID(ctx, claims.Sub)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	var req domain.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.valid.Struct(req); err != nil {
		response.WriteError(w, http.StatusUnprocessableEntity, formatValidationError(err))
		return
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}
	if req.EmailNotify != nil {
		user.EmailNotify = *req.EmailNotify
	}

	if err := repo.Update(ctx, user); err != nil {
		h.log.Error("Update profile failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": user})
}
