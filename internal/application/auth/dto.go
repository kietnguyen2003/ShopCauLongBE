package auth

import domainAuth "kafka-order-demo/backend/internal/domain/auth"

type LoginRequest struct {
	Username string
	Password string
	IP       string
}

type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

type ChangePasswordRequest struct {
	UserID      uint
	OldPassword string
	NewPassword string
}

type ForgotPasswordRequest struct {
	Email string
}

type ResetPasswordRequest struct {
	ResetToken  string
	NewPassword string
}

type UserResponse struct {
	ID       uint
	Username string
	Email    string
	IsAdmin  bool
}

type Session struct {
	UserID      uint     `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	IsAdmin     bool     `json:"is_admin"`
}

type AuthResponse struct {
	Token        string
	RefreshToken string
	User         UserResponse `json:"user"`
}

func toUserResponse(user *domainAuth.User) UserResponse {
	return UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		IsAdmin:  user.IsAdmin,
	}
}
