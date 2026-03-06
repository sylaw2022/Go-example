package domain

import "errors"

// User represents a user entity in our domain
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var ErrUserNotFound = errors.New("user not found")
