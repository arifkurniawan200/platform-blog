package domain

import "time"

// Article status constants
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)

// Article represents a blog article
type Article struct {
	ID          string    `json:"id"`
	AuthorID    string    `json:"author_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Subtitle    string    `json:"subtitle,omitempty"`
	Content     string    `json:"content"`
	CoverImage  string    `json:"cover_image,omitempty"`
	ReadingTime int       `json:"reading_time"`
	Status      string    `json:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	ViewCount   int       `json:"view_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// Relations (populated in list/get)
	Tags     []Tag       `json:"tags,omitempty"`
	Author   *AuthorInfo `json:"author,omitempty"`
	ClapCount int        `json:"clap_count,omitempty"`
}

// AuthorInfo is minimal author data embedded in article
type AuthorInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// Tag represents a content tag
type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Comment represents a comment on an article
type Comment struct {
	ID        string    `json:"id"`
	ArticleID string    `json:"article_id"`
	UserID    string    `json:"user_id"`
	ParentID  *string   `json:"parent_id,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateArticleRequest is the payload for creating an article
type CreateArticleRequest struct {
	Title      string   `json:"title" validate:"required,min=3,max=200"`
	Subtitle   string   `json:"subtitle" validate:"max=300"`
	Content    string   `json:"content" validate:"required,min=10"`
	CoverImage string   `json:"cover_image"`
	Tags       []string `json:"tags"`
}

// UpdateArticleRequest is the payload for updating an article
type UpdateArticleRequest struct {
	Title      string   `json:"title" validate:"omitempty,min=3,max=200"`
	Subtitle   string   `json:"subtitle" validate:"max=300"`
	Content    string   `json:"content" validate:"omitempty,min=10"`
	CoverImage string   `json:"cover_image"`
	Tags       []string `json:"tags"`
}

// ── Comments ──

// CreateCommentRequest is the payload for creating a comment
type CreateCommentRequest struct {
	Content  string  `json:"content" validate:"required,min=1,max=2000"`
	ParentID *string `json:"parent_id,omitempty"`
}

// CommentWithUser is a comment enriched with user info
type CommentWithUser struct {
	Comment
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// ── Claps ──

// ClapRequest is the payload for clapping an article
type ClapRequest struct {
	Count int `json:"count" validate:"required,min=1,max=50"`
}

// ClapInfo holds per-article clap data
type ClapInfo struct {
	ArticleID   string `json:"article_id"`
	TotalClaps  int    `json:"total_claps"`
	UserClaps   int    `json:"user_claps"`
}

// ── Bookmarks ──

// BookmarkInfo holds a single bookmark entry
type BookmarkInfo struct {
	ArticleID string `json:"article_id"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
}
