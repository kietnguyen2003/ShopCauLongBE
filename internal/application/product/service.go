package product

import (
	"kafka-order-demo/backend/internal/domain/product"
)

type Service struct {
	productRepo product.ProductRepository
}

func NewService(productRepo product.ProductRepository) *Service {
	return &Service{
		productRepo: productRepo,
	}
}

func (s *Service) GetProducts() ([]*product.Product, error) {
	return s.productRepo.GetAll()
}

func (s *Service) GetProduct(id uint) (*product.Product, error) {
	return s.productRepo.GetByID(id)
}

func (s *Service) GetProductsByCategory(category string) ([]*product.Product, error) {
	return s.productRepo.GetByCategory(category)
}

func (s *Service) CreateProduct(name, description string, price float64, stock int, image, category string) (*product.Product, error) {
	prod, err := product.NewProduct(name, description, price, stock, image, category)
	if err != nil {
		return nil, err
	}

	err = s.productRepo.Create(prod)
	if err != nil {
		return nil, err
	}

	return prod, nil
}

func (s *Service) UpdateProductStock(id uint, stock int) error {
	prod, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}

	err = prod.UpdateStock(stock)
	if err != nil {
		return err
	}

	return s.productRepo.Update(prod)
}