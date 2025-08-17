package models

import (
	"time"
	"errors"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"`
	IsAdmin   bool      `json:"is_admin" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUser(username, email, password string, isAdmin bool) (*User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	return &User{
		Username:  username,
		Email:     email,
		Password:  password,
		IsAdmin:   isAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (u *User) IsValidForLogin() bool {
	return u.Username != "" && u.Password != ""
}

func (u *User) UpdatePassword(newPassword string) error {
	if newPassword == "" {
		return errors.New("password cannot be empty")
	}
	u.Password = newPassword
	u.UpdatedAt = time.Now()
	return nil
}