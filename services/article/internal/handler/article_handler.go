package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/arifkurniawan200/platform-blog/pkg/middleware"
	"github.com/arifkurniawan200/platform-blog/pkg/pagination"
	"github.com/arifkurniawan200/platform-blog/pkg/response"
	"github.com/arifkurniawan200/platform-blog/services/article/internal/domain"
	"github.com/arifkurniawan200/platform-blog/services/article/internal/repository"
	"github.com/arifkurniawan200/platform-blog/services/article/internal/usecase"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// ArticleHandler handles article HTTP endpoints
type ArticleHandler struct {
	uc    *usecase.ArticleUsecase
	log   *zap.Logger
	valid *validator.Validate
}

// NewArticleHandler creates a new article handler
func NewArticleHandler(uc *usecase.ArticleUsecase, log *zap.Logger) *ArticleHandler {
	return &ArticleHandler{uc: uc, log: log, valid: validator.New()}
}

// RegisterRoutes registers article routes
func (h *ArticleHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/articles", h.Create)
	mux.HandleFunc("GET /api/v1/articles", h.List)
	mux.HandleFunc("GET /api/v1/articles/{slug}", h.GetBySlug)
}

// Create handles POST /api/v1/articles
func (h *ArticleHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := middleware.GetClaims(ctx)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req domain.CreateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.valid.Struct(req); err != nil {
		response.WriteError(w, http.StatusUnprocessableEntity, "Validation failed: "+err.Error())
		return
	}

	article, err := h.uc.Create(ctx, claims.Sub, &req)
	if err != nil {
		h.log.Error("Create article failed", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to create article")
		return
	}

	response.WriteJSON(w, http.StatusCreated, article)
}

// GetBySlug handles GET /api/v1/articles/{slug}
func (h *ArticleHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	slug := r.PathValue("slug")
	if slug == "" {
		response.WriteError(w, http.StatusBadRequest, "Missing article slug")
		return
	}

	article, err := h.uc.GetBySlug(ctx, slug)
	if err != nil {
		if err == repository.ErrArticleNotFound {
			response.WriteError(w, http.StatusNotFound, "Article not found")
			return
		}
		h.log.Error("Get article failed", "error", err, "slug", slug)
		response.WriteError(w, http.StatusInternalServerError, "Failed to get article")
		return
	}

	response.WriteJSON(w, http.StatusOK, article)
}

// List handles GET /api/v1/articles
func (h *ArticleHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	limit, offset := pagination.ParseQueryParams(r)

	articles, err := h.uc.List(ctx, limit, offset)
	if err != nil {
		h.log.Error("List articles failed", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to list articles")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":   articles,
		"limit":  limit,
		"offset": offset,
	})
}
