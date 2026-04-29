package auth

import (
	"errors"

	domainAuth "kafka-order-demo/backend/internal/domain/auth"
)

type Service struct {
	userRepo       UserRepository
	passwordHasher PasswordHasher
	tokenProvider  TokenProvider
}

func NewService(userRepo UserRepository, passwordHasher PasswordHasher, tokenProvider TokenProvider) *Service {
	return &Service{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		tokenProvider:  tokenProvider,
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
	hashedPassword, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := domainAuth.NewUser(req.Username, req.Email, hashedPassword, false)
	if err != nil {
		return nil, err
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	// Generate token
	token, err := s.tokenProvider.Generate(user.ID, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil
}

func (s *Service) Login(req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errors.New("username không tồn tại")
	}

	err = s.passwordHasher.Compare(user.Password, req.Password)
	if err != nil {
		return nil, errors.New("mật khẩu không đúng")
	}

	token, err := s.tokenProvider.Generate(user.ID, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil
}

func (s *Service) AdminLogin(req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errors.New("username không tồn tại")
	}

	if !user.IsAdmin {
		return nil, errors.New("access denied")
	}

	err = s.passwordHasher.Compare(user.Password, req.Password)
	if err != nil {
		return nil, errors.New("mật khẩu không đúng")
	}

	token, err := s.tokenProvider.Generate(user.ID, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil
}
