package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FirstName   string    `gorm:"not null"`
	LastName    string    `gorm:"not null"`
	Username    string    `gorm:"uniqueIndex;not null"`
	Email       string    `gorm:"uniqueIndex;not null"`
	Password    string    `gorm:"not null"`
	PhoneNumber string
	Status      string    `gorm:"not null;default:'active'"`
	Role        string    `gorm:"not null;default:'cashier'"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// Task represents a task in the system
type Task struct {
	ID          string    `gorm:"primaryKey"`
	Title       string    `gorm:"not null"`
	Status      string    `gorm:"not null;default:'todo'"`
	Label       string    `gorm:"not null;default:'feature'"`
	Priority    string    `gorm:"not null;default:'medium'"`
	Assignee    string
	Description string
	DueDate     *time.Time
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// BeforeCreate generates a task ID in format TASK-XXXX
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		var count int64
		tx.Model(&Task{}).Count(&count)
		t.ID = generateTaskID(int(count) + 1)
	}
	return nil
}

func generateTaskID(num int) string {
	return "TASK-" + padNumber(num, 4)
}

func padNumber(num, width int) string {
	s := ""
	for i := 0; i < width; i++ {
		s = string('0'+num%10) + s
		num /= 10
	}
	return s
}

// App represents an app integration
type App struct {
	ID        string `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Desc      string `gorm:"not null"`
	Logo      string
	Connected bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// ChatUser represents a chat user
type ChatUser struct {
	ID       string `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex;not null"`
	FullName string `gorm:"not null"`
	Title    string
	Profile  string
}

// ChatMessage represents a message in a chat
type ChatMessage struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	ChatID    string    `gorm:"index;not null"`
	Sender    string    `gorm:"not null"`
	Message   string    `gorm:"not null"`
	Timestamp time.Time `gorm:"not null"`
}

// ChatConversation represents a chat conversation
type ChatConversation struct {
	ID       string        `gorm:"primaryKey"`
	Username string        `gorm:"not null"`
	FullName string        `gorm:"not null"`
	Title    string
	Profile  string
	Messages []ChatMessage `gorm:"foreignKey:ChatID;references:ID"`
}
