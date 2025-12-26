package services

import (
	"context"
	"errors"
	"fmt"

	api "github.com/sunfmin/shadcn-admin-go/api/gen/admin"
	"github.com/sunfmin/shadcn-admin-go/internal/models"
	"gorm.io/gorm"
)

// AppService interface for app operations
type AppService interface {
	List(ctx context.Context, params api.ListAppsParams) (*api.AppListResponse, error)
	Connect(ctx context.Context, params api.ConnectAppParams) (*api.App, error)
	Disconnect(ctx context.Context, params api.DisconnectAppParams) (*api.App, error)
}

// appServiceImpl implements AppService
type appServiceImpl struct {
	db *gorm.DB
}

// appServiceBuilder is the builder for AppService
type appServiceBuilder struct {
	db *gorm.DB
}

// NewAppService creates a new AppService builder
func NewAppService(db *gorm.DB) *appServiceBuilder {
	return &appServiceBuilder{db: db}
}

// Build creates the AppService
func (b *appServiceBuilder) Build() AppService {
	return &appServiceImpl{db: b.db}
}

// List implements AppService
func (s *appServiceImpl) List(ctx context.Context, params api.ListAppsParams) (*api.AppListResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	query := s.db.WithContext(ctx).Model(&models.App{})

	// Apply type filter
	if appType, ok := params.Type.Get(); ok {
		switch appType {
		case api.ListAppsTypeConnected:
			query = query.Where("connected = ?", true)
		case api.ListAppsTypeNotConnected:
			query = query.Where("connected = ?", false)
		// api.ListAppsTypeAll - no filter needed
		}
	}

	// Apply name filter
	if filter, ok := params.Filter.Get(); ok && filter != "" {
		query = query.Where("name ILIKE ?", "%"+filter+"%")
	}

	// Apply sort
	if sort, ok := params.Sort.Get(); ok {
		switch sort {
		case api.ListAppsSortAsc:
			query = query.Order("name ASC")
		case api.ListAppsSortDesc:
			query = query.Order("name DESC")
		}
	}

	var apps []models.App
	if err := query.Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("list apps: %w", err)
	}

	data := make([]api.App, len(apps))
	for i, a := range apps {
		data[i] = appToAPI(a)
	}

	return &api.AppListResponse{
		Data: data,
	}, nil
}

// Connect implements AppService
func (s *appServiceImpl) Connect(ctx context.Context, params api.ConnectAppParams) (*api.App, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var app models.App
	if err := s.db.WithContext(ctx).Where("id = ?", params.AppId).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAppNotFound
		}
		return nil, fmt.Errorf("get app: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&app).Update("connected", true).Error; err != nil {
		return nil, fmt.Errorf("connect app: %w", err)
	}

	app.Connected = true
	result := appToAPI(app)
	return &result, nil
}

// Disconnect implements AppService
func (s *appServiceImpl) Disconnect(ctx context.Context, params api.DisconnectAppParams) (*api.App, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var app models.App
	if err := s.db.WithContext(ctx).Where("id = ?", params.AppId).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAppNotFound
		}
		return nil, fmt.Errorf("get app: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&app).Update("connected", false).Error; err != nil {
		return nil, fmt.Errorf("disconnect app: %w", err)
	}

	app.Connected = false
	result := appToAPI(app)
	return &result, nil
}

// appToAPI converts a models.App to api.App
func appToAPI(a models.App) api.App {
	result := api.App{
		ID:        a.ID,
		Name:      a.Name,
		Desc:      a.Desc,
		Connected: a.Connected,
	}

	if a.Logo != "" {
		result.Logo = api.NewOptString(a.Logo)
	}

	return result
}
