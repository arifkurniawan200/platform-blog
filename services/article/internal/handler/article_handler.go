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
	withAuth := func(fn http.HandlerFunc) http.Handler {
		return middleware.JWTMiddleware(fn)
	}

	// Public routes (no auth required)
	mux.HandleFunc("GET /api/v1/articles", h.List)
	mux.HandleFunc("GET /api/v1/articles/{slug}", h.GetBySlug)
	mux.HandleFunc("GET /api/v1/articles/{slug}/comments", h.ListComments)
	mux.HandleFunc("GET /api/v1/articles/{slug}/clap", h.GetClapInfo)
	mux.HandleFunc("GET /api/v1/search", h.Search)
	mux.HandleFunc("GET /api/v1/users/{userID}/stats", h.GetUserStats)

	// Protected routes (auth required)
	mux.Handle("POST /api/v1/articles", withAuth(h.Create))
	mux.Handle("POST /api/v1/articles/{slug}/comments", withAuth(h.CreateComment))
	mux.Handle("DELETE /api/v1/articles/{slug}/comments/{id}", withAuth(h.DeleteComment))
	mux.Handle("POST /api/v1/articles/{slug}/clap", withAuth(h.Clap))
	mux.Handle("POST /api/v1/articles/{slug}/bookmark", withAuth(h.Bookmark))
	mux.Handle("DELETE /api/v1/articles/{slug}/bookmark", withAuth(h.Unbookmark))
	mux.Handle("GET /api/v1/bookmarks", withAuth(h.ListBookmarks))
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
		h.log.Error("Create article failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to create article")
		return
	}

	response.WriteJSON(w, http.StatusCreated, map[string]interface{}{"data": article})
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
		h.log.Error("Get article failed", zap.Error(err), zap.String("slug", slug))
		response.WriteError(w, http.StatusInternalServerError, "Failed to get article")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": article})
}

// List handles GET /api/v1/articles
func (h *ArticleHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	limit, offset := pagination.ParseQueryParams(r)

	articles, err := h.uc.List(ctx, limit, offset)
	if err != nil {
		h.log.Error("List articles failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to list articles")
		return
	}
	if articles == nil {
		articles = []*domain.Article{}
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
		h.log.Error("Create comment failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	// Fire email notification async (best-effort)
	go h.notifyAuthorOfComment(article.ID, claims.Sub, article.Title, req.Content)

	response.WriteJSON(w, http.StatusCreated, map[string]interface{}{"data": comment})
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
		h.log.Error("List comments failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to list comments")
		return
	}
	if comments == nil {
		comments = []*domain.Comment{}
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
		h.log.Error("Delete comment failed", zap.Error(err))
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
		h.log.Error("Clap failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to clap article")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": info})
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
		h.log.Error("Get clap info failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to get clap info")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": info})
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
		h.log.Error("Bookmark failed", zap.Error(err))
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
		h.log.Error("Unbookmark failed", zap.Error(err))
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
		h.log.Error("List bookmarks failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to list bookmarks")
		return
	}
	if bookmarks == nil {
		bookmarks = []*domain.BookmarkInfo{}
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data": bookmarks,
	})
}

// Search handles GET /api/v1/search?q=...
func (h *ArticleHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	q := r.URL.Query().Get("q")
	if q == "" {
		response.WriteError(w, http.StatusBadRequest, "Missing query parameter 'q'")
		return
	}

	limit, offset := pagination.ParseQueryParams(r)

	results, err := h.uc.Search(ctx, q, limit, offset)
	if err != nil {
		h.log.Error("Search failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Search failed")
		return
	}
	if results == nil {
		results = []*domain.ArticleSearchResult{}
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data": results,
	})
}

// GetUserStats handles GET /api/v1/users/{username}/stats
func (h *ArticleHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	uID := r.PathValue("userID")
	if uID == "" {
		response.WriteError(w, http.StatusBadRequest, "Missing user identifier")
		return
	}

	stats, err := h.uc.GetUserStats(ctx, uID)
	if err != nil {
		h.log.Error("Get stats failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to get stats")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data": stats,
	})
}

// notifyAuthorOfComment fires an async email notification when someone comments.
// Best-effort: failures are logged but never returned to the user.
func (h *ArticleHandler) notifyAuthorOfComment(articleID, commenterID, articleTitle, snippet string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	article, err := h.uc.GetBySlug(ctx, "") // dummy, we need FindByID
	_ = article
	if err != nil {
		return
	}

	// Notification logic: call auth service to check email_notify, then send via himalaya.
	// This is a stub for now — full implementation requires auth service HTTP call + himalaya CLI.
	h.log.Info("Comment notification would fire",
		zap.String("article_id", articleID),
		zap.String("commenter_id", commenterID),
		zap.String("title", articleTitle),
	)
}
