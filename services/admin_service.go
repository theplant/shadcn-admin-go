package services

import (
	"context"

	api "github.com/sunfmin/shadcn-admin-go/api/gen/admin"
	"gorm.io/gorm"
)

// AdminService implements api.Handler interface
type AdminService struct {
	db *gorm.DB
}

// Ensure AdminService implements the generated Handler interface
var _ api.Handler = (*AdminService)(nil)

// adminServiceBuilder is the builder for AdminService
type adminServiceBuilder struct {
	db *gorm.DB
}

// NewAdminService creates a new AdminService builder
func NewAdminService(db *gorm.DB) *adminServiceBuilder {
	return &adminServiceBuilder{db: db}
}

// Build creates the AdminService
func (b *adminServiceBuilder) Build() *AdminService {
	return &AdminService{db: b.db}
}

// NewError creates an error response
func NewError(ctx context.Context, err error) *api.ErrorResponse {
	return &api.ErrorResponse{
		Message: err.Error(),
	}
}
