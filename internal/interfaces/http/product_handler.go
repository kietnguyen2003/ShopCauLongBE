package http

import (
	"errors"
	"net/http"
	"strconv"

	appProduct "kafka-order-demo/backend/internal/application/product"

	"github.com/gin-gonic/gin"
)

const (
	defaultProductPage  = 1
	defaultProductLimit = 12
	maxProductLimit     = 100
)

type ProductHandler struct {
	productService *appProduct.Service
}

func NewProductHandler(productService *appProduct.Service) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	query, err := parseProductQuery(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	products, err := h.productService.GetProductsWithQuery(query)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get products successfully", toProductListHTTPResponse(*products))
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	product, err := h.productService.GetProduct(uint(id))
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Product not found")
		return
	}

	successResponse(c, http.StatusOK, "Get product successfully", toProductHTTPResponse(*product))
}

func (h *ProductHandler) SearchProducts(c *gin.Context) {
	keyword := c.Query("q")

	products, err := h.productService.SearchProducts(keyword)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Search products successfully", toProductHTTPResponses(products))
}

func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
	category := c.Param("category")

	products, err := h.productService.GetProductsByCategory(category)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get products by category successfully", toProductHTTPResponses(products))
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req productRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	product, err := h.productService.CreateProduct(toProductInput(req))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusCreated, "Create product successfully", toProductHTTPResponse(*product))
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var req productUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	product, err := h.productService.UpdateProduct(uint(id), toProductUpdateInput(req))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Update product successfully", toProductHTTPResponse(*product))
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	if err := h.productService.DeleteProduct(uint(id)); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Delete product successfully", nil)
}

func (h *ProductHandler) UpdateProductStock(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var req updateProductStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.productService.UpdateProductStock(uint(id), req.Stock); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Update product stock successfully", nil)
}

func parseProductQuery(c *gin.Context) (appProduct.ProductQuery, error) {
	page, err := parsePositiveIntQuery(c, "page", defaultProductPage)
	if err != nil {
		return appProduct.ProductQuery{}, err
	}

	limit, err := parsePositiveIntQuery(c, "limit", defaultProductLimit)
	if err != nil {
		return appProduct.ProductQuery{}, err
	}
	if limit > maxProductLimit {
		limit = maxProductLimit
	}

	minPrice, err := parseOptionalNonNegativeInt64Query(c, "min_price")
	if err != nil {
		return appProduct.ProductQuery{}, err
	}

	maxPrice, err := parseOptionalNonNegativeInt64Query(c, "max_price")
	if err != nil {
		return appProduct.ProductQuery{}, err
	}

	if minPrice != nil && maxPrice != nil && *minPrice > *maxPrice {
		return appProduct.ProductQuery{}, errors.New("min_price cannot be greater than max_price")
	}

	categoryID, err := parseOptionalPositiveUintQuery(c, "category_id")
	if err != nil {
		return appProduct.ProductQuery{}, err
	}

	return appProduct.ProductQuery{
		Page:       page,
		Limit:      limit,
		Search:     c.Query("search"),
		CategoryID: categoryID,
		Category:   c.Query("category"),
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		Sort:       c.Query("sort"),
	}, nil
}

func parsePositiveIntQuery(c *gin.Context, key string, defaultValue int) (int, error) {
	value := c.Query(key)
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, errors.New(key + " must be a positive integer")
	}

	return parsed, nil
}

func parseOptionalNonNegativeInt64Query(c *gin.Context, key string) (*int64, error) {
	value := c.Query(key)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return nil, errors.New(key + " must be a non-negative integer")
	}

	return &parsed, nil
}

func parseOptionalPositiveUintQuery(c *gin.Context, key string) (uint, error) {
	value := c.Query(key)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil || parsed == 0 {
		return 0, errors.New(key + " must be a positive integer")
	}

	return uint(parsed), nil
}
