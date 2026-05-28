package usecase

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/arifkurniawan200/platform-blog/services/article/internal/domain"
	"github.com/arifkurniawan200/platform-blog/services/article/internal/repository"
	"github.com/google/uuid"
)

// ArticleUsecase handles article business logic
type ArticleUsecase struct {
	repo repository.ArticleRepository
}

// NewArticleUsecase creates a new article usecase
func NewArticleUsecase(repo repository.ArticleRepository) *ArticleUsecase {
	return &ArticleUsecase{repo: repo}
}

// Create creates a new article as draft
func (uc *ArticleUsecase) Create(ctx context.Context, authorID string, req *domain.CreateArticleRequest) (*domain.Article, error) {
	slug := GenerateSlug(req.Title)

	article := &domain.Article{
		AuthorID:    authorID,
		Title:       req.Title,
		Slug:        slug,
		Subtitle:    req.Subtitle,
		Content:     req.Content,
		CoverImage:  req.CoverImage,
		ReadingTime: EstimateReadingTime(req.Content),
		Status:      domain.StatusDraft,
	}

	if err := uc.repo.Create(ctx, article); err != nil {
		return nil, fmt.Errorf("create article: %w", err)
	}

	// Save tags if provided
	if len(req.Tags) > 0 {
		if err := uc.saveTags(ctx, article.ID, req.Tags); err != nil {
			return nil, err
		}
	}

	return article, nil
}

// GetBySlug retrieves an article by its slug
func (uc *ArticleUsecase) GetBySlug(ctx context.Context, slug string) (*domain.Article, error) {
	article, err := uc.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	tags, err := uc.repo.GetArticleTags(ctx, article.ID)
	if err == nil {
		article.Tags = tags
	}
	return article, nil
}

// List returns published articles with pagination
func (uc *ArticleUsecase) List(ctx context.Context, limit, offset int) ([]*domain.Article, error) {
	return uc.repo.List(ctx, limit, offset)
}

func (uc *ArticleUsecase) saveTags(ctx context.Context, articleID string, tagNames []string) error {
	var tagIDs []string
	for _, name := range tagNames {
		slug := GenerateSlug(name)
		tag, err := uc.repo.FindOrCreateTag(ctx, name, slug)
		if err != nil {
			return fmt.Errorf("save tag %q: %w", name, err)
		}
		tagIDs = append(tagIDs, tag.ID)
	}
	return uc.repo.AddArticleTags(ctx, articleID, tagIDs)
}

// ---- Helpers ----

var nonAlphaRegex = regexp.MustCompile(`[^a-z0-9]+`)

// GenerateSlug creates a URL-friendly slug from a title
func GenerateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' || r == '-' {
			return r
		}
		return ' '
	}, slug)
	slug = nonAlphaRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = uuid.New().String()[:8]
	}
	return slug
}

// EstimateReadingTime estimates reading time in minutes (avg 200 wpm)
func EstimateReadingTime(content string) int {
	words := strings.Fields(content)
	minutes := int(math.Ceil(float64(len(words)) / 200.0))
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}
