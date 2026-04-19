package product

import (
	domainProduct "kafka-order-demo/backend/internal/domain/product"
)

type Service struct {
	productRepo ProductRepository
}

func NewService(productRepo ProductRepository) *Service {
	return &Service{
		productRepo: productRepo,
	}
}

func (s *Service) GetProducts() ([]ProductResponse, error) {
	products, err := s.productRepo.GetAll()
	if err != nil {
		return nil, err
	}

	return toProductResponses(products), nil
}

func (s *Service) GetProduct(id uint) (*ProductResponse, error) {
	prod, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := toProductResponse(prod)
	return &response, nil
}

func (s *Service) GetProductsByCategory(category string) ([]ProductResponse, error) {
	products, err := s.productRepo.GetByCategory(category)
	if err != nil {
		return nil, err
	}

	return toProductResponses(products), nil
}

func (s *Service) CreateProduct(name, description string, price float64, stock int, image, category string) (*ProductResponse, error) {
	prod, err := domainProduct.NewProduct(name, description, price, stock, image, category)
	if err != nil {
		return nil, err
	}

	err = s.productRepo.Create(prod)
	if err != nil {
		return nil, err
	}

	response := toProductResponse(prod)
	return &response, nil
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
