package category

import domainCategory "kafka-order-demo/backend/internal/domain/category"

type Service struct {
	categoryRepo CategoryRepository
}

func NewService(categoryRepo CategoryRepository) *Service {
	return &Service{
		categoryRepo: categoryRepo,
	}
}

func (s *Service) GetCategories() ([]CategoryResponse, error) {
	categories, err := s.categoryRepo.GetAll()
	if err != nil {
		return nil, err
	}

	return toCategoryResponses(categories), nil
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

	response := toCategoryResponse(category)
	return &response, nil
}

func (s *Service) UpdateCategory(id uint, req CategoryRequest) (*CategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := category.Update(req.Name, req.Description, req.Image); err != nil {
		return nil, err
	}

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, err
	}

	response := toCategoryResponse(category)
	return &response, nil
}

func (s *Service) DeleteCategory(id uint) error {
	if _, err := s.categoryRepo.GetByID(id); err != nil {
		return err
	}

	return s.categoryRepo.Delete(id)
}
