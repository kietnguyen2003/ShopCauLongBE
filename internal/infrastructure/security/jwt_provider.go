package security

import (
	"errors"
	"time"

	appAuth "kafka-order-demo/backend/internal/application/auth"

	"github.com/golang-jwt/jwt/v5"
)

type JWTProvider struct {
	secret string
}

func NewJWTProvider(secret string) *JWTProvider {
	return &JWTProvider{secret: secret}
}

func (p *JWTProvider) Generate(userID uint, isAdmin bool) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"is_admin": isAdmin,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(p.secret))
}

func (p *JWTProvider) Validate(tokenString string) (*appAuth.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(p.secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userIDValue, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid user_id claim")
	}

	isAdminValue, ok := claims["is_admin"].(bool)
	if !ok {
		return nil, errors.New("invalid is_admin claim")
	}

	return &appAuth.TokenClaims{
		UserID:  uint(userIDValue),
		IsAdmin: isAdminValue,
	}, nil
}
