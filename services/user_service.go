package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	api "github.com/sunfmin/shadcn-admin-go/api/gen/admin"
	"github.com/sunfmin/shadcn-admin-go/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService interface for user operations
type UserService interface {
	List(ctx context.Context, params api.ListUsersParams) (*api.UserListResponse, error)
	Create(ctx context.Context, req *api.CreateUserRequest) (*api.User, error)
	Get(ctx context.Context, params api.GetUserParams) (api.GetUserRes, error)
	Update(ctx context.Context, req *api.UpdateUserRequest, params api.UpdateUserParams) (api.UpdateUserRes, error)
	Delete(ctx context.Context, params api.DeleteUserParams) (api.DeleteUserRes, error)
	Invite(ctx context.Context, req *api.InviteUserRequest) (*api.User, error)
}

// userServiceImpl implements UserService
type userServiceImpl struct {
	db *gorm.DB
}

// userServiceBuilder is the builder for UserService
type userServiceBuilder struct {
	db *gorm.DB
}

// NewUserService creates a new UserService builder
func NewUserService(db *gorm.DB) *userServiceBuilder {
	return &userServiceBuilder{db: db}
}

// Build creates the UserService
func (b *userServiceBuilder) Build() UserService {
	return &userServiceImpl{db: b.db}
}

// List implements UserService
func (s *userServiceImpl) List(ctx context.Context, params api.ListUsersParams) (*api.UserListResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	page := params.Page.Or(1)
	pageSize := params.PageSize.Or(10)
	offset := (page - 1) * pageSize

	query := s.db.WithContext(ctx).Model(&models.User{})

	// Apply filters
	if len(params.Status) > 0 {
		statuses := make([]string, len(params.Status))
		for i, st := range params.Status {
			statuses[i] = string(st)
		}
		query = query.Where("status IN ?", statuses)
	}

	if len(params.Role) > 0 {
		roles := make([]string, len(params.Role))
		for i, r := range params.Role {
			roles[i] = string(r)
		}
		query = query.Where("role IN ?", roles)
	}

	if username, ok := params.Username.Get(); ok && username != "" {
		query = query.Where("username ILIKE ?", "%"+username+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	var users []models.User
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	data := make([]api.User, len(users))
	for i, u := range users {
		data[i] = userToAPI(u)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &api.UserListResponse{
		Data: data,
		Meta: api.PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}, nil
}

// Create implements UserService
func (s *userServiceImpl) Create(ctx context.Context, req *api.CreateUserRequest) (*api.User, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Generate username from email
	username := strings.Split(req.Email, "@")[0]

	// Generate a default password (in production, send email to set password)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("changeme123"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      string(req.Role),
		Status:    "active",
	}

	if phone, ok := req.PhoneNumber.Get(); ok {
		user.PhoneNumber = phone
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		if isDuplicateKeyError(err) {
			return nil, fmt.Errorf("create user: %w", ErrDuplicateEmail)
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	result := userToAPI(*user)
	return &result, nil
}

// Get implements UserService
func (s *userServiceImpl) Get(ctx context.Context, params api.GetUserParams) (api.GetUserRes, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", params.UserId).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &api.GetUserNotFound{}, nil
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	result := userToAPI(user)
	return &result, nil
}

// Update implements UserService
func (s *userServiceImpl) Update(ctx context.Context, req *api.UpdateUserRequest, params api.UpdateUserParams) (api.UpdateUserRes, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", params.UserId).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &api.UpdateUserNotFound{}, nil
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	updates := make(map[string]interface{})

	if firstName, ok := req.FirstName.Get(); ok {
		updates["first_name"] = firstName
	}
	if lastName, ok := req.LastName.Get(); ok {
		updates["last_name"] = lastName
	}
	if email, ok := req.Email.Get(); ok {
		updates["email"] = email
	}
	if phone, ok := req.PhoneNumber.Get(); ok {
		updates["phone_number"] = phone
	}
	if status, ok := req.Status.Get(); ok {
		updates["status"] = string(status)
	}
	if role, ok := req.Role.Get(); ok {
		updates["role"] = string(role)
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update user: %w", err)
		}
	}

	// Reload user
	if err := s.db.WithContext(ctx).First(&user, "id = ?", params.UserId).Error; err != nil {
		return nil, fmt.Errorf("reload user: %w", err)
	}

	result := userToAPI(user)
	return &result, nil
}

// Delete implements UserService
func (s *userServiceImpl) Delete(ctx context.Context, params api.DeleteUserParams) (api.DeleteUserRes, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result := s.db.WithContext(ctx).Where("id = ?", params.UserId).Delete(&models.User{})
	if result.Error != nil {
		return nil, fmt.Errorf("delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return &api.DeleteUserNotFound{}, nil
	}

	return &api.DeleteUserNoContent{}, nil
}

// Invite implements UserService
func (s *userServiceImpl) Invite(ctx context.Context, req *api.InviteUserRequest) (*api.User, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Generate username from email
	username := strings.Split(req.Email, "@")[0]

	// Generate a temporary password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("invited123"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		FirstName: "Invited",
		LastName:  "User",
		Username:  username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      string(req.Role),
		Status:    "invited",
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		if isDuplicateKeyError(err) {
			return nil, fmt.Errorf("invite user: %w", ErrDuplicateEmail)
		}
		return nil, fmt.Errorf("invite user: %w", err)
	}

	result := userToAPI(*user)
	return &result, nil
}

// userToAPI converts a models.User to api.User
func userToAPI(u models.User) api.User {
	result := api.User{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Username:  u.Username,
		Email:     u.Email,
		Status:    api.UserStatus(u.Status),
		Role:      api.UserRole(u.Role),
		CreatedAt: api.NewOptDateTime(u.CreatedAt),
		UpdatedAt: api.NewOptDateTime(u.UpdatedAt),
	}

	if u.PhoneNumber != "" {
		result.PhoneNumber = api.NewOptString(u.PhoneNumber)
	}

	return result
}

// isDuplicateKeyError checks if the error is a duplicate key error
func isDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}
