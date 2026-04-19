package auth

import domainAuth "kafka-order-demo/backend/internal/domain/auth"

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

type TokenClaims struct {
	UserID  uint
	IsAdmin bool
}
