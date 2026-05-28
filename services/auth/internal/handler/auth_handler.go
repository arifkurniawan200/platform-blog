package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/arifkurniawan200/platform-blog/pkg/response"
	"github.com/arifkurniawan200/platform-blog/services/auth/internal/domain"
	"github.com/arifkurniawan200/platform-blog/services/auth/internal/usecase"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// AuthUsecaseInterface defines the auth business logic (for testability)
type AuthUsecaseInterface interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error)
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
			h.log.Error("Register failed", "error", err)
			response.WriteError(w, http.StatusInternalServerError, "Registration failed")
		}
		return
	}

	response.WriteJSON(w, http.StatusCreated, authResp)
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
		h.log.Error("Login failed", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "Login failed")
		return
	}

	response.WriteJSON(w, http.StatusOK, authResp)
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
