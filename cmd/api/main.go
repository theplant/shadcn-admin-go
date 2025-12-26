package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sunfmin/shadcn-admin-go/handlers"
	"github.com/sunfmin/shadcn-admin-go/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/shadcn_admin?sslmode=disable"
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := services.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Configure error details visibility (hide in production)
	if os.Getenv("HIDE_ERROR_DETAILS") == "true" {
		handlers.SetHideErrorDetails(true)
	}

	// Create individual domain services
	authService := services.NewAuthService(db).Build()
	userService := services.NewUserService(db).Build()
	taskService := services.NewTaskService(db).Build()
	appService := services.NewAppService(db).Build()
	chatService := services.NewChatService(db).Build()
	dashboardService := services.NewDashboardService().Build()

	// Create OgenHandler with all services
	handler := services.NewOgenHandler().
		WithAuthService(authService).
		WithUserService(userService).
		WithTaskService(taskService).
		WithAppService(appService).
		WithChatService(chatService).
		WithDashboardService(dashboardService).
		Build()

	// Create router with ogen server
	router, err := handlers.NewRouter(handler).Build()
	if err != nil {
		log.Fatalf("Failed to create router: %v", err)
	}

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
