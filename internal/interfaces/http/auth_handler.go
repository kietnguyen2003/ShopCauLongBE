package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	appAuth "kafka-order-demo/backend/internal/application/auth"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Register(toRegisterInput(req))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toAuthHTTPResponse(resp))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(toLoginInput(req))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toAuthHTTPResponse(resp))
}

func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.AdminLogin(toLoginInput(req))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toAuthHTTPResponse(resp))
}

func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims, err := h.tokenProvider.Validate(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
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
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
