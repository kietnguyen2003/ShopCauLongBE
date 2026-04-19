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

type userResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

func toRegisterInput(req registerRequest) appAuth.RegisterRequest {
	return appAuth.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}
}

func toLoginInput(req loginRequest) appAuth.LoginRequest {
	return appAuth.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}
}

func toAuthHTTPResponse(resp *appAuth.AuthResponse) authResponse {
	return authResponse{
		Token: resp.Token,
		User: userResponse{
			ID:       resp.User.ID,
			Username: resp.User.Username,
			Email:    resp.User.Email,
			IsAdmin:  resp.User.IsAdmin,
		},
	}
}
