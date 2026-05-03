package http

import (
	"net/http"
	"strconv"

	appReview "kafka-order-demo/backend/internal/application/review"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	reviewService *appReview.Service
}

func NewReviewHandler(reviewService *appReview.Service) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

func (h *ReviewHandler) GetProductReviews(c *gin.Context) {
	productID, err := parseUintParam(c, "id", "Invalid product ID")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	reviews, err := h.reviewService.GetProductReviews(productID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Product not found")
		return
	}

	successResponse(c, http.StatusOK, "Get product reviews successfully", toReviewHTTPResponses(reviews))
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	productID, err := parseUintParam(c, "id", "Invalid product ID")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req createReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	review, err := h.reviewService.CreateReview(toReviewInput(req, productID, userID.(uint)))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusCreated, "Create review successfully", toReviewHTTPResponse(*review))
}

func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	reviewID, err := parseUintParam(c, "id", "Invalid review ID")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req updateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	review, err := h.reviewService.UpdateReview(reviewID, toReviewUpdateInput(req, userID.(uint)))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Update review successfully", toReviewHTTPResponse(*review))
}

func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	reviewID, err := parseUintParam(c, "id", "Invalid review ID")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := h.reviewService.DeleteReview(reviewID, userID.(uint)); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Delete review successfully", nil)
}

func parseUintParam(c *gin.Context, key, message string) (uint, error) {
	id, err := strconv.ParseUint(c.Param(key), 10, 32)
	if err != nil {
		return 0, &paramError{message: message}
	}

	return uint(id), nil
}

type paramError struct {
	message string
}

func (e *paramError) Error() string {
	return e.message
}
