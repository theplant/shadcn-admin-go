package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	api "github.com/sunfmin/shadcn-admin-go/api/gen/admin"
	"github.com/sunfmin/shadcn-admin-go/internal/models"
	"gorm.io/gorm"
)

// ChatService interface for chat operations
type ChatService interface {
	List(ctx context.Context, params api.ListChatsParams) (*api.ChatListResponse, error)
	Get(ctx context.Context, params api.GetChatParams) (api.GetChatRes, error)
	SendMessage(ctx context.Context, req *api.SendMessageRequest, params api.SendMessageParams) (*api.ChatMessage, error)
}

// chatServiceImpl implements ChatService
type chatServiceImpl struct {
	db *gorm.DB
}

// chatServiceBuilder is the builder for ChatService
type chatServiceBuilder struct {
	db *gorm.DB
}

// NewChatService creates a new ChatService builder
func NewChatService(db *gorm.DB) *chatServiceBuilder {
	return &chatServiceBuilder{db: db}
}

// Build creates the ChatService
func (b *chatServiceBuilder) Build() ChatService {
	return &chatServiceImpl{db: b.db}
}

// List implements ChatService
func (s *chatServiceImpl) List(ctx context.Context, params api.ListChatsParams) (*api.ChatListResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	query := s.db.WithContext(ctx).Model(&models.ChatConversation{}).Preload("Messages")

	// Apply search filter
	if search, ok := params.Search.Get(); ok && search != "" {
		query = query.Where("full_name ILIKE ? OR username ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var conversations []models.ChatConversation
	if err := query.Find(&conversations).Error; err != nil {
		return nil, fmt.Errorf("list chats: %w", err)
	}

	data := make([]api.ChatConversation, len(conversations))
	for i, c := range conversations {
		data[i] = chatConversationToAPI(c)
	}

	return &api.ChatListResponse{
		Data: data,
	}, nil
}

// Get implements ChatService
func (s *chatServiceImpl) Get(ctx context.Context, params api.GetChatParams) (api.GetChatRes, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var conversation models.ChatConversation
	if err := s.db.WithContext(ctx).Preload("Messages").Where("id = ?", params.ChatId).First(&conversation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &api.GetChatNotFound{}, nil
		}
		return nil, fmt.Errorf("get chat: %w", err)
	}

	result := chatConversationToAPI(conversation)
	return &result, nil
}

// SendMessage implements ChatService
func (s *chatServiceImpl) SendMessage(ctx context.Context, req *api.SendMessageRequest, params api.SendMessageParams) (*api.ChatMessage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check if chat exists
	var conversation models.ChatConversation
	if err := s.db.WithContext(ctx).Where("id = ?", params.ChatId).First(&conversation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatNotFound
		}
		return nil, fmt.Errorf("get chat: %w", err)
	}

	message := &models.ChatMessage{
		ChatID:    params.ChatId,
		Sender:    "current_user", // In production, get from auth context
		Message:   req.Message,
		Timestamp: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(message).Error; err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}

	return &api.ChatMessage{
		Sender:    message.Sender,
		Message:   message.Message,
		Timestamp: message.Timestamp,
	}, nil
}

// chatConversationToAPI converts a models.ChatConversation to api.ChatConversation
func chatConversationToAPI(c models.ChatConversation) api.ChatConversation {
	messages := make([]api.ChatMessage, len(c.Messages))
	for i, m := range c.Messages {
		messages[i] = api.ChatMessage{
			Sender:    m.Sender,
			Message:   m.Message,
			Timestamp: m.Timestamp,
		}
	}

	result := api.ChatConversation{
		ID:       c.ID,
		Username: c.Username,
		FullName: c.FullName,
		Messages: messages,
	}

	if c.Title != "" {
		result.Title = api.NewOptString(c.Title)
	}
	if c.Profile != "" {
		result.Profile = api.NewOptString(c.Profile)
	}

	return result
}
