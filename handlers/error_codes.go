package handlers

import (
	"net/http"

	"github.com/sunfmin/shadcn-admin-go/services"
)

// ErrorCode represents a structured error with HTTP status and service error mapping
type ErrorCode struct {
	Code       string
	Message    string
	HTTPStatus int
	ServiceErr error // Maps to service sentinel error (nil for HTTP-only errors)
}

// errorCodes is the singleton containing all error codes
var errorCodes = struct {
	// Service errors (mapped from services.Err*)
	UserNotFound       ErrorCode
	TaskNotFound       ErrorCode
	AppNotFound        ErrorCode
	ChatNotFound       ErrorCode
	InvalidCredentials ErrorCode
	Unauthorized       ErrorCode
	DuplicateEmail     ErrorCode
	DuplicateUsername  ErrorCode

	// HTTP-only errors (no service mapping)
	BadRequest       ErrorCode
	InternalError    ErrorCode
	RequestCancelled ErrorCode
	RequestTimeout   ErrorCode
}{
	// Service errors
	UserNotFound: ErrorCode{
		Code:       "USER_NOT_FOUND",
		Message:    "User not found",
		HTTPStatus: http.StatusNotFound,
		ServiceErr: services.ErrUserNotFound,
	},
	TaskNotFound: ErrorCode{
		Code:       "TASK_NOT_FOUND",
		Message:    "Task not found",
		HTTPStatus: http.StatusNotFound,
		ServiceErr: services.ErrTaskNotFound,
	},
	AppNotFound: ErrorCode{
		Code:       "APP_NOT_FOUND",
		Message:    "App not found",
		HTTPStatus: http.StatusNotFound,
		ServiceErr: services.ErrAppNotFound,
	},
	ChatNotFound: ErrorCode{
		Code:       "CHAT_NOT_FOUND",
		Message:    "Chat not found",
		HTTPStatus: http.StatusNotFound,
		ServiceErr: services.ErrChatNotFound,
	},
	InvalidCredentials: ErrorCode{
		Code:       "INVALID_CREDENTIALS",
		Message:    "Invalid email or password",
		HTTPStatus: http.StatusUnauthorized,
		ServiceErr: services.ErrInvalidCredentials,
	},
	Unauthorized: ErrorCode{
		Code:       "UNAUTHORIZED",
		Message:    "Authentication required",
		HTTPStatus: http.StatusUnauthorized,
		ServiceErr: services.ErrUnauthorized,
	},
	DuplicateEmail: ErrorCode{
		Code:       "DUPLICATE_EMAIL",
		Message:    "Email already exists",
		HTTPStatus: http.StatusConflict,
		ServiceErr: services.ErrDuplicateEmail,
	},
	DuplicateUsername: ErrorCode{
		Code:       "DUPLICATE_USERNAME",
		Message:    "Username already exists",
		HTTPStatus: http.StatusConflict,
		ServiceErr: services.ErrDuplicateUsername,
	},

	// HTTP-only errors
	BadRequest: ErrorCode{
		Code:       "BAD_REQUEST",
		Message:    "Invalid request",
		HTTPStatus: http.StatusBadRequest,
	},
	InternalError: ErrorCode{
		Code:       "INTERNAL_ERROR",
		Message:    "An internal error occurred",
		HTTPStatus: http.StatusInternalServerError,
	},
	RequestCancelled: ErrorCode{
		Code:       "REQUEST_CANCELLED",
		Message:    "Request was cancelled",
		HTTPStatus: 499, // Client Closed Request
	},
	RequestTimeout: ErrorCode{
		Code:       "REQUEST_TIMEOUT",
		Message:    "Request timed out",
		HTTPStatus: http.StatusGatewayTimeout,
	},
}

// Errors provides access to the error codes singleton
var Errors = errorCodes

// AllErrors returns all error codes for iteration (used by mapServiceError)
func AllErrors() []ErrorCode {
	return []ErrorCode{
		errorCodes.UserNotFound,
		errorCodes.TaskNotFound,
		errorCodes.AppNotFound,
		errorCodes.ChatNotFound,
		errorCodes.InvalidCredentials,
		errorCodes.Unauthorized,
		errorCodes.DuplicateEmail,
		errorCodes.DuplicateUsername,
		errorCodes.BadRequest,
		errorCodes.InternalError,
		errorCodes.RequestCancelled,
		errorCodes.RequestTimeout,
	}
}
