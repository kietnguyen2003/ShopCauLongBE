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

func (s *Service) GetMe(userID uint) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	response := toUserResponse(user)
	return &response, nil
}

func (s *Service) RefreshToken(userID uint) (*AuthResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
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

func (s *Service) ChangePassword(req ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	err = s.passwordHasher.Compare(user.Password, req.OldPassword)
	if err != nil {
		return errors.New("old password is incorrect")
	}

	hashedPassword, err := s.passwordHasher.Hash(req.NewPassword)
	if err != nil {
		return err
	}

	if err := user.UpdatePassword(hashedPassword); err != nil {
		return err
	}

	return s.userRepo.Update(user)
}

func (s *Service) ForgotPassword(req ForgotPasswordRequest) (string, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return "", errors.New("email not found")
	}

	return s.tokenProvider.Generate(user.ID, user.IsAdmin)
}

func (s *Service) ResetPassword(req ResetPasswordRequest) error {
	claims, err := s.tokenProvider.Validate(req.ResetToken)
	if err != nil {
		return errors.New("invalid reset token")
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	hashedPassword, err := s.passwordHasher.Hash(req.NewPassword)
	if err != nil {
		return err
	}

	if err := user.UpdatePassword(hashedPassword); err != nil {
		return err
	}

	return s.userRepo.Update(user)
}
