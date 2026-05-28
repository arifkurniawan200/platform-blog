package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Exp   int64  `json:"exp"`
	Iat   int64  `json:"iat"`
}

// JWTMiddleware validates JWT tokens and injects claims into context
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"success":false,"error":{"code":"Unauthorized","message":"Missing Authorization header"}}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"success":false,"error":{"code":"Unauthorized","message":"Invalid Authorization header format"}}`, http.StatusUnauthorized)
			return
		}

		claims, err := ValidateJWT(parts[1])
		if err != nil {
			http.Error(w, `{"success":false,"error":{"code":"Unauthorized","message":"`+err.Error()+`"}}`, http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = WithClaims(ctx, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ValidateJWT validates a JWT token and returns claims (demo implementation)
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if claims.Exp > 0 && time.Unix(claims.Exp, 0).Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

type contextKey string

const claimsContextKey contextKey = "claims"

// WithClaims adds claims to context
func WithClaims(ctx context.Context, claims *JWTClaims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// GetClaims retrieves claims from context
func GetClaims(ctx context.Context) (*JWTClaims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*JWTClaims)
	return claims, ok
}

// Errors
var (
	ErrInvalidToken    = &JWTError{Message: "invalid token"}
	ErrTokenExpired    = &JWTError{Message: "token expired"}
	ErrTokenNotYetValid = &JWTError{Message: "token not yet valid"}
)

type JWTError struct{ Message string }

func (e *JWTError) Error() string { return e.Message }
