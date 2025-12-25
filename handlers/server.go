package handlers

import (
	api "github.com/sunfmin/shadcn-admin-go/api/gen/admin"
)

// NewServer creates an ogen server with proper error handling configured
// This wrapper ensures all service errors are mapped to user-friendly HTTP responses
func NewServer(h api.Handler) (*api.Server, error) {
	return api.NewServer(
		h,
		api.WithErrorHandler(OgenErrorHandler),
	)
}
