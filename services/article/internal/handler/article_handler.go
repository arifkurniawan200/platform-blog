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

	// Comments
	mux.HandleFunc("POST /api/v1/articles/{slug}/comments", h.CreateComment)
	mux.HandleFunc("GET /api/v1/articles/{slug}/comments", h.ListComments)
	mux.HandleFunc("DELETE /api/v1/articles/{slug}/comments/{id}", h.DeleteComment)

	// Claps
	mux.HandleFunc("POST /api/v1/articles/{slug}/clap", h.Clap)
	mux.HandleFunc("GET /api/v1/articles/{slug}/clap", h.GetClapInfo)

	// Bookmarks
	mux.HandleFunc("POST /api/v1/articles/{slug}/bookmark", h.Bookmark)
	mux.HandleFunc("DELETE /api/v1/articles/{slug}/bookmark", h.Unbookmark)
	mux.HandleFunc("GET /api/v1/bookmarks", h.ListBookmarks)
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

// ── Comments ──

// CreateComment handles POST /api/v1/articles/{slug}/comments
func (h *ArticleHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	claims, ok := middleware.GetClaims(ctx)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		response.WriteError(w, http.StatusBadRequest, "Missing article slug")
		return
	}

	var req domain.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.valid.Struct(req); err != nil {
		response.WriteError(w, http.StatusUnprocessableEntity, "Validation failed: "+err.Error())
		return
	}

	// Resolve slug to article ID
	article, err := h.uc.GetBySlug(ctx, slug)
	if err != nil {
		if err == repository.ErrArticleNotFound {
			response.WriteError(w, http.StatusNotFound, "Article not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to find article")
		return
	}

	comment, err := h.uc.CreateComment(ctx, claims.Sub, article.ID, &req)
	if err != nil {
		h.log.Error("Create comment failed", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	response.WriteJSON(w, http.StatusCreated, comment)
}

// ListComments handles GET /api/v1/articles/{slug}/comments
func (h *ArticleHandler) ListComments(w http.ResponseWriter, r *http.Request) {
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
		response.WriteError(w, http.StatusInternalServerError, "Failed to find article")
		return
	}

	comments, err := h.uc.ListComments(ctx, article.ID)
	if err != nil {
		h.log.Error("List comments failed", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to list comments")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data": comments,
	})
}

// DeleteComment handles DELETE /api/v1/articles/{slug}/comments/{id}
func (h *ArticleHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	claims, ok := middleware.GetClaims(ctx)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	commentID := r.PathValue("id")
	if commentID == "" {
		response.WriteError(w, http.StatusBadRequest, "Missing comment ID")
		return
	}

	if err := h.uc.DeleteComment(ctx, commentID, claims.Sub); err != nil {
		h.log.Error("Delete comment failed", "error", err)
		response.WriteError(w, http.StatusNotFound, "Comment not found or not authorized")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Claps ──

// Clap handles POST /api/v1/articles/{slug}/clap
func (h *ArticleHandler) Clap(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	claims, ok := middleware.GetClaims(ctx)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		response.WriteError(w, http.StatusBadRequest, "Missing article slug")
		return
	}

	var req domain.ClapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.valid.Struct(req); err != nil {
		response.WriteError(w, http.StatusUnprocessableEntity, "Validation failed: "+err.Error())
		return
	}

	article, err := h.uc.GetBySlug(ctx, slug)
	if err != nil {
		if err == repository.ErrArticleNotFound {
			response.WriteError(w, http.StatusNotFound, "Article not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to find article")
		return
	}

	info, err := h.uc.ClapArticle(ctx, claims.Sub, article.ID, &req)
	if err != nil {
		h.log.Error("Clap failed", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to clap article")
		return
	}

	response.WriteJSON(w, http.StatusOK, info)
}

// GetClapInfo handles GET /api/v1/articles/{slug}/clap
func (h *ArticleHandler) GetClapInfo(w http.ResponseWriter, r *http.Request) {
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
		response.WriteError(w, http.StatusInternalServerError, "Failed to find article")
		return
	}

	// Get user claims for user-specific clap count (optional – may be anonymous)
	userID := ""
	if claims, ok := middleware.GetClaims(ctx); ok {
		userID = claims.Sub
	}

	info, err := h.uc.GetClapInfo(ctx, userID, article.ID)
	if err != nil {
		h.log.Error("Get clap info failed", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to get clap info")
		return
	}

	response.WriteJSON(w, http.StatusOK, info)
}

// ── Bookmarks ──

// Bookmark handles POST /api/v1/articles/{slug}/bookmark
func (h *ArticleHandler) Bookmark(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	claims, ok := middleware.GetClaims(ctx)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

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
		response.WriteError(w, http.StatusInternalServerError, "Failed to find article")
		return
	}

	if err := h.uc.BookmarkArticle(ctx, claims.Sub, article.ID); err != nil {
		h.log.Error("Bookmark failed", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to bookmark article")
		return
	}

	response.WriteJSON(w, http.StatusCreated, map[string]string{"status": "bookmarked"})
}

// Unbookmark handles DELETE /api/v1/articles/{slug}/bookmark
func (h *ArticleHandler) Unbookmark(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	claims, ok := middleware.GetClaims(ctx)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

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
		response.WriteError(w, http.StatusInternalServerError, "Failed to find article")
		return
	}

	if err := h.uc.UnbookmarkArticle(ctx, claims.Sub, article.ID); err != nil {
		h.log.Error("Unbookmark failed", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to unbookmark article")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListBookmarks handles GET /api/v1/bookmarks
func (h *ArticleHandler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	claims, ok := middleware.GetClaims(ctx)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	limit, offset := pagination.ParseQueryParams(r)

	bookmarks, err := h.uc.ListBookmarks(ctx, claims.Sub, limit, offset)
	if err != nil {
		h.log.Error("List bookmarks failed", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to list bookmarks")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data": bookmarks,
	})
}
