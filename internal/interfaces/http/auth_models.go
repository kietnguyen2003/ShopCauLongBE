package http

import appAuth "kafka-order-demo/backend/internal/application/auth"

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	ResetToken  string `json:"reset_token"`
	NewPassword string `json:"new_password"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
	UserID       uint   `json:"user_id"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type userResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
}

type authResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         userResponse `json:"user"`
}

func toRegisterInput(req registerRequest) appAuth.RegisterRequest {
	return appAuth.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}
}

func toLoginInput(req loginRequest, ip string) appAuth.LoginRequest {
	return appAuth.LoginRequest{
		Username: req.Username,
		Password: req.Password,
		IP:       ip,
	}
}

func toChangePasswordInput(req changePasswordRequest, userID uint) appAuth.ChangePasswordRequest {
	return appAuth.ChangePasswordRequest{
		UserID:      userID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}
}

func toForgotPasswordInput(req forgotPasswordRequest) appAuth.ForgotPasswordRequest {
	return appAuth.ForgotPasswordRequest{
		Email: req.Email,
	}
}

func toResetPasswordInput(req resetPasswordRequest) appAuth.ResetPasswordRequest {
	return appAuth.ResetPasswordRequest{
		ResetToken:  req.ResetToken,
		NewPassword: req.NewPassword,
	}
}

func toAuthHTTPResponse(resp *appAuth.AuthResponse) authResponse {
	return authResponse{
		Token:        resp.Token,
		RefreshToken: resp.RefreshToken,
		User: userResponse{
			ID:       resp.User.ID,
			Username: resp.User.Username,
			Email:    resp.User.Email,
			IsAdmin:  resp.User.IsAdmin,
		},
	}
}

func toUserHTTPResponse(resp *appAuth.UserResponse) userResponse {
	return userResponse{
		ID:       resp.ID,
		Username: resp.Username,
		Email:    resp.Email,
		IsAdmin:  resp.IsAdmin,
	}
}
