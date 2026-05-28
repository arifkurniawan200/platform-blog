package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arifkurniawan200/platform-blog/pkg/middleware"
	"github.com/arifkurniawan200/platform-blog/services/article/internal/domain"
	"github.com/arifkurniawan200/platform-blog/services/article/internal/usecase"
	"go.uber.org/zap"
)

// ── Mock Repository ──

type mockArticleRepo struct {
	articles  map[string]*domain.Article
	comments  map[string]*domain.Comment
	claps     map[string]map[string]int // articleID -> userID -> count
	bookmarks map[string]map[string]bool // userID -> articleID -> bookmarked
	tags      map[string]*domain.Tag
}

func newMockRepo() *mockArticleRepo {
	return &mockArticleRepo{
		articles:  make(map[string]*domain.Article),
		comments:  make(map[string]*domain.Comment),
		claps:     make(map[string]map[string]int),
		bookmarks: make(map[string]map[string]bool),
		tags:      make(map[string]*domain.Tag),
	}
}

// Article CRUD
func (m *mockArticleRepo) Create(ctx context.Context, a *domain.Article) error { m.articles[a.ID] = a; return nil }
func (m *mockArticleRepo) FindBySlug(ctx context.Context, slug string) (*domain.Article, error) {
	for _, a := range m.articles {
		if a.Slug == slug {
			return a, nil
		}
	}
	return nil, errArticleNotFound
}
func (m *mockArticleRepo) FindByID(ctx context.Context, id string) (*domain.Article, error) {
	if a, ok := m.articles[id]; ok {
		return a, nil
	}
	return nil, errArticleNotFound
}
func (m *mockArticleRepo) Update(ctx context.Context, a *domain.Article) error { m.articles[a.ID] = a; return nil }
func (m *mockArticleRepo) Delete(ctx context.Context, id string) error          { delete(m.articles, id); return nil }
func (m *mockArticleRepo) List(ctx context.Context, limit, offset int) ([]*domain.Article, error) {
	var result []*domain.Article
	for _, a := range m.articles {
		result = append(result, a)
	}
	return result, nil
}
func (m *mockArticleRepo) ListByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*domain.Article, error) {
	return nil, nil
}
func (m *mockArticleRepo) FindOrCreateTag(ctx context.Context, name, slug string) (*domain.Tag, error) {
	if t, ok := m.tags[slug]; ok {
		return t, nil
	}
	t := &domain.Tag{ID: "tag-1", Name: name, Slug: slug}
	m.tags[slug] = t
	return t, nil
}
func (m *mockArticleRepo) AddArticleTags(ctx context.Context, articleID string, tagIDs []string) error { return nil }
func (m *mockArticleRepo) GetArticleTags(ctx context.Context, articleID string) ([]domain.Tag, error)   { return nil, nil }

// Comments
func (m *mockArticleRepo) CreateComment(ctx context.Context, c *domain.Comment) error {
	m.comments[c.ID] = c
	return nil
}
func (m *mockArticleRepo) ListCommentsByArticle(ctx context.Context, articleID string) ([]*domain.Comment, error) {
	var result []*domain.Comment
	for _, c := range m.comments {
		if c.ArticleID == articleID {
			result = append(result, c)
		}
	}
	return result, nil
}
func (m *mockArticleRepo) DeleteComment(ctx context.Context, id, userID string) error {
	if c, ok := m.comments[id]; ok && c.UserID == userID {
		delete(m.comments, id)
		return nil
	}
	return errCommentNotFound
}

// Claps
func (m *mockArticleRepo) AddClap(ctx context.Context, userID, articleID string, count int) (int, error) {
	if m.claps[articleID] == nil {
		m.claps[articleID] = make(map[string]int)
	}
	m.claps[articleID][userID] += count
	return m.claps[articleID][userID], nil
}
func (m *mockArticleRepo) GetClapCount(ctx context.Context, articleID string) (int, error) {
	total := 0
	for _, c := range m.claps[articleID] {
		total += c
	}
	return total, nil
}
func (m *mockArticleRepo) GetUserClapCount(ctx context.Context, userID, articleID string) (int, error) {
	if m.claps[articleID] == nil {
		return 0, nil
	}
	return m.claps[articleID][userID], nil
}

// Bookmarks
func (m *mockArticleRepo) AddBookmark(ctx context.Context, userID, articleID string) error {
	if m.bookmarks[userID] == nil {
		m.bookmarks[userID] = make(map[string]bool)
	}
	m.bookmarks[userID][articleID] = true
	return nil
}
func (m *mockArticleRepo) RemoveBookmark(ctx context.Context, userID, articleID string) error {
	if m.bookmarks[userID] != nil {
		delete(m.bookmarks[userID], articleID)
	}
	return nil
}
func (m *mockArticleRepo) IsBookmarked(ctx context.Context, userID, articleID string) (bool, error) {
	if m.bookmarks[userID] == nil {
		return false, nil
	}
	return m.bookmarks[userID][articleID], nil
}
func (m *mockArticleRepo) ListBookmarks(ctx context.Context, userID string, limit, offset int) ([]*domain.BookmarkInfo, error) {
	var result []*domain.BookmarkInfo
	for articleID := range m.bookmarks[userID] {
		a := m.articles[articleID]
		if a != nil {
			result = append(result, &domain.BookmarkInfo{
				ArticleID: a.ID,
				Title:     a.Title,
				Slug:      a.Slug,
			})
		}
	}
	return result, nil
}

// errors
var (
	errArticleNotFound  = &testError{"article not found"}
	errCommentNotFound  = &testError{"comment not found"}
)
type testError struct{ msg string }
func (e *testError) Error() string { return e.msg }

// helper
func newArticle(id, slug, title string) *domain.Article {
	return &domain.Article{
		ID:     id,
		Slug:   slug,
		Title:  title,
		AuthorID: "author-1",
		Content:  "<p>Test content for the article</p>",
		Status:   domain.StatusPublished,
	}
}

func newComment(id, articleID, userID, content string) *domain.Comment {
	return &domain.Comment{
		ID:        id,
		ArticleID: articleID,
		UserID:    userID,
		Content:   content,
	}
}

func setupHandler() (*ArticleHandler, *mockArticleRepo) {
	repo := newMockRepo()
	uc := usecase.NewArticleUsecase(repo)
	logger := zap.NewNop()
	return NewArticleHandler(uc, logger), repo
}

func addClaims(ctx context.Context, sub string) context.Context {
	return middleware.WithClaims(ctx, &middleware.JWTClaims{Sub: sub, Email: "test@test.com"})
}

// ── Comment Tests ──

func TestCreateComment(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article

	body := `{"content": "Great article!"}`
	req := httptest.NewRequest("POST", "/api/v1/articles/hello-world/comments", bytes.NewReader([]byte(body)))
	req = req.WithContext(addClaims(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.CreateComment(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(repo.comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(repo.comments))
	}
}

func TestCreateCommentValidation(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article

	body := `{"content": ""}`
	req := httptest.NewRequest("POST", "/api/v1/articles/hello-world/comments", bytes.NewReader([]byte(body)))
	req = req.WithContext(addClaims(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.CreateComment(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestListComments(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article
	repo.comments["c1"] = newComment("c1", "a1", "user-1", "Nice!")
	repo.comments["c2"] = newComment("c2", "a1", "user-2", "Thanks!")

	req := httptest.NewRequest("GET", "/api/v1/articles/hello-world/comments", nil)
	rec := httptest.NewRecorder()

	h.ListComments(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Data []domain.Comment `json:"data"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(resp.Data))
	}
}

func TestDeleteComment(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article
	repo.comments["c1"] = newComment("c1", "a1", "user-1", "My comment")

	req := httptest.NewRequest("DELETE", "/api/v1/articles/hello-world/comments/c1", nil)
	req = req.WithContext(addClaims(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.DeleteComment(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	if len(repo.comments) != 0 {
		t.Fatalf("expected 0 comments after delete, got %d", len(repo.comments))
	}
}

func TestDeleteCommentWrongUser(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article
	repo.comments["c1"] = newComment("c1", "a1", "user-1", "My comment")

	req := httptest.NewRequest("DELETE", "/api/v1/articles/hello-world/comments/c1", nil)
	req = req.WithContext(addClaims(req.Context(), "user-2")) // different user
	rec := httptest.NewRecorder()

	h.DeleteComment(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// ── Clap Tests ──

func TestClapArticle(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article

	body := `{"count": 5}`
	req := httptest.NewRequest("POST", "/api/v1/articles/hello-world/clap", bytes.NewReader([]byte(body)))
	req = req.WithContext(addClaims(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.Clap(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var info domain.ClapInfo
	json.NewDecoder(rec.Body).Decode(&info)
	if info.TotalClaps != 5 {
		t.Fatalf("expected 5 total claps, got %d", info.TotalClaps)
	}
	if info.UserClaps != 5 {
		t.Fatalf("expected 5 user claps, got %d", info.UserClaps)
	}
}

func TestClapMultipleUsers(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article

	// User 1 claps 5
	body1 := `{"count": 5}`
	req1 := httptest.NewRequest("POST", "/api/v1/articles/hello-world/clap", bytes.NewReader([]byte(body1)))
	req1 = req1.WithContext(addClaims(req1.Context(), "user-1"))
	rec1 := httptest.NewRecorder()
	h.Clap(rec1, req1)

	// User 2 claps 3
	body2 := `{"count": 3}`
	req2 := httptest.NewRequest("POST", "/api/v1/articles/hello-world/clap", bytes.NewReader([]byte(body2)))
	req2 = req2.WithContext(addClaims(req2.Context(), "user-2"))
	rec2 := httptest.NewRecorder()
	h.Clap(rec2, req2)

	var info domain.ClapInfo
	json.NewDecoder(rec2.Body).Decode(&info)
	if info.TotalClaps != 8 {
		t.Fatalf("expected 8 total claps, got %d", info.TotalClaps)
	}
}

func TestGetClapInfo(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article
	repo.claps["a1"] = map[string]int{"user-1": 10, "user-2": 5}

	req := httptest.NewRequest("GET", "/api/v1/articles/hello-world/clap", nil)
	req = req.WithContext(addClaims(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetClapInfo(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var info domain.ClapInfo
	json.NewDecoder(rec.Body).Decode(&info)
	if info.TotalClaps != 15 {
		t.Fatalf("expected 15 total claps, got %d", info.TotalClaps)
	}
	if info.UserClaps != 10 {
		t.Fatalf("expected 10 user claps, got %d", info.UserClaps)
	}
}

// ── Bookmark Tests ──

func TestBookmarkArticle(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article

	req := httptest.NewRequest("POST", "/api/v1/articles/hello-world/bookmark", nil)
	req = req.WithContext(addClaims(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.Bookmark(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	if !repo.bookmarks["user-1"]["a1"] {
		t.Fatal("expected article to be bookmarked")
	}
}

func TestUnbookmarkArticle(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article
	repo.bookmarks["user-1"] = map[string]bool{"a1": true}

	req := httptest.NewRequest("DELETE", "/api/v1/articles/hello-world/bookmark", nil)
	req = req.WithContext(addClaims(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.Unbookmark(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	if repo.bookmarks["user-1"]["a1"] {
		t.Fatal("expected article to be unbookmarked")
	}
}

func TestListBookmarks(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article
	repo.bookmarks["user-1"] = map[string]bool{"a1": true}

	req := httptest.NewRequest("GET", "/api/v1/bookmarks", nil)
	req = req.WithContext(addClaims(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.ListBookmarks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data []domain.BookmarkInfo `json:"data"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 bookmark, got %d", len(resp.Data))
	}
	if resp.Data[0].Title != "Hello World" {
		t.Fatalf("expected 'Hello World', got '%s'", resp.Data[0].Title)
	}
}

// ── Auth enforcement tests ──

func TestCreateCommentUnauthorized(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article

	body := `{"content": "test"}`
	req := httptest.NewRequest("POST", "/api/v1/articles/hello-world/comments", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()

	h.CreateComment(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestClapUnauthorized(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article

	body := `{"count": 5}`
	req := httptest.NewRequest("POST", "/api/v1/articles/hello-world/clap", bytes.NewReader([]byte(body)))
	rec := httptest.NewRecorder()

	h.Clap(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBookmarkUnauthorized(t *testing.T) {
	h, repo := setupHandler()
	article := newArticle("a1", "hello-world", "Hello World")
	repo.articles["a1"] = article

	req := httptest.NewRequest("POST", "/api/v1/articles/hello-world/bookmark", nil)
	rec := httptest.NewRecorder()

	h.Bookmark(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
