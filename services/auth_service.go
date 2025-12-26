package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	api "github.com/sunfmin/shadcn-admin-go/api/gen/admin"
	"github.com/sunfmin/shadcn-admin-go/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService interface for authentication operations
type AuthService interface {
	Login(ctx context.Context, req *api.LoginRequest) (api.LoginRes, error)
	Logout(ctx context.Context) error
	GetCurrentUser(ctx context.Context) (api.GetCurrentUserRes, error)
}

// authServiceImpl implements AuthService
type authServiceImpl struct {
	db *gorm.DB
}

// authServiceBuilder is the builder for AuthService
type authServiceBuilder struct {
	db *gorm.DB
}

// NewAuthService creates a new AuthService builder
func NewAuthService(db *gorm.DB) *authServiceBuilder {
	return &authServiceBuilder{db: db}
}

// Build creates the AuthService
func (b *authServiceBuilder) Build() AuthService {
	return &authServiceImpl{db: b.db}
}

// Login implements AuthService
func (s *authServiceImpl) Login(ctx context.Context, req *api.LoginRequest) (api.LoginRes, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &api.ErrorResponse{Message: ErrInvalidCredentials.Error()}, nil
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return &api.ErrorResponse{Message: ErrInvalidCredentials.Error()}, nil
	}

	// Generate token expiry (24 hours from now)
	exp := int(time.Now().Add(24 * time.Hour).Unix())

	return &api.LoginResponse{
		User: api.AuthUser{
			AccountNo: user.ID.String(),
			Email:     user.Email,
			Role:      []string{user.Role},
			Exp:       exp,
		},
		AccessToken: generateAccessToken(user.ID.String(), exp),
	}, nil
}

// Logout implements AuthService
func (s *authServiceImpl) Logout(ctx context.Context) error {
	return nil
}

// GetCurrentUser implements AuthService
func (s *authServiceImpl) GetCurrentUser(ctx context.Context) (api.GetCurrentUserRes, error) {
	return &api.GetCurrentUserUnauthorized{}, nil
}

// generateAccessToken generates a simple access token
func generateAccessToken(userID string, exp int) string {
	return fmt.Sprintf("token_%s_%d", userID, exp)
}
