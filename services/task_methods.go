package services

import (
	"context"
	"errors"
	"fmt"

	api "github.com/sunfmin/shadcn-admin-go/api/gen/admin"
	"github.com/sunfmin/shadcn-admin-go/internal/models"
	"gorm.io/gorm"
)

// ListTasks implements api.Handler.
func (s *AdminService) ListTasks(ctx context.Context, params api.ListTasksParams) (*api.TaskListResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	page := params.Page.Or(1)
	pageSize := params.PageSize.Or(10)
	offset := (page - 1) * pageSize

	query := s.db.WithContext(ctx).Model(&models.Task{})

	// Apply filters
	if len(params.Status) > 0 {
		statuses := make([]string, len(params.Status))
		for i, s := range params.Status {
			statuses[i] = string(s)
		}
		query = query.Where("status IN ?", statuses)
	}

	if len(params.Priority) > 0 {
		priorities := make([]string, len(params.Priority))
		for i, p := range params.Priority {
			priorities[i] = string(p)
		}
		query = query.Where("priority IN ?", priorities)
	}

	if filter, ok := params.Filter.Get(); ok && filter != "" {
		query = query.Where("title ILIKE ? OR id ILIKE ?", "%"+filter+"%", "%"+filter+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count tasks: %w", err)
	}

	var tasks []models.Task
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	data := make([]api.Task, len(tasks))
	for i, t := range tasks {
		data[i] = taskToAPI(t)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &api.TaskListResponse{
		Data: data,
		Meta: api.PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}, nil
}

// CreateTask implements api.Handler.
func (s *AdminService) CreateTask(ctx context.Context, req *api.CreateTaskRequest) (*api.Task, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	task := &models.Task{
		Title:    req.Title,
		Status:   string(req.Status),
		Label:    string(req.Label),
		Priority: string(req.Priority),
	}

	if assignee, ok := req.Assignee.Get(); ok {
		task.Assignee = assignee
	}
	if desc, ok := req.Description.Get(); ok {
		task.Description = desc
	}
	if dueDate, ok := req.DueDate.Get(); ok {
		task.DueDate = &dueDate
	}

	if err := s.db.WithContext(ctx).Create(task).Error; err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	result := taskToAPI(*task)
	return &result, nil
}

// GetTask implements api.Handler.
func (s *AdminService) GetTask(ctx context.Context, params api.GetTaskParams) (api.GetTaskRes, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var task models.Task
	if err := s.db.WithContext(ctx).Where("id = ?", params.TaskId).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &api.ErrorResponse{Message: ErrTaskNotFound.Error()}, nil
		}
		return nil, fmt.Errorf("get task: %w", err)
	}

	result := taskToAPI(task)
	return &result, nil
}

// UpdateTask implements api.Handler.
func (s *AdminService) UpdateTask(ctx context.Context, req *api.UpdateTaskRequest, params api.UpdateTaskParams) (api.UpdateTaskRes, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var task models.Task
	if err := s.db.WithContext(ctx).Where("id = ?", params.TaskId).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &api.UpdateTaskNotFound{}, nil
		}
		return nil, fmt.Errorf("get task: %w", err)
	}

	updates := make(map[string]interface{})

	if title, ok := req.Title.Get(); ok {
		updates["title"] = title
	}
	if status, ok := req.Status.Get(); ok {
		updates["status"] = string(status)
	}
	if label, ok := req.Label.Get(); ok {
		updates["label"] = string(label)
	}
	if priority, ok := req.Priority.Get(); ok {
		updates["priority"] = string(priority)
	}
	if assignee, ok := req.Assignee.Get(); ok {
		updates["assignee"] = assignee
	}
	if desc, ok := req.Description.Get(); ok {
		updates["description"] = desc
	}
	if dueDate, ok := req.DueDate.Get(); ok {
		updates["due_date"] = dueDate
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&task).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update task: %w", err)
		}
	}

	// Reload task
	if err := s.db.WithContext(ctx).First(&task, "id = ?", params.TaskId).Error; err != nil {
		return nil, fmt.Errorf("reload task: %w", err)
	}

	result := taskToAPI(task)
	return &result, nil
}

// DeleteTask implements api.Handler.
func (s *AdminService) DeleteTask(ctx context.Context, params api.DeleteTaskParams) (api.DeleteTaskRes, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result := s.db.WithContext(ctx).Where("id = ?", params.TaskId).Delete(&models.Task{})
	if result.Error != nil {
		return nil, fmt.Errorf("delete task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return &api.DeleteTaskNotFound{}, nil
	}

	return &api.DeleteTaskNoContent{}, nil
}

// taskToAPI converts a models.Task to api.Task
func taskToAPI(t models.Task) api.Task {
	result := api.Task{
		ID:        t.ID,
		Title:     t.Title,
		Status:    api.TaskStatus(t.Status),
		Label:     api.TaskLabel(t.Label),
		Priority:  api.TaskPriority(t.Priority),
		CreatedAt: api.NewOptDateTime(t.CreatedAt),
		UpdatedAt: api.NewOptDateTime(t.UpdatedAt),
	}

	if t.Assignee != "" {
		result.Assignee = api.NewOptString(t.Assignee)
	}
	if t.Description != "" {
		result.Description = api.NewOptString(t.Description)
	}
	if t.DueDate != nil {
		result.DueDate = api.NewOptDateTime(*t.DueDate)
	}

	return result
}
