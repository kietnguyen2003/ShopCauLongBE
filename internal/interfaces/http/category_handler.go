package http

import (
	"net/http"
	"strconv"

	appCategory "kafka-order-demo/backend/internal/application/category"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService *appCategory.Service
}

func NewCategoryHandler(categoryService *appCategory.Service) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.categoryService.GetCategories()
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get categories successfully", toCategoryHTTPResponses(categories))
}

func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	category, err := h.categoryService.GetCategory(uint(id))
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Category not found")
		return
	}

	successResponse(c, http.StatusOK, "Get category successfully", toCategoryHTTPResponse(*category))
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	category, err := h.categoryService.CreateCategory(toCategoryInput(req))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusCreated, "Create category successfully", toCategoryHTTPResponse(*category))
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	category, err := h.categoryService.UpdateCategory(uint(id), toCategoryInput(req))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Update category successfully", toCategoryHTTPResponse(*category))
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	if err := h.categoryService.DeleteCategory(uint(id)); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Delete category successfully", nil)
}
