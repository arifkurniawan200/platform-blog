package repository

import (
	"context"
	"errors"

	"github.com/arifkurniawan200/platform-blog/services/article/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrArticleNotFound = errors.New("article not found")
)

// ArticleRepository defines the data access interface
type ArticleRepository interface {
	Create(ctx context.Context, article *domain.Article) error
	FindBySlug(ctx context.Context, slug string) (*domain.Article, error)
	FindByID(ctx context.Context, id string) (*domain.Article, error)
	Update(ctx context.Context, article *domain.Article) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*domain.Article, error)
	ListByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*domain.Article, error)
	// Tags
	FindOrCreateTag(ctx context.Context, name, slug string) (*domain.Tag, error)
	AddArticleTags(ctx context.Context, articleID string, tagIDs []string) error
	GetArticleTags(ctx context.Context, articleID string) ([]domain.Tag, error)
	// Comments
	CreateComment(ctx context.Context, comment *domain.Comment) error
	ListCommentsByArticle(ctx context.Context, articleID string) ([]*domain.Comment, error)
	DeleteComment(ctx context.Context, id, userID string) error
	// Claps
	AddClap(ctx context.Context, userID, articleID string, count int) (int, error)
	GetClapCount(ctx context.Context, articleID string) (int, error)
	GetUserClapCount(ctx context.Context, userID, articleID string) (int, error)
	// Bookmarks
	AddBookmark(ctx context.Context, userID, articleID string) error
	RemoveBookmark(ctx context.Context, userID, articleID string) error
	IsBookmarked(ctx context.Context, userID, articleID string) (bool, error)
	ListBookmarks(ctx context.Context, userID string, limit, offset int) ([]*domain.BookmarkInfo, error)
	// Search
	Search(ctx context.Context, query string, limit, offset int) ([]*domain.ArticleSearchResult, error)
	// Stats
	GetUserStats(ctx context.Context, userID string) (*domain.UserStats, error)
}

type pgArticleRepo struct {
	pool *pgxpool.Pool
}

func NewArticleRepository(pool *pgxpool.Pool) ArticleRepository {
	return &pgArticleRepo{pool: pool}
}

func (r *pgArticleRepo) Create(ctx context.Context, a *domain.Article) error {
	var publishedAt interface{}
	if a.PublishedAt != nil {
		publishedAt = *a.PublishedAt
	}
	return r.pool.QueryRow(ctx,
		`INSERT INTO articles (author_id, title, slug, subtitle, content, cover_image, reading_time, status, published_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at, updated_at`,
		a.AuthorID, a.Title, a.Slug, a.Subtitle, a.Content,
		a.CoverImage, a.ReadingTime, a.Status, publishedAt,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (r *pgArticleRepo) FindBySlug(ctx context.Context, slug string) (*domain.Article, error) {
	a := &domain.Article{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, author_id, title, slug, COALESCE(subtitle, ''), content, COALESCE(cover_image, ''),
		        reading_time, status, published_at, view_count, created_at, updated_at
		 FROM articles WHERE slug = $1`, slug,
	).Scan(&a.ID, &a.AuthorID, &a.Title, &a.Slug, &a.Subtitle, &a.Content,
		&a.CoverImage, &a.ReadingTime, &a.Status, &a.PublishedAt, &a.ViewCount,
		&a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrArticleNotFound
	}
	return a, err
}

func (r *pgArticleRepo) FindByID(ctx context.Context, id string) (*domain.Article, error) {
	a := &domain.Article{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, author_id, title, slug, COALESCE(subtitle, ''), content, COALESCE(cover_image, ''),
		        reading_time, status, published_at, view_count, created_at, updated_at
		 FROM articles WHERE id = $1`, id,
	).Scan(&a.ID, &a.AuthorID, &a.Title, &a.Slug, &a.Subtitle, &a.Content,
		&a.CoverImage, &a.ReadingTime, &a.Status, &a.PublishedAt, &a.ViewCount,
		&a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrArticleNotFound
	}
	return a, err
}

func (r *pgArticleRepo) Update(ctx context.Context, a *domain.Article) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE articles SET title = $1, subtitle = $2, content = $3, cover_image = $4,
		        reading_time = $5, status = $6, published_at = $7, updated_at = NOW()
		 WHERE id = $8`,
		a.Title, a.Subtitle, a.Content, a.CoverImage, a.ReadingTime,
		a.Status, a.PublishedAt, a.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrArticleNotFound
	}
	return nil
}

func (r *pgArticleRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM articles WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrArticleNotFound
	}
	return nil
}

func (r *pgArticleRepo) List(ctx context.Context, limit, offset int) ([]*domain.Article, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, author_id, title, slug, COALESCE(subtitle, ''), COALESCE(cover_image, ''),
		        reading_time, status, published_at, view_count, created_at, updated_at
		 FROM articles WHERE status = 'published'
		 ORDER BY published_at DESC LIMIT $1 OFFSET $2`, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

func (r *pgArticleRepo) ListByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*domain.Article, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, author_id, title, slug, COALESCE(subtitle, ''), COALESCE(cover_image, ''),
		        reading_time, status, published_at, view_count, created_at, updated_at
		 FROM articles WHERE author_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, authorID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

func (r *pgArticleRepo) FindOrCreateTag(ctx context.Context, name, slug string) (*domain.Tag, error) {
	tag := &domain.Tag{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tags (name, slug) VALUES ($1, $2)
		 ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id, name, slug`, name, slug,
	).Scan(&tag.ID, &tag.Name, &tag.Slug)
	return tag, err
}

func (r *pgArticleRepo) AddArticleTags(ctx context.Context, articleID string, tagIDs []string) error {
	for _, tagID := range tagIDs {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			articleID, tagID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *pgArticleRepo) GetArticleTags(ctx context.Context, articleID string) ([]domain.Tag, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT t.id, t.name, t.slug FROM tags t
		 JOIN article_tags at ON t.id = at.tag_id
		 WHERE at.article_id = $1`, articleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func scanArticles(rows pgx.Rows) ([]*domain.Article, error) {
	var articles []*domain.Article
	for rows.Next() {
		a := &domain.Article{}
		if err := rows.Scan(&a.ID, &a.AuthorID, &a.Title, &a.Slug, &a.Subtitle,
			&a.CoverImage, &a.ReadingTime, &a.Status, &a.PublishedAt,
			&a.ViewCount, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}

// ── Comment implementations ──

func (r *pgArticleRepo) CreateComment(ctx context.Context, c *domain.Comment) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO comments (article_id, user_id, parent_id, content)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		c.ArticleID, c.UserID, c.ParentID, c.Content,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *pgArticleRepo) ListCommentsByArticle(ctx context.Context, articleID string) ([]*domain.Comment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, article_id, user_id, parent_id, content, created_at, updated_at
		 FROM comments WHERE article_id = $1
		 ORDER BY created_at ASC`, articleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		c := &domain.Comment{}
		if err := rows.Scan(&c.ID, &c.ArticleID, &c.UserID, &c.ParentID,
			&c.Content, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *pgArticleRepo) DeleteComment(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM comments WHERE id = $1 AND user_id = $2`, id, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("comment not found or not owned by user")
	}
	return nil
}

// ── Clap implementations ──

func (r *pgArticleRepo) AddClap(ctx context.Context, userID, articleID string, count int) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO claps (user_id, article_id, count) VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, article_id) DO UPDATE SET count = claps.count + $3
		 RETURNING count`,
		userID, articleID, count,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *pgArticleRepo) GetClapCount(ctx context.Context, articleID string) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(count), 0) FROM claps WHERE article_id = $1`,
		articleID,
	).Scan(&total)
	return total, err
}

func (r *pgArticleRepo) GetUserClapCount(ctx context.Context, userID, articleID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT count FROM claps WHERE user_id = $1 AND article_id = $2`,
		userID, articleID,
	).Scan(&count)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return count, err
}

// ── Bookmark implementations ──

func (r *pgArticleRepo) AddBookmark(ctx context.Context, userID, articleID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO bookmarks (user_id, article_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, articleID,
	)
	return err
}

func (r *pgArticleRepo) RemoveBookmark(ctx context.Context, userID, articleID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM bookmarks WHERE user_id = $1 AND article_id = $2`,
		userID, articleID,
	)
	return err
}

func (r *pgArticleRepo) IsBookmarked(ctx context.Context, userID, articleID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM bookmarks WHERE user_id = $1 AND article_id = $2)`,
		userID, articleID,
	).Scan(&exists)
	return exists, err
}

func (r *pgArticleRepo) ListBookmarks(ctx context.Context, userID string, limit, offset int) ([]*domain.BookmarkInfo, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT a.id, a.title, a.slug, b.created_at
		 FROM bookmarks b JOIN articles a ON b.article_id = a.id
		 WHERE b.user_id = $1
		 ORDER BY b.created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []*domain.BookmarkInfo
	for rows.Next() {
		bm := &domain.BookmarkInfo{}
		var createdAt interface{}
		if err := rows.Scan(&bm.ArticleID, &bm.Title, &bm.Slug, &createdAt); err != nil {
			return nil, err
		}
		if t, ok := createdAt.(interface{ String() string }); ok {
			bm.CreatedAt = t.String()
		} else {
			bm.CreatedAt = "unknown"
		}
		bookmarks = append(bookmarks, bm)
	}
	return bookmarks, nil
}

// ── Search ──

func (r *pgArticleRepo) Search(ctx context.Context, query string, limit, offset int) ([]*domain.ArticleSearchResult, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT a.id, a.title, a.slug, COALESCE(a.subtitle, ''), COALESCE(a.cover_image, ''),
		        a.reading_time, a.published_at,
		        COALESCE(cc.clap_count, 0) AS clap_count,
		        COALESCE(cm.comment_count, 0) AS comment_count,
		        u.username, COALESCE(u.display_name, u.username),
		        ts_rank(to_tsvector('simple', a.title || ' ' || COALESCE(a.subtitle,'') || ' ' || a.content), plainto_tsquery('simple', $1)) AS rank
		         FROM articles a
		         LEFT JOIN users u ON a.author_id = u.id
		         LEFT JOIN (
		           SELECT article_id, SUM(count) AS clap_count FROM claps GROUP BY article_id
		         ) cc ON a.id = cc.article_id
		         LEFT JOIN (
		           SELECT article_id, COUNT(*) AS comment_count FROM comments GROUP BY article_id
		         ) cm ON a.id = cm.article_id
		         WHERE a.status = 'published'
		           AND to_tsvector('simple', a.title || ' ' || COALESCE(a.subtitle,'') || ' ' || a.content) @@ plainto_tsquery('simple', $1)
		 ORDER BY rank DESC
		 LIMIT $2 OFFSET $3`,
		query, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.ArticleSearchResult
	for rows.Next() {
		sr := &domain.ArticleSearchResult{}
		if err := rows.Scan(&sr.ID, &sr.Title, &sr.Slug, &sr.Subtitle, &sr.CoverImage,
			&sr.ReadingTime, &sr.PublishedAt, &sr.ClapCount, &sr.CommentCount,
			&sr.AuthorUsername, &sr.AuthorDisplayName, &sr.Rank); err != nil {
			return nil, err
		}
		results = append(results, sr)
	}
	return results, nil
}

// ── User Stats ──

func (r *pgArticleRepo) GetUserStats(ctx context.Context, userID string) (*domain.UserStats, error) {
	stats := &domain.UserStats{UserID: userID}

	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM articles WHERE author_id = $1 AND status = 'published'`,
		userID,
	).Scan(&stats.ArticleCount)
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(count), 0) FROM claps WHERE article_id IN
		 (SELECT id FROM articles WHERE author_id = $1)`,
		userID,
	).Scan(&stats.TotalClaps)
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(view_count), 0) FROM articles WHERE author_id = $1`,
		userID,
	).Scan(&stats.TotalViews)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
