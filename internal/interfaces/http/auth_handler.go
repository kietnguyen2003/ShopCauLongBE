package http

import (
	"fmt"
	"net/http"
	"strings"

	appAuth "kafka-order-demo/backend/internal/application/auth"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService   *appAuth.Service
	tokenProvider appAuth.TokenProvider
}

func NewAuthHandler(authService *appAuth.Service, tokenProvider appAuth.TokenProvider) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		tokenProvider: tokenProvider,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.authService.Register(toRegisterInput(req))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusCreated, "Register successfully", toAuthHTTPResponse(resp))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.authService.Login(toLoginInput(req))
	if err != nil {
		errorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Login successfully", toAuthHTTPResponse(resp))
}

func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	fmt.Println("Request body:", req)

	resp, err := h.authService.AdminLogin(toLoginInput(req))
	if err != nil {
		errorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Admin login successfully", toAuthHTTPResponse(resp))
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	resp, err := h.authService.GetMe(userID.(uint))
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get current user successfully", toUserHTTPResponse(resp))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	successResponse(c, http.StatusOK, "Logout successfully", nil)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	resp, err := h.authService.RefreshToken(userID.(uint))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Refresh token successfully", toAuthHTTPResponse(resp))
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.authService.ChangePassword(toChangePasswordInput(req, userID.(uint))); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Change password successfully", nil)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	resetToken, err := h.authService.ForgotPassword(toForgotPasswordInput(req))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Forgot password successfully", gin.H{"reset_token": resetToken})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.authService.ResetPassword(toResetPasswordInput(req)); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Reset password successfully", nil)
}

func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}
		if tokenString == "" {
			tokenString = c.Query("token")
		}
		if tokenString == "" {
			errorResponse(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		claims, err := h.tokenProvider.Validate(tokenString)
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, "Invalid token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("is_admin", claims.IsAdmin)

		c.Next()
	}
}

func (h *AuthHandler) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			errorResponse(c, http.StatusForbidden, "Admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
