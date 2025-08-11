package handlers

import (
	"net/http"

	"kafka-order-demo/backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{
		db: db,
	}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	var products []models.Product
	
	if err := h.db.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	
	if err := h.db.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}