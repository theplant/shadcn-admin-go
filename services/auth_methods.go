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

// Login implements api.Handler.
func (s *AdminService) Login(ctx context.Context, req *api.LoginRequest) (api.LoginRes, error) {
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

// Logout implements api.Handler.
func (s *AdminService) Logout(ctx context.Context) error {
	return nil
}

// GetCurrentUser implements api.Handler.
func (s *AdminService) GetCurrentUser(ctx context.Context) (api.GetCurrentUserRes, error) {
	return &api.GetCurrentUserUnauthorized{}, nil
}

// generateAccessToken generates a simple access token
func generateAccessToken(userID string, exp int) string {
	return fmt.Sprintf("token_%s_%d", userID, exp)
}
