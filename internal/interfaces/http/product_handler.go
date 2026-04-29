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
