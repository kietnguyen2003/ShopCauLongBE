package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appProduct "kafka-order-demo/backend/internal/application/product"
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
	products, err := h.productService.GetProducts()
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get products successfully", toProductHTTPResponses(products))
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

	var req productRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	product, err := h.productService.UpdateProduct(uint(id), toProductInput(req))
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
