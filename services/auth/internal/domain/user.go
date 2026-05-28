package domain

import "time"

// User represents a registered user
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
	Bio          string    `json:"bio,omitempty"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	EmailNotify  bool      `json:"email_notify_comments"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RegisterRequest is the payload for user registration
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest is the payload for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is returned on successful login/register
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

// Follow represents a follower relationship
type Follow struct {
	FollowerID  string    `json:"follower_id"`
	FollowingID string    `json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// UpdateProfileRequest is the payload for updating user profile
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name" validate:"omitempty,min=1,max=100"`
	Bio         *string `json:"bio" validate:"omitempty,max=500"`
	AvatarURL   *string `json:"avatar_url" validate:"omitempty,url"`
	EmailNotify *bool   `json:"email_notify_comments"`
}

// ProfileResponse is the user profile returned to clients
type ProfileResponse struct {
	User
	ArticleCount int `json:"article_count"`
}
