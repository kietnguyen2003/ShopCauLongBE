package category

import (
	"context"
	"encoding/json"
	"time"

	domainCategory "kafka-order-demo/backend/internal/domain/category"
)

const categoriesCacheKey = "categories:all"
const categoriesCacheTTL = 45 * time.Minute

type Service struct {
	categoryRepo CategoryRepository
	cache        CacheStore
}

func NewService(categoryRepo CategoryRepository, cache CacheStore) *Service {
	return &Service{
		categoryRepo: categoryRepo,
		cache:        cache,
	}
}

func (s *Service) GetCategories(ctx context.Context) ([]CategoryResponse, error) {
	if cached, ok := s.getCachedCategories(ctx); ok {
		return cached, nil
	}

	categories, err := s.categoryRepo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := toCategoryResponses(categories)
	s.setCachedCategories(ctx, responses)
	return responses, nil
}

func (s *Service) GetCategory(id uint) (*CategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := toCategoryResponse(category)
	return &response, nil
}

func (s *Service) CreateCategory(req CategoryRequest) (*CategoryResponse, error) {
	category, err := domainCategory.NewCategory(req.Name, req.Description, req.Image)
	if err != nil {
		return nil, err
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, err
	}

	s.invalidateCategoriesCache()
	response := toCategoryResponse(category)
	return &response, nil
}

func (s *Service) UpdateCategory(id uint, req CategoryUpdateRequest) (*CategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	name := category.Name
	description := category.Description
	image := category.Image

	if req.Name != nil {
		name = *req.Name
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.Image != nil {
		image = *req.Image
	}

	if err := category.Update(name, description, image); err != nil {
		return nil, err
	}

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, err
	}

	s.invalidateCategoriesCache()
	response := toCategoryResponse(category)
	return &response, nil
}

func (s *Service) DeleteCategory(id uint) error {
	if _, err := s.categoryRepo.GetByID(id); err != nil {
		return err
	}

	if err := s.categoryRepo.Delete(id); err != nil {
		return err
	}

	s.invalidateCategoriesCache()
	return nil
}

func (s *Service) getCachedCategories(ctx context.Context) ([]CategoryResponse, bool) {
	if s.cache == nil {
		return nil, false
	}

	cached, err := s.cache.Get(ctx, categoriesCacheKey)
	if err != nil {
		return nil, false
	}

	var categories []CategoryResponse
	if err := json.Unmarshal([]byte(cached), &categories); err != nil {
		return nil, false
	}

	return categories, true
}

func (s *Service) setCachedCategories(ctx context.Context, categories []CategoryResponse) {
	if s.cache == nil {
		return
	}

	data, err := json.Marshal(categories)
	if err != nil {
		return
	}

	_ = s.cache.Set(ctx, categoriesCacheKey, data, categoriesCacheTTL)
}

func (s *Service) invalidateCategoriesCache() {
	if s.cache == nil {
		return
	}

	_ = s.cache.Delete(context.Background(), categoriesCacheKey)
}
