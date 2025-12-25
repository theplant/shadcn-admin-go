package services

import "errors"

// Sentinel errors for the admin service
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrTaskNotFound        = errors.New("task not found")
	ErrAppNotFound         = errors.New("app not found")
	ErrChatNotFound        = errors.New("chat not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrDuplicateEmail      = errors.New("email already exists")
	ErrDuplicateUsername   = errors.New("username already exists")
)
