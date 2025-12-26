package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	api "github.com/sunfmin/shadcn-admin-go/api/gen/admin"
)

// ogenEncoder is an interface for ogen-generated types that have MarshalJSON
type ogenEncoder interface {
	MarshalJSON() ([]byte, error)
}

// newAPIRequest creates an HTTP request with an ogen API struct as body
func newAPIRequest(t *testing.T, method, path string, body ogenEncoder) *http.Request {
	t.Helper()
	var bodyReader io.Reader
	var contentLength int64

	if body != nil {
		data, err := body.MarshalJSON()
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(data)
		contentLength = int64(len(data))
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = contentLength
	}
	return req
}

func TestAuthLogin(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	defer truncateTables(db, "users")

	testUser := createTestUser(t, db, "test@test.com", "password123", "admin")

	handler := createTestHandler(db)
	server, err := api.NewServer(handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	testCases := []struct {
		name       string
		request    *api.LoginRequest
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid credentials",
			request: &api.LoginRequest{
				Email:    "test@test.com",
				Password: "password123",
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "invalid password",
			request: &api.LoginRequest{
				Email:    "test@test.com",
				Password: "wrongpassword",
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name: "user not found",
			request: &api.LoginRequest{
				Email:    "notfound@test.com",
				Password: "password123",
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := newAPIRequest(t, "POST", "/auth/login", tc.request)
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.wantStatus, rec.Code, rec.Body.String())
			}

			if !tc.wantErr {
				var response api.LoginResponse
				respBody, _ := io.ReadAll(rec.Body)
				if err := json.Unmarshal(respBody, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// Build expected from input/fixtures only - do NOT copy from actual
				expected := api.LoginResponse{
					User: api.AuthUser{
						Email: testUser.Email,
						Role:  []string{testUser.Role},
					},
				}

				// Use IgnoreFields for generated fields
				opts := cmp.Options{
					cmpopts.IgnoreFields(api.AuthUser{}, "AccountNo", "Exp"),
					cmpopts.IgnoreFields(api.LoginResponse{}, "AccessToken"),
				}
				if diff := cmp.Diff(expected, response, opts...); diff != "" {
					t.Errorf("LoginResponse mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestAuthLogout(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	handler := createTestHandler(db)
	server, err := api.NewServer(handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestAuthGetCurrentUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	handler := createTestHandler(db)
	server, err := api.NewServer(handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/auth/me", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestUserCRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	defer truncateTables(db, "users")

	handler := createTestHandler(db)
	server, err := api.NewServer(handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	var createdUserID string

	t.Run("create user", func(t *testing.T) {
		createReq := &api.CreateUserRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@test.com",
			Role:      api.UserRoleAdmin,
		}
		req := newAPIRequest(t, "POST", "/users", createReq)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusCreated, rec.Code, rec.Body.String())
		}

		var response api.User
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		createdUserID = response.ID.String()

		// Build expected from input only - do NOT copy from actual
		expected := api.User{
			FirstName: "John",
			LastName:  "Doe",
			Username:  "john.doe",
			Email:     "john.doe@test.com",
			Status:    api.UserStatusActive,
			Role:      api.UserRoleAdmin,
		}

		// Use IgnoreFields for generated fields (ID, timestamps)
		opts := cmp.Options{
			cmpopts.IgnoreFields(api.User{}, "ID", "CreatedAt", "UpdatedAt"),
		}
		if diff := cmp.Diff(expected, response, opts...); diff != "" {
			t.Errorf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("list users", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.UserListResponse
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if len(response.Data) != 1 {
			t.Errorf("Expected 1 user, got %d", len(response.Data))
		}
	})

	t.Run("get user - found", func(t *testing.T) {
		if createdUserID == "" {
			t.Skip("No user created")
		}
		req := httptest.NewRequest("GET", "/users/"+createdUserID, nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("get user - not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/00000000-0000-0000-0000-000000000000", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("delete user - not found", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/users/00000000-0000-0000-0000-000000000000", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}

func TestUserInvite(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	defer truncateTables(db, "users")

	handler := createTestHandler(db)
	server, err := api.NewServer(handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	inviteReq := &api.InviteUserRequest{
		Email: "invited@test.com",
		Role:  api.UserRoleCashier,
	}
	req := newAPIRequest(t, "POST", "/users/invite", inviteReq)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var response api.User
	respBody, _ := io.ReadAll(rec.Body)
	json.Unmarshal(respBody, &response)

	// Build expected from input only - do NOT copy from actual
	// Service sets default names for invited users
	expected := api.User{
		FirstName: "Invited", // Default for invited users
		LastName:  "User",    // Default for invited users
		Email:     inviteReq.Email,
		Status:    api.UserStatusInvited,
		Role:      inviteReq.Role,
	}

	// Use IgnoreFields for generated fields (ID, Username, timestamps)
	opts := cmp.Options{
		cmpopts.IgnoreFields(api.User{}, "ID", "Username", "CreatedAt", "UpdatedAt"),
	}
	if diff := cmp.Diff(expected, response, opts...); diff != "" {
		t.Errorf("User mismatch (-want +got):\n%s", diff)
	}
}

func TestTaskCRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	defer truncateTables(db, "tasks")

	handler := createTestHandler(db)
	server, err := api.NewServer(handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	var createdTaskID string

	t.Run("create task", func(t *testing.T) {
		createReq := &api.CreateTaskRequest{
			Title:    "Test Task",
			Status:   api.TaskStatusTodo,
			Label:    api.TaskLabelFeature,
			Priority: api.TaskPriorityHigh,
		}
		req := newAPIRequest(t, "POST", "/tasks", createReq)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusCreated, rec.Code, rec.Body.String())
		}

		var response api.Task
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		createdTaskID = response.ID

		// Build expected from input only - do NOT copy from actual
		expected := api.Task{
			Title:    createReq.Title,
			Status:   createReq.Status,
			Label:    createReq.Label,
			Priority: createReq.Priority,
		}

		// Use IgnoreFields for generated fields (ID, timestamps)
		opts := cmp.Options{
			cmpopts.IgnoreFields(api.Task{}, "ID", "CreatedAt", "UpdatedAt"),
		}
		if diff := cmp.Diff(expected, response, opts...); diff != "" {
			t.Errorf("Task mismatch (-want +got):\n%s", diff)
		}

		// Verify ID format
		if len(response.ID) < 5 || response.ID[:5] != "TASK-" {
			t.Errorf("Expected task ID to start with 'TASK-', got '%s'", response.ID)
		}
	})

	t.Run("list tasks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.TaskListResponse
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if len(response.Data) != 1 {
			t.Errorf("Expected 1 task, got %d", len(response.Data))
		}
	})

	t.Run("get task - found", func(t *testing.T) {
		if createdTaskID == "" {
			t.Skip("No task created")
		}
		req := httptest.NewRequest("GET", "/tasks/"+createdTaskID, nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("get task - not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks/TASK-9999", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("update task", func(t *testing.T) {
		if createdTaskID == "" {
			t.Skip("No task created")
		}
		updateReq := &api.UpdateTaskRequest{
			Title:  api.NewOptString("Updated Task Title"),
			Status: api.NewOptTaskStatus(api.TaskStatusInProgress),
		}
		req := newAPIRequest(t, "PUT", "/tasks/"+createdTaskID, updateReq)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var response api.Task
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		// Build expected from update request - do NOT copy from actual
		expected := api.Task{
			ID:     createdTaskID,
			Title:  updateReq.Title.Value,
			Status: updateReq.Status.Value,
			// Label and Priority not updated, so we need to check them from original create
			Label:    api.TaskLabelFeature,
			Priority: api.TaskPriorityHigh,
		}

		// Use IgnoreFields for generated fields (timestamps)
		opts := cmp.Options{
			cmpopts.IgnoreFields(api.Task{}, "CreatedAt", "UpdatedAt"),
		}
		if diff := cmp.Diff(expected, response, opts...); diff != "" {
			t.Errorf("Task mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("delete task", func(t *testing.T) {
		if createdTaskID == "" {
			t.Skip("No task created")
		}
		req := httptest.NewRequest("DELETE", "/tasks/"+createdTaskID, nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
	})
}

func TestAppOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	defer truncateTables(db, "apps")

	createTestApp(t, db, "slack", "Slack", "Team messaging", false)
	createTestApp(t, db, "github", "GitHub", "Code hosting", true)

	handler := createTestHandler(db)
	server, err := api.NewServer(handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	t.Run("list apps", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/apps", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.AppListResponse
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if len(response.Data) != 2 {
			t.Errorf("Expected 2 apps, got %d", len(response.Data))
		}
	})

	t.Run("connect app", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/apps/slack/connect", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var response api.App
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		// Build expected from fixture only - do NOT copy from actual
		expected := api.App{
			ID:        "slack",
			Name:      "Slack",
			Desc:      "Team messaging",
			Connected: true, // After connect
		}

		// Use IgnoreFields for optional fields
		opts := cmp.Options{
			cmpopts.IgnoreFields(api.App{}, "Logo"),
		}
		if diff := cmp.Diff(expected, response, opts...); diff != "" {
			t.Errorf("App mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("disconnect app", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/apps/github/disconnect", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		var response api.App
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		// Build expected from fixture only - do NOT copy from actual
		expected := api.App{
			ID:        "github",
			Name:      "GitHub",
			Desc:      "Code hosting",
			Connected: false, // After disconnect
		}

		// Use IgnoreFields for optional fields
		opts := cmp.Options{
			cmpopts.IgnoreFields(api.App{}, "Logo"),
		}
		if diff := cmp.Diff(expected, response, opts...); diff != "" {
			t.Errorf("App mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestChatOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	defer truncateTables(db, "chat_messages", "chat_conversations")

	chat := createTestChat(t, db, "chat-1", "johndoe", "John Doe")
	createTestChatMessage(t, db, chat.ID, "johndoe", "Hello!")

	handler := createTestHandler(db)
	server, err := api.NewServer(handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	t.Run("list chats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/chats", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.ChatListResponse
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if len(response.Data) != 1 {
			t.Errorf("Expected 1 chat, got %d", len(response.Data))
		}
	})

	t.Run("get chat - found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/chats/chat-1", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.ChatConversation
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if len(response.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(response.Messages))
		}
	})

	t.Run("get chat - not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/chats/nonexistent", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("send message", func(t *testing.T) {
		sendReq := &api.SendMessageRequest{
			Message: "Hello back!",
		}
		req := newAPIRequest(t, "POST", "/chats/chat-1/messages", sendReq)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusCreated, rec.Code, rec.Body.String())
		}

		var response api.ChatMessage
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		// Build expected from input only - do NOT copy from actual
		expected := api.ChatMessage{
			Message: sendReq.Message,
		}

		// Use IgnoreFields for generated fields (Sender, Timestamp)
		opts := cmp.Options{
			cmpopts.IgnoreFields(api.ChatMessage{}, "Sender", "Timestamp"),
		}
		if diff := cmp.Diff(expected, response, opts...); diff != "" {
			t.Errorf("ChatMessage mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestDashboardEndpoints(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	handler := createTestHandler(db)
	server, err := api.NewServer(handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	t.Run("get dashboard stats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/dashboard/stats", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.DashboardStats
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if !response.TotalRevenue.Value.IsSet() {
			t.Error("Expected TotalRevenue.Value to be set")
		}
	})

	t.Run("get dashboard overview", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/dashboard/overview", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.DashboardOverview
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if len(response.Data) != 12 {
			t.Errorf("Expected 12 months of data, got %d", len(response.Data))
		}
	})

	t.Run("get recent sales", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/dashboard/recent-sales", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.RecentSalesResponse
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if len(response.Data) != 5 {
			t.Errorf("Expected 5 recent sales, got %d", len(response.Data))
		}
	})
}

func TestUserFilters(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	defer truncateTables(db, "users")

	createTestUser(t, db, "active1@test.com", "pass123", "admin")
	createTestUser(t, db, "active2@test.com", "pass123", "cashier")

	ctx := context.Background()
	db.WithContext(ctx).Exec("UPDATE users SET status = 'inactive' WHERE email = 'active2@test.com'")

	handler := createTestHandler(db)
	server, err := api.NewServer(handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	t.Run("filter by status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users?status=active", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.UserListResponse
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if len(response.Data) != 1 {
			t.Errorf("Expected 1 active user, got %d", len(response.Data))
		}
	})

	t.Run("filter by role", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users?role=admin", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.UserListResponse
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if len(response.Data) != 1 {
			t.Errorf("Expected 1 admin user, got %d", len(response.Data))
		}
	})
}

func TestTaskFilters(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	defer truncateTables(db, "tasks")

	createTestTask(t, db, "TASK-0001", "Bug fix", "todo", "bug", "high")
	createTestTask(t, db, "TASK-0002", "Feature request", "in progress", "feature", "medium")
	createTestTask(t, db, "TASK-0003", "Documentation", "done", "documentation", "low")

	handler := createTestHandler(db)
	server, err := api.NewServer(handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	t.Run("filter by status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks?status=todo", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.TaskListResponse
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if len(response.Data) != 1 {
			t.Errorf("Expected 1 todo task, got %d", len(response.Data))
		}
	})

	t.Run("filter by priority", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks?priority=high", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.TaskListResponse
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if len(response.Data) != 1 {
			t.Errorf("Expected 1 high priority task, got %d", len(response.Data))
		}
	})

	t.Run("search filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks?filter=Bug", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response api.TaskListResponse
		respBody, _ := io.ReadAll(rec.Body)
		json.Unmarshal(respBody, &response)

		if len(response.Data) != 1 {
			t.Errorf("Expected 1 task matching 'Bug', got %d", len(response.Data))
		}
	})
}
