package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	domainAuth "kafka-order-demo/backend/internal/domain/auth"
)

const refreshTokenTTL = 7 * 24 * time.Hour
const sessionTTL = 24 * time.Hour
const loginRateLimitWindow = 5 * time.Minute
const maxLoginFailures = 5

var ErrTooManyLoginAttempts = errors.New("too many login attempts, please try again later")

type Service struct {
	userRepo          UserRepository
	passwordHasher    PasswordHasher
	tokenProvider     TokenProvider
	refreshTokenStore RefreshTokenStore
	loginRateLimiter  LoginRateLimiter
	sessionStore      SessionStore
}

func NewService(userRepo UserRepository, passwordHasher PasswordHasher, tokenProvider TokenProvider, authCache AuthCache) *Service {
	return &Service{
		userRepo:          userRepo,
		passwordHasher:    passwordHasher,
		tokenProvider:     tokenProvider,
		refreshTokenStore: authCache,
		loginRateLimiter:  authCache,
		sessionStore:      authCache,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
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

	return s.issueAuthTokens(ctx, user)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	if err := s.ensureLoginAllowed(ctx, req.Username, req.IP); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		_ = s.recordLoginFailure(ctx, req.Username, req.IP)
		return nil, errors.New("username không tồn tại")
	}

	err = s.passwordHasher.Compare(user.Password, req.Password)
	if err != nil {
		_ = s.recordLoginFailure(ctx, req.Username, req.IP)
		return nil, errors.New("mật khẩu không đúng")
	}

	if err := s.loginRateLimiter.ClearLoginFailures(ctx, req.Username, req.IP); err != nil {
		return nil, err
	}

	return s.issueAuthTokens(ctx, user)
}

func (s *Service) AdminLogin(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	if err := s.ensureLoginAllowed(ctx, req.Username, req.IP); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		_ = s.recordLoginFailure(ctx, req.Username, req.IP)
		return nil, errors.New("username không tồn tại")
	}

	if !user.IsAdmin {
		_ = s.recordLoginFailure(ctx, req.Username, req.IP)
		return nil, errors.New("access denied")
	}

	err = s.passwordHasher.Compare(user.Password, req.Password)
	if err != nil {
		_ = s.recordLoginFailure(ctx, req.Username, req.IP)
		return nil, errors.New("mật khẩu không đúng")
	}

	if err := s.loginRateLimiter.ClearLoginFailures(ctx, req.Username, req.IP); err != nil {
		return nil, err
	}

	return s.issueAuthTokens(ctx, user)
}

func (s *Service) GetMe(ctx context.Context, userID uint) (*UserResponse, error) {
	session, err := s.sessionStore.GetSession(ctx, userID)
	if err != nil {
		return nil, errors.New("session not found")
	}

	return &UserResponse{
		ID:       session.UserID,
		Username: session.Username,
		Email:    session.Email,
		IsAdmin:  session.IsAdmin,
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string, userIDReq uint) (*AuthResponse, error) {
	userID, err := s.refreshTokenStore.GetRefreshTokenUserID(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if userID != userIDReq {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if err := s.refreshTokenStore.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	return s.issueAuthTokens(ctx, user)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	userID, err := s.refreshTokenStore.GetRefreshTokenUserID(ctx, refreshToken)
	if err != nil {
		return s.refreshTokenStore.DeleteRefreshToken(ctx, refreshToken)
	}

	if err := s.refreshTokenStore.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return err
	}

	return s.sessionStore.DeleteSession(ctx, userID)
}

func (s *Service) GetSession(ctx context.Context, userID uint) (*Session, error) {
	return s.sessionStore.GetSession(ctx, userID)
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

func (s *Service) issueAuthTokens(ctx context.Context, user *domainAuth.User) (*AuthResponse, error) {
	token, err := s.tokenProvider.Generate(user.ID, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := s.refreshTokenStore.StoreRefreshToken(ctx, refreshToken, user.ID, refreshTokenTTL); err != nil {
		return nil, err
	}

	if err := s.sessionStore.StoreSession(ctx, Session{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        userRole(user.IsAdmin),
		Permissions: []string{},
		IsAdmin:     user.IsAdmin,
	}, sessionTTL); err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         toUserResponse(user),
	}, nil
}

func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func (s *Service) ensureLoginAllowed(ctx context.Context, username, ip string) error {
	failures, err := s.loginRateLimiter.GetLoginFailureCount(ctx, username, ip)
	if err != nil {
		return err
	}
	if failures >= maxLoginFailures {
		return ErrTooManyLoginAttempts
	}
	return nil
}

func (s *Service) recordLoginFailure(ctx context.Context, username, ip string) error {
	failures, err := s.loginRateLimiter.IncrementLoginFailure(ctx, username, ip, loginRateLimitWindow)
	if err != nil {
		return err
	}
	if failures > maxLoginFailures {
		return ErrTooManyLoginAttempts
	}
	return nil
}

func userRole(isAdmin bool) string {
	if isAdmin {
		return "admin"
	}
	return "user"
}
