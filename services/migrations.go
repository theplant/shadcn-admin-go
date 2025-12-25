package services

import (
	"github.com/sunfmin/shadcn-admin-go/internal/models"
	"gorm.io/gorm"
)

// AutoMigrate runs database migrations for all models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.App{},
		&models.ChatUser{},
		&models.ChatConversation{},
		&models.ChatMessage{},
	)
}
