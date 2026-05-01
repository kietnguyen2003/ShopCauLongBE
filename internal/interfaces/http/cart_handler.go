package http

import (
	"net/http"
	"strconv"

	appCart "kafka-order-demo/backend/internal/application/cart"

	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	cartService *appCart.Service
}

func NewCartHandler(cartService *appCart.Service) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	cart, err := h.cartService.GetCart(userID.(uint))
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get cart successfully", toCartHTTPResponse(*cart))
}

func (h *CartHandler) AddItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req addCartItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	cart, err := h.cartService.AddItems(toCartItemsInput(req, userID.(uint)))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusCreated, "Add cart item successfully", toCartHTTPResponse(*cart))
}

func (h *CartHandler) UpdateItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := parseCartItemID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid cart item ID")
		return
	}

	var req updateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	cart, err := h.cartService.UpdateItemQuantity(userID.(uint), uint(id), req.Quantity)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Update cart item successfully", toCartHTTPResponse(*cart))
}

func (h *CartHandler) DeleteItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := parseCartItemID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid cart item ID")
		return
	}

	cart, err := h.cartService.DeleteItem(userID.(uint), uint(id))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Delete cart item successfully", toCartHTTPResponse(*cart))
}

func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := h.cartService.ClearCart(userID.(uint)); err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Clear cart successfully", nil)
}

func parseCartItemID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 32)
}
