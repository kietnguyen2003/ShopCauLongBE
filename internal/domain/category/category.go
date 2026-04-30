package category

import (
	"errors"
	"time"
)

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

type Category struct {
	ID          uint
	Name        string
	Description string
	Image       string
	Status      string
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
		Status:      StatusActive,
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

func (c *Category) Deactivate() {
	c.Status = StatusInactive
	c.UpdatedAt = time.Now()
}
