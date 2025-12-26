package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	api "github.com/sunfmin/shadcn-admin-go/api/gen/admin"
	"github.com/sunfmin/shadcn-admin-go/internal/models"
	"github.com/sunfmin/shadcn-admin-go/services"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates a PostgreSQL test container and returns a GORM DB connection
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	cleanup := func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		cleanup()
		t.Fatalf("Failed to get connection string: %v", err)
	}

	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		cleanup()
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := services.AutoMigrate(db); err != nil {
		cleanup()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db, cleanup
}

// truncateTables truncates the specified tables in reverse order
func truncateTables(db *gorm.DB, tables ...string) {
	for i := len(tables) - 1; i >= 0; i-- {
		db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tables[i]))
	}
}

// createTestUser creates a test user with hashed password
func createTestUser(t *testing.T, db *gorm.DB, email, password, role string) *models.User {
	t.Helper()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Extract username from email (part before @)
	username := email
	if idx := len(email) - len("@test.com"); idx > 0 && email[idx:] == "@test.com" {
		username = email[:idx]
	}

	user := &models.User{
		FirstName: "Test",
		LastName:  "User",
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		Role:      role,
		Status:    "active",
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// createTestTask creates a test task
func createTestTask(t *testing.T, db *gorm.DB, id, title, status, label, priority string) *models.Task {
	t.Helper()

	task := &models.Task{
		ID:       id,
		Title:    title,
		Status:   status,
		Label:    label,
		Priority: priority,
	}

	if err := db.Create(task).Error; err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	return task
}

// createTestApp creates a test app
func createTestApp(t *testing.T, db *gorm.DB, id, name, desc string, connected bool) *models.App {
	t.Helper()

	app := &models.App{
		ID:        id,
		Name:      name,
		Desc:      desc,
		Connected: connected,
	}

	if err := db.Create(app).Error; err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}

	return app
}

// createTestChat creates a test chat conversation
func createTestChat(t *testing.T, db *gorm.DB, id, username, fullName string) *models.ChatConversation {
	t.Helper()

	chat := &models.ChatConversation{
		ID:       id,
		Username: username,
		FullName: fullName,
	}

	if err := db.Create(chat).Error; err != nil {
		t.Fatalf("Failed to create test chat: %v", err)
	}

	return chat
}

// createTestChatMessage creates a test chat message
func createTestChatMessage(t *testing.T, db *gorm.DB, chatID, sender, message string) *models.ChatMessage {
	t.Helper()

	msg := &models.ChatMessage{
		ChatID:    chatID,
		Sender:    sender,
		Message:   message,
		Timestamp: time.Now(),
	}

	if err := db.Create(msg).Error; err != nil {
		t.Fatalf("Failed to create test chat message: %v", err)
	}

	return msg
}

// createTestHandler creates an OgenHandler with all services for testing
func createTestHandler(db *gorm.DB) api.Handler {
	authService := services.NewAuthService(db).Build()
	userService := services.NewUserService(db).Build()
	taskService := services.NewTaskService(db).Build()
	appService := services.NewAppService(db).Build()
	chatService := services.NewChatService(db).Build()
	dashboardService := services.NewDashboardService().Build()

	return services.NewOgenHandler().
		WithAuthService(authService).
		WithUserService(userService).
		WithTaskService(taskService).
		WithAppService(appService).
		WithChatService(chatService).
		WithDashboardService(dashboardService).
		Build()
}
