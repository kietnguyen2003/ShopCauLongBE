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

func (s *Service) SearchProducts(keyword string) ([]ProductResponse, error) {
	products, err := s.productRepo.Search(keyword)
	if err != nil {
		return nil, err
	}

	return toProductResponses(products), nil
}

func (s *Service) CreateProduct(req ProductRequest) (*ProductResponse, error) {
	prod, err := domainProduct.NewProduct(req.Name, req.Description, req.Price, req.Stock, req.Image, req.Category)
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

func (s *Service) UpdateProduct(id uint, req ProductRequest) (*ProductResponse, error) {
	prod, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	err = prod.UpdateDetails(req.Name, req.Description, req.Price, req.Stock, req.Image, req.Category)
	if err != nil {
		return nil, err
	}

	if err := s.productRepo.Update(prod); err != nil {
		return nil, err
	}

	response := toProductResponse(prod)
	return &response, nil
}

func (s *Service) DeleteProduct(id uint) error {
	if _, err := s.productRepo.GetByID(id); err != nil {
		return err
	}

	return s.productRepo.Delete(id)
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
