package category

import (
	"context"
	"time"

	domainCategory "kafka-order-demo/backend/internal/domain/category"
)

type CategoryRepository interface {
	Create(category *domainCategory.Category) error
	GetByID(id uint) (*domainCategory.Category, error)
	GetAll() ([]*domainCategory.Category, error)
	Update(category *domainCategory.Category) error
	Delete(id uint) error
}

type CacheStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}
