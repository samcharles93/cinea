package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/reugn/go-quartz/quartz"
	"github.com/samcharles93/cinea/internal/entity"
	"github.com/samcharles93/cinea/internal/logger"
	"github.com/samcharles93/cinea/internal/persistence"
)

type TaskExecutor interface {
	Execute(ctx context.Context, config string) error
	Description() string
}

type Scheduler interface {
	Start(ctx context.Context)
	Shutdown(ctx context.Context)
	RegisterTask(taskType string, executor TaskExecutor)
	LoadTasks(ctx context.Context) error
}

type scheduler struct {
	scheduler quartz.Scheduler
	appLogger logger.Logger
	tasks     map[string]TaskExecutor
	repo      persistence.SchedulerRepository
}

func NewScheduler(appLogger logger.Logger, repo persistence.SchedulerRepository) (Scheduler, error) {
	sched, err := quartz.NewStdScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to initialise scheduler instance: %w", err)
	}

	return &scheduler{
		scheduler: sched,
		appLogger: appLogger,
		tasks:     make(map[string]TaskExecutor),
		repo:      repo,
	}, nil
}

func (s *scheduler) Start(ctx context.Context) {
	s.appLogger.Info().Msg("Starting the scheduler")
	s.scheduler.Start(ctx)
}

func (s *scheduler) Shutdown(ctx context.Context) {
	s.appLogger.Info().Msg("Shutting down the scheduler")
	s.scheduler.Stop()
}

func (s *scheduler) RegisterTask(taskType string, executor TaskExecutor) {
	s.tasks[taskType] = executor
}

func (s *scheduler) LoadTasks(ctx context.Context) error {
	tasks, err := s.repo.ListTasks(ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if !task.Enabled {
			continue
		}

		executor, ok := s.tasks[task.Type]
		if !ok {
			s.appLogger.Warn().Str("type", task.Type).Str("task", task.Name).Msg("Unknown task type, skipping")
			continue
		}

		if err := s.scheduleTask(task, executor); err != nil {
			s.appLogger.Error().Err(err).Str("task", task.Name).Msg("Failed to schedule task")
		}
	}

	return nil
}

type taskWrapper struct {
	task      *entity.ScheduledTask
	executor  TaskExecutor
	appLogger logger.Logger
	repo      persistence.SchedulerRepository
}

func (s *scheduler) scheduleTask(task *entity.ScheduledTask, executor TaskExecutor) error {
	job := &taskWrapper{
		task:      task,
		executor:  executor,
		appLogger: s.appLogger,
		repo:      s.repo,
	}

	intervalDuration, err := time.ParseDuration(task.Interval)
	if err != nil {
		return fmt.Errorf("invalid interval '%s' for task '%s': %w", task.Interval, task.Name, err)
	}

	// Create the trigger based on task interval
	trigger := quartz.NewSimpleTrigger(intervalDuration)
	jobDetail := quartz.NewJobDetail(job, quartz.NewJobKey(task.Name))

	// Schedule the task
	return s.scheduler.ScheduleJob(jobDetail, trigger)
}

func (w *taskWrapper) Execute(ctx context.Context) error {
	w.appLogger.Info().Str("task", w.task.Name).Msg("Task starting")

	w.task.Status = entity.StatusRunning
	w.task.LastRun = time.Now()
	if err := w.repo.UpdateTask(ctx, w.task); err != nil {
		w.appLogger.Error().Err(err).Str("task", w.task.Name).Msg("failed to update task status")
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Execute the task
	err := w.executor.Execute(ctx, w.task.Config)

	// Update task status based on the result
	w.task.Status = entity.StatusIdle
	if err != nil {
		w.task.Status = entity.StatusFailed
		w.appLogger.Error().Err(err).Str("task", w.task.Name).Msg("Task execution failed")
	}

	// Calculate and set the next run time
	intervalDuration, parseErr := time.ParseDuration(w.task.Interval)
	if parseErr != nil {
		w.appLogger.Error().Err(parseErr).Str("task", w.task.Name).Str("interval", w.task.Interval).Msg("Failed to parse task interval")
		return fmt.Errorf("failed to parse task interval '%s': %w", w.task.Interval, parseErr)
	}
	w.task.NextRun = time.Now().Add(intervalDuration)

	// Update task in database
	if err := w.repo.UpdateTask(ctx, w.task); err != nil {
		w.appLogger.Error().Err(err).Str("task", w.task.Name).Msg("Failed to update task status after execution")
		return fmt.Errorf("failed to update task status after execution: %w", err)
	}

	return err
}

func (w *taskWrapper) Description() string {
	return w.executor.Description()
}
