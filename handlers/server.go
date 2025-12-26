package handlers

import (
	"net/http"

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

// RouterBuilder builds an HTTP router with the ogen server and optional middleware
type RouterBuilder struct {
	handler     api.Handler
	middlewares []func(http.Handler) http.Handler
}

// NewRouter creates a new RouterBuilder
func NewRouter(handler api.Handler) *RouterBuilder {
	return &RouterBuilder{
		handler:     handler,
		middlewares: make([]func(http.Handler) http.Handler, 0),
	}
}

// WithMiddleware adds a middleware to the router
func (b *RouterBuilder) WithMiddleware(mw func(http.Handler) http.Handler) *RouterBuilder {
	b.middlewares = append(b.middlewares, mw)
	return b
}

// Build creates the HTTP handler with all configured middleware
func (b *RouterBuilder) Build() (http.Handler, error) {
	server, err := NewServer(b.handler)
	if err != nil {
		return nil, err
	}

	var handler http.Handler = server

	// Apply middlewares in reverse order so they execute in the order they were added
	for i := len(b.middlewares) - 1; i >= 0; i-- {
		handler = b.middlewares[i](handler)
	}

	return handler, nil
}
