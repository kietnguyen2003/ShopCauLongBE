package product

import (
	"errors"
	domainProduct "kafka-order-demo/backend/internal/domain/product"
	"math"
)

var ErrCategoryNotFound = errors.New("category not found")

type Service struct {
	productRepo  ProductRepository
	categoryRepo CategoryRepository
}

func NewService(productRepo ProductRepository, categoryRepo CategoryRepository) *Service {
	return &Service{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *Service) GetProducts() ([]ProductResponse, error) {
	products, err := s.productRepo.GetAll()
	if err != nil {
		return nil, err
	}

	return toProductResponses(products), nil
}

func (s *Service) GetProductsWithQuery(query ProductQuery) (*ProductListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 12
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	products, total, err := s.productRepo.GetWithQuery(query)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(query.Limit)))
	}

	return &ProductListResponse{
		Items:      toProductResponses(products),
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
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
	categoryID, categoryName, err := s.resolveCategory(req.CategoryID, req.Category)
	if err != nil {
		return nil, err
	}

	prod, err := domainProduct.NewProduct(req.Name, req.Description, req.Price, req.Stock, req.Image, categoryID, categoryName)
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

func (s *Service) UpdateProduct(id uint, req ProductUpdateRequest) (*ProductResponse, error) {
	prod, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	name := prod.Name
	description := prod.Description
	price := prod.Price
	stock := prod.Stock
	image := prod.Image
	categoryID := prod.CategoryID
	category := prod.Category

	if req.Name != nil {
		name = *req.Name
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.Price != nil {
		price = *req.Price
	}
	if req.Stock != nil {
		stock = *req.Stock
	}
	if req.Image != nil {
		image = *req.Image
	}
	if req.CategoryID != nil {
		categoryID = *req.CategoryID
	}
	if req.Category != nil {
		category = *req.Category
	}

	categoryID, category, err = s.resolveCategory(categoryID, category)
	if err != nil {
		return nil, err
	}

	err = prod.UpdateDetails(name, description, price, stock, image, categoryID, category)
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

func (s *Service) resolveCategory(categoryID uint, categoryName string) (uint, string, error) {
	if categoryID != 0 {
		category, err := s.categoryRepo.GetByID(categoryID)
		if err != nil {
			return 0, "", ErrCategoryNotFound
		}
		return category.ID, category.Name, nil
	}

	if categoryName == "" {
		return 0, "", ErrCategoryNotFound
	}

	category, err := s.categoryRepo.GetByName(categoryName)
	if err != nil {
		return 0, "", ErrCategoryNotFound
	}

	return category.ID, category.Name, nil
}
