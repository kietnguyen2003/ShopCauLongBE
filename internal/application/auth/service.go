package auth

import (
	"errors"
	"time"
	
	"kafka-order-demo/backend/internal/domain/auth"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	userRepo  auth.UserRepository
	jwtSecret string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string    `json:"token"`
	User  *auth.User `json:"user"`
}

func NewService(userRepo auth.UserRepository, jwtSecret string) *Service {
	return &Service{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) Register(req RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	existingUser, _ = s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := auth.NewUser(req.Username, req.Email, string(hashedPassword), false)
	if err != nil {
		return nil, err
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	// Generate token
	token, err := s.generateToken(user.ID, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *Service) Login(req LoginRequest) (*AuthResponse, error) {
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

func (s *Service) AdminLogin(req LoginRequest) (*AuthResponse, error) {
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

func (s *Service) generateToken(userID uint, isAdmin bool) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"is_admin": isAdmin,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}