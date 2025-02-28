package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/samcharles93/cinea/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SchedulerRepository interface {
	ListTasks(ctx context.Context) ([]*entity.ScheduledTask, error)
	GetTaskByID(ctx context.Context, id uint) (*entity.ScheduledTask, error)
	AddTask(ctx context.Context, task *entity.ScheduledTask) error
	UpdateTask(ctx context.Context, task *entity.ScheduledTask) error
	RemoveTask(ctx context.Context, id uint) error

	ToggleTaskStatus(ctx context.Context, id uint) error
	// RestoreDefaultTasks(ctx context.Context) error // old method, to be removed
}

type schedulerRepository struct {
	db *gorm.DB
}

func NewSchedulerRepository(db *gorm.DB) SchedulerRepository {
	return &schedulerRepository{
		db: db,
	}
}

func (r *schedulerRepository) ListTasks(ctx context.Context) ([]*entity.ScheduledTask, error) {
	var tasks []*entity.ScheduledTask
	result := r.db.WithContext(ctx).Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", result.Error)
	}
	return tasks, nil
}

func (r *schedulerRepository) GetTaskByID(ctx context.Context, id uint) (*entity.ScheduledTask, error) {
	var task entity.ScheduledTask
	result := r.db.WithContext(ctx).First(&task, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task by ID: %w", result.Error)
	}
	return &task, nil
}

func (r *schedulerRepository) AddTask(ctx context.Context, task *entity.ScheduledTask) error {
	result := r.db.WithContext(ctx).Clauses(clause.Returning{}).Create(task)
	if result.Error != nil {
		return fmt.Errorf("failed to add task: %w", result.Error)
	}
	return nil
}

func (r *schedulerRepository) UpdateTask(ctx context.Context, task *entity.ScheduledTask) error {
	result := r.db.WithContext(ctx).Save(task)
	if result.Error != nil {
		return fmt.Errorf("failed to update task: %w", result.Error)
	}
	return nil
}

func (r *schedulerRepository) RemoveTask(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.ScheduledTask{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to remove task: %w", result.Error)
	}
	return nil
}

func (r *schedulerRepository) ToggleTaskStatus(ctx context.Context, id uint) error {
	task, err := r.GetTaskByID(ctx, id)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("task with ID %d not found", id)
	}

	// Toggle the Enabled status.  This assumes you have an 'Enabled' field.
	result := r.db.WithContext(ctx).Model(task).Update("enabled", !task.Enabled)
	if result.Error != nil {
		return fmt.Errorf("failed to toggle task status: %w", result.Error)
	}
	return nil
}
