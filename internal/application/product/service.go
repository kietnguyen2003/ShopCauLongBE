package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	domainProduct "kafka-order-demo/backend/internal/domain/product"
)

var ErrCategoryNotFound = errors.New("category not found")

const productDetailCacheTTL = 10 * time.Minute
const categoryProductsCacheTTL = 10 * time.Minute

type Service struct {
	productRepo  ProductRepository
	categoryRepo CategoryRepository
	cache        CacheStore
}

func NewService(productRepo ProductRepository, categoryRepo CategoryRepository, cache CacheStore) *Service {
	return &Service{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		cache:        cache,
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

func (s *Service) GetProduct(ctx context.Context, id uint) (*ProductResponse, error) {
	cacheKey := productDetailCacheKey(id)
	if cached, ok := getCached[ProductResponse](ctx, s.cache, cacheKey); ok {
		return &cached, nil
	}

	prod, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := toProductResponse(prod)
	setCached(ctx, s.cache, cacheKey, response, productDetailCacheTTL)
	return &response, nil
}

func (s *Service) GetProductsByCategory(ctx context.Context, category string) ([]ProductResponse, error) {
	cacheKey := categoryProductsCacheKey(category)
	if cached, ok := getCached[[]ProductResponse](ctx, s.cache, cacheKey); ok {
		return cached, nil
	}

	products, err := s.productRepo.GetByCategory(category)
	if err != nil {
		return nil, err
	}

	responses := toProductResponses(products)
	setCached(ctx, s.cache, cacheKey, responses, categoryProductsCacheTTL)
	return responses, nil
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

	oldCategory := prod.Category
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

	s.invalidateProductCache(id, oldCategory, category)
	response := toProductResponse(prod)
	return &response, nil
}

func (s *Service) DeleteProduct(id uint) error {
	prod, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.Delete(id); err != nil {
		return err
	}

	s.invalidateProductCache(id, prod.Category)
	return nil
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

	if err := s.productRepo.Update(prod); err != nil {
		return err
	}

	s.invalidateProductCache(id, prod.Category)
	return nil
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

func (s *Service) invalidateProductCache(id uint, categories ...string) {
	if s.cache == nil {
		return
	}

	keys := []string{productDetailCacheKey(id)}
	seen := map[string]bool{}
	for _, category := range categories {
		cacheKey := categoryProductsCacheKey(category)
		if category == "" || seen[cacheKey] {
			continue
		}
		seen[cacheKey] = true
		keys = append(keys, cacheKey)
	}
	_ = s.cache.Delete(context.Background(), keys...)
}

func productDetailCacheKey(id uint) string {
	return fmt.Sprintf("product:detail:%d", id)
}

func categoryProductsCacheKey(category string) string {
	return "products:category:" + strings.ToLower(strings.TrimSpace(category))
}

func getCached[T any](ctx context.Context, cache CacheStore, key string) (T, bool) {
	var value T
	if cache == nil {
		return value, false
	}

	cached, err := cache.Get(ctx, key)
	if err != nil {
		return value, false
	}

	if err := json.Unmarshal([]byte(cached), &value); err != nil {
		return value, false
	}

	return value, true
}

func setCached(ctx context.Context, cache CacheStore, key string, value any, ttl time.Duration) {
	if cache == nil {
		return
	}

	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	_ = cache.Set(ctx, key, data, ttl)
}
