package category

import domainCategory "kafka-order-demo/backend/internal/domain/category"

type CategoryRepository interface {
	Create(category *domainCategory.Category) error
	GetByID(id uint) (*domainCategory.Category, error)
	GetAll() ([]*domainCategory.Category, error)
	Update(category *domainCategory.Category) error
	Delete(id uint) error
}
