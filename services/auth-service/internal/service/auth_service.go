package service

import (
	"errors"
	"time"
	
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(req RegisterRequest) (*AuthResponse, error) {
	existingUser, _ := s.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	existingUser, _ = s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := models.NewUser(req.Username, req.Email, string(hashedPassword), false)
	if err != nil {
		return nil, err
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	token, err := s.generateToken(user.ID, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.generateToken(user.ID, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) AdminLogin(req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsAdmin {
		return nil, errors.New("access denied")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.generateToken(user.ID, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID := uint(claims["user_id"].(float64))
	return s.userRepo.GetByID(userID)
}

func (s *AuthService) generateToken(userID uint, isAdmin bool) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"is_admin": isAdmin,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}