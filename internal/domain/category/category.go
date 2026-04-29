package category

import (
	"errors"
	"time"
)

type Category struct {
	ID          uint
	Name        string
	Description string
	Image       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewCategory(name, description, image string) (*Category, error) {
	if name == "" {
		return nil, errors.New("category name cannot be empty")
	}

	return &Category{
		Name:        name,
		Description: description,
		Image:       image,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (c *Category) Update(name, description, image string) error {
	if name == "" {
		return errors.New("category name cannot be empty")
	}

	c.Name = name
	c.Description = description
	c.Image = image
	c.UpdatedAt = time.Now()
	return nil
}
