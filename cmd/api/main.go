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

	// Create service (implements ogen Handler interface)
	service := services.NewAdminService(db).Build()

	// Create server via handlers package (includes ErrorHandler)
	server, err := handlers.NewServer(service)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, server); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
