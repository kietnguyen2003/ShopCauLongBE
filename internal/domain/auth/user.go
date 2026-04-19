package auth

import (
	"errors"
	"time"
)

// User represents the user domain entity
type User struct {
	ID        uint
	Username  string
	Email     string
	Password  string
	IsAdmin   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser creates a new user with validation
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

// IsValidForLogin validates user credentials
func (u *User) IsValidForLogin() bool {
	return u.Username != "" && u.Password != ""
}

// UpdatePassword updates user password
func (u *User) UpdatePassword(newPassword string) error {
	if newPassword == "" {
		return errors.New("password cannot be empty")
	}
	u.Password = newPassword
	u.UpdatedAt = time.Now()
	return nil
}
