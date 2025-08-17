package service

import (
	"errors"
	"product-service/internal/models"
	"product-service/internal/repository"
)

type ProductService struct {
	productRepo repository.ProductRepository
}

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"required,gte=0"`
	Image       string  `json:"image"`
	Category    string  `json:"category"`
}

type UpdateStockRequest struct {
	Stock int `json:"stock" binding:"required,gte=0"`
}

func NewProductService(productRepo repository.ProductRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

func (s *ProductService) CreateProduct(req CreateProductRequest) (*models.Product, error) {
	product, err := models.NewProduct(req.Name, req.Description, req.Price, req.Stock, req.Image, req.Category)
	if err != nil {
		return nil, err
	}

	err = s.productRepo.Create(product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetProducts() ([]*models.Product, error) {
	return s.productRepo.GetAll()
}

func (s *ProductService) GetProductByID(id uint) (*models.Product, error) {
	return s.productRepo.GetByID(id)
}

func (s *ProductService) GetProductsByCategory(category string) ([]*models.Product, error) {
	return s.productRepo.GetByCategory(category)
}

func (s *ProductService) UpdateStock(id uint, stock int) error {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return errors.New("product not found")
	}

	err = product.UpdateStock(stock)
	if err != nil {
		return err
	}

	return s.productRepo.Update(product)
}

func (s *ProductService) DecreaseStock(id uint, quantity int) error {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return errors.New("product not found")
	}

	err = product.DecreaseStock(quantity)
	if err != nil {
		return err
	}

	return s.productRepo.Update(product)
}

func (s *ProductService) DeleteProduct(id uint) error {
	return s.productRepo.Delete(id)
}