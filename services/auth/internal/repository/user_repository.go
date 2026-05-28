package repository

import (
	"context"
	"errors"
	"time"

	"github.com/arifkurniawan200/platform-blog/services/auth/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserExists        = errors.New("user already exists")
	ErrUsernameTaken     = errors.New("username already taken")
	ErrEmailTaken        = errors.New("email already registered")
)

// UserRepository defines the data access interface
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	List(ctx context.Context, limit, offset int) ([]*domain.User, error)
	GetArticleCount(ctx context.Context, authorID string) (int, error)
}

type pgUserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a PostgreSQL user repository
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &pgUserRepo{pool: pool}
}

func (r *pgUserRepo) Create(ctx context.Context, user *domain.User) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash, display_name)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		user.Username, user.Email, user.PasswordHash, user.DisplayName,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *pgUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, display_name, bio, avatar_url, email_notify_comments, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName,
		&user.Bio, &user.AvatarURL, &user.EmailNotify, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return user, err
}

func (r *pgUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, display_name, bio, avatar_url, email_notify_comments, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.Bio, &user.AvatarURL, &user.EmailNotify, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return user, err
}

func (r *pgUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	user := &domain.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, display_name, bio, avatar_url, email_notify_comments, created_at, updated_at
		 FROM users WHERE username = $1`, username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName,
		&user.Bio, &user.AvatarURL, &user.EmailNotify, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return user, err
}

func (r *pgUserRepo) Update(ctx context.Context, user *domain.User) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET display_name = $1, bio = $2, avatar_url = $3, email_notify_comments = $4, updated_at = NOW()
		 WHERE id = $5`,
		user.DisplayName, user.Bio, user.AvatarURL, user.EmailNotify, user.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *pgUserRepo) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, username, email, display_name, bio, avatar_url, email_notify_comments, created_at, updated_at
		 FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName,
			&u.Bio, &u.AvatarURL, &u.EmailNotify, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *pgUserRepo) GetArticleCount(ctx context.Context, authorID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM articles WHERE author_id = $1 AND status = 'published'`,
		authorID,
	).Scan(&count)
	return count, err
}
