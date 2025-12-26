package services

import (
	"context"

	api "github.com/sunfmin/shadcn-admin-go/api/gen/admin"
)

// ErrMissingRequired is returned when a required service is not configured
var ErrMissingRequired = ErrUnauthorized

// OgenHandler implements the ogen-generated api.Handler interface
// It delegates to the underlying domain services
type OgenHandler struct {
	authService      AuthService
	userService      UserService
	taskService      TaskService
	appService       AppService
	chatService      ChatService
	dashboardService DashboardService
}

// OgenHandlerBuilder builds an OgenHandler with optional services
type OgenHandlerBuilder struct {
	authService      AuthService
	userService      UserService
	taskService      TaskService
	appService       AppService
	chatService      ChatService
	dashboardService DashboardService
}

// NewOgenHandler creates a new OgenHandler builder
func NewOgenHandler() *OgenHandlerBuilder {
	return &OgenHandlerBuilder{}
}

// WithAuthService adds auth service
func (b *OgenHandlerBuilder) WithAuthService(svc AuthService) *OgenHandlerBuilder {
	b.authService = svc
	return b
}

// WithUserService adds user service
func (b *OgenHandlerBuilder) WithUserService(svc UserService) *OgenHandlerBuilder {
	b.userService = svc
	return b
}

// WithTaskService adds task service
func (b *OgenHandlerBuilder) WithTaskService(svc TaskService) *OgenHandlerBuilder {
	b.taskService = svc
	return b
}

// WithAppService adds app service
func (b *OgenHandlerBuilder) WithAppService(svc AppService) *OgenHandlerBuilder {
	b.appService = svc
	return b
}

// WithChatService adds chat service
func (b *OgenHandlerBuilder) WithChatService(svc ChatService) *OgenHandlerBuilder {
	b.chatService = svc
	return b
}

// WithDashboardService adds dashboard service
func (b *OgenHandlerBuilder) WithDashboardService(svc DashboardService) *OgenHandlerBuilder {
	b.dashboardService = svc
	return b
}

// Build creates the OgenHandler instance
func (b *OgenHandlerBuilder) Build() *OgenHandler {
	return &OgenHandler{
		authService:      b.authService,
		userService:      b.userService,
		taskService:      b.taskService,
		appService:       b.appService,
		chatService:      b.chatService,
		dashboardService: b.dashboardService,
	}
}

// Ensure OgenHandler implements api.Handler
var _ api.Handler = (*OgenHandler)(nil)

// ============================================================================
// Auth Operations - delegate to AuthService
// ============================================================================

// Login implements api.Handler
func (h *OgenHandler) Login(ctx context.Context, req *api.LoginRequest) (api.LoginRes, error) {
	if h.authService == nil {
		return nil, ErrMissingRequired
	}
	return h.authService.Login(ctx, req)
}

// Logout implements api.Handler
func (h *OgenHandler) Logout(ctx context.Context) error {
	if h.authService == nil {
		return ErrMissingRequired
	}
	return h.authService.Logout(ctx)
}

// GetCurrentUser implements api.Handler
func (h *OgenHandler) GetCurrentUser(ctx context.Context) (api.GetCurrentUserRes, error) {
	if h.authService == nil {
		return nil, ErrMissingRequired
	}
	return h.authService.GetCurrentUser(ctx)
}

// ============================================================================
// User Operations - delegate to UserService
// ============================================================================

// ListUsers implements api.Handler
func (h *OgenHandler) ListUsers(ctx context.Context, params api.ListUsersParams) (*api.UserListResponse, error) {
	if h.userService == nil {
		return nil, ErrMissingRequired
	}
	return h.userService.List(ctx, params)
}

// CreateUser implements api.Handler
func (h *OgenHandler) CreateUser(ctx context.Context, req *api.CreateUserRequest) (*api.User, error) {
	if h.userService == nil {
		return nil, ErrMissingRequired
	}
	return h.userService.Create(ctx, req)
}

// GetUser implements api.Handler
func (h *OgenHandler) GetUser(ctx context.Context, params api.GetUserParams) (api.GetUserRes, error) {
	if h.userService == nil {
		return nil, ErrMissingRequired
	}
	return h.userService.Get(ctx, params)
}

// UpdateUser implements api.Handler
func (h *OgenHandler) UpdateUser(ctx context.Context, req *api.UpdateUserRequest, params api.UpdateUserParams) (api.UpdateUserRes, error) {
	if h.userService == nil {
		return nil, ErrMissingRequired
	}
	return h.userService.Update(ctx, req, params)
}

// DeleteUser implements api.Handler
func (h *OgenHandler) DeleteUser(ctx context.Context, params api.DeleteUserParams) (api.DeleteUserRes, error) {
	if h.userService == nil {
		return nil, ErrMissingRequired
	}
	return h.userService.Delete(ctx, params)
}

// InviteUser implements api.Handler
func (h *OgenHandler) InviteUser(ctx context.Context, req *api.InviteUserRequest) (*api.User, error) {
	if h.userService == nil {
		return nil, ErrMissingRequired
	}
	return h.userService.Invite(ctx, req)
}

// ============================================================================
// Task Operations - delegate to TaskService
// ============================================================================

// ListTasks implements api.Handler
func (h *OgenHandler) ListTasks(ctx context.Context, params api.ListTasksParams) (*api.TaskListResponse, error) {
	if h.taskService == nil {
		return nil, ErrMissingRequired
	}
	return h.taskService.List(ctx, params)
}

// CreateTask implements api.Handler
func (h *OgenHandler) CreateTask(ctx context.Context, req *api.CreateTaskRequest) (*api.Task, error) {
	if h.taskService == nil {
		return nil, ErrMissingRequired
	}
	return h.taskService.Create(ctx, req)
}

// GetTask implements api.Handler
func (h *OgenHandler) GetTask(ctx context.Context, params api.GetTaskParams) (api.GetTaskRes, error) {
	if h.taskService == nil {
		return nil, ErrMissingRequired
	}
	return h.taskService.Get(ctx, params)
}

// UpdateTask implements api.Handler
func (h *OgenHandler) UpdateTask(ctx context.Context, req *api.UpdateTaskRequest, params api.UpdateTaskParams) (api.UpdateTaskRes, error) {
	if h.taskService == nil {
		return nil, ErrMissingRequired
	}
	return h.taskService.Update(ctx, req, params)
}

// DeleteTask implements api.Handler
func (h *OgenHandler) DeleteTask(ctx context.Context, params api.DeleteTaskParams) (api.DeleteTaskRes, error) {
	if h.taskService == nil {
		return nil, ErrMissingRequired
	}
	return h.taskService.Delete(ctx, params)
}

// ============================================================================
// App Operations - delegate to AppService
// ============================================================================

// ListApps implements api.Handler
func (h *OgenHandler) ListApps(ctx context.Context, params api.ListAppsParams) (*api.AppListResponse, error) {
	if h.appService == nil {
		return nil, ErrMissingRequired
	}
	return h.appService.List(ctx, params)
}

// ConnectApp implements api.Handler
func (h *OgenHandler) ConnectApp(ctx context.Context, params api.ConnectAppParams) (*api.App, error) {
	if h.appService == nil {
		return nil, ErrMissingRequired
	}
	return h.appService.Connect(ctx, params)
}

// DisconnectApp implements api.Handler
func (h *OgenHandler) DisconnectApp(ctx context.Context, params api.DisconnectAppParams) (*api.App, error) {
	if h.appService == nil {
		return nil, ErrMissingRequired
	}
	return h.appService.Disconnect(ctx, params)
}

// ============================================================================
// Chat Operations - delegate to ChatService
// ============================================================================

// ListChats implements api.Handler
func (h *OgenHandler) ListChats(ctx context.Context, params api.ListChatsParams) (*api.ChatListResponse, error) {
	if h.chatService == nil {
		return nil, ErrMissingRequired
	}
	return h.chatService.List(ctx, params)
}

// GetChat implements api.Handler
func (h *OgenHandler) GetChat(ctx context.Context, params api.GetChatParams) (api.GetChatRes, error) {
	if h.chatService == nil {
		return nil, ErrMissingRequired
	}
	return h.chatService.Get(ctx, params)
}

// SendMessage implements api.Handler
func (h *OgenHandler) SendMessage(ctx context.Context, req *api.SendMessageRequest, params api.SendMessageParams) (*api.ChatMessage, error) {
	if h.chatService == nil {
		return nil, ErrMissingRequired
	}
	return h.chatService.SendMessage(ctx, req, params)
}

// ============================================================================
// Dashboard Operations - delegate to DashboardService
// ============================================================================

// GetDashboardStats implements api.Handler
func (h *OgenHandler) GetDashboardStats(ctx context.Context) (*api.DashboardStats, error) {
	if h.dashboardService == nil {
		return nil, ErrMissingRequired
	}
	return h.dashboardService.GetStats(ctx)
}

// GetDashboardOverview implements api.Handler
func (h *OgenHandler) GetDashboardOverview(ctx context.Context) (*api.DashboardOverview, error) {
	if h.dashboardService == nil {
		return nil, ErrMissingRequired
	}
	return h.dashboardService.GetOverview(ctx)
}

// GetRecentSales implements api.Handler
func (h *OgenHandler) GetRecentSales(ctx context.Context) (*api.RecentSalesResponse, error) {
	if h.dashboardService == nil {
		return nil, ErrMissingRequired
	}
	return h.dashboardService.GetRecentSales(ctx)
}
