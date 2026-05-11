package auth

import (
	"context"
	"time"

	domainAuth "kafka-order-demo/backend/internal/domain/auth"
)

type UserRepository interface {
	Create(user *domainAuth.User) error
	GetByID(id uint) (*domainAuth.User, error)
	GetByUsername(username string) (*domainAuth.User, error)
	GetByEmail(email string) (*domainAuth.User, error)
	Update(user *domainAuth.User) error
	Delete(id uint) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type TokenProvider interface {
	Generate(userID uint, isAdmin bool) (string, error)
	Validate(token string) (*TokenClaims, error)
}

type RefreshTokenStore interface {
	StoreRefreshToken(ctx context.Context, token string, userID uint, expiration time.Duration) error
	GetRefreshTokenUserID(ctx context.Context, token string) (uint, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

type LoginRateLimiter interface {
	GetLoginFailureCount(ctx context.Context, username, ip string) (int64, error)
	IncrementLoginFailure(ctx context.Context, username, ip string, expiration time.Duration) (int64, error)
	ClearLoginFailures(ctx context.Context, username, ip string) error
}

type SessionStore interface {
	StoreSession(ctx context.Context, session Session, expiration time.Duration) error
	GetSession(ctx context.Context, userID uint) (*Session, error)
	DeleteSession(ctx context.Context, userID uint) error
}

type AuthCache interface {
	RefreshTokenStore
	LoginRateLimiter
	SessionStore
}

type TokenClaims struct {
	UserID  uint
	IsAdmin bool
}
