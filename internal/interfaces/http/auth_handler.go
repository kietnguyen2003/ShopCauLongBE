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

func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
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
