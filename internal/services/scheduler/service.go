package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"sing-box-web-panel/internal/domain"
)

type Repo interface {
	Create(ctx context.Context, t *domain.ScheduledTask) error
	GetByID(ctx context.Context, id int64) (*domain.ScheduledTask, error)
	List(ctx context.Context) ([]domain.ScheduledTask, error)
	ListEnabled(ctx context.Context) ([]domain.ScheduledTask, error)
	Update(ctx context.Context, t *domain.ScheduledTask) error
	Delete(ctx context.Context, id int64) error
	SetLastRun(ctx context.Context, id int64, at time.Time) error
	SetNextRun(ctx context.Context, id int64, at time.Time) error
}

type ActionHandler func(ctx context.Context, params json.RawMessage) error

type Service struct {
	repo     Repo
	log      *slog.Logger
	cron     *cron.Cron
	handlers map[string]ActionHandler
	entries  map[int64]cron.EntryID
	mu       sync.Mutex
}

func New(repo Repo, log *slog.Logger) *Service {
	return &Service{
		repo:     repo,
		log:      log,
		cron:     cron.New(cron.WithSeconds()),
		handlers: make(map[string]ActionHandler),
		entries:  make(map[int64]cron.EntryID),
	}
}

func (s *Service) RegisterAction(action string, h ActionHandler) {
	s.handlers[action] = h
}

func (s *Service) Start() error {
	s.reloadTasks(context.Background())
	s.cron.Start()
	return nil
}

func (s *Service) Stop() context.Context {
	return s.cron.Stop()
}

func (s *Service) List(ctx context.Context) ([]domain.ScheduledTask, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, t *domain.ScheduledTask) error {
	if err := s.repo.Create(ctx, t); err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	if t.Enabled {
		s.scheduleTask(t)
	}
	return nil
}

func (s *Service) Update(ctx context.Context, t *domain.ScheduledTask) error {
	existing, err := s.repo.GetByID(ctx, t.ID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	if err := s.repo.Update(ctx, t); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	s.mu.Lock()
	if eid, ok := s.entries[t.ID]; ok {
		s.cron.Remove(eid)
		delete(s.entries, t.ID)
	}
	if existing.Enabled {
		s.cron.Remove(s.entries[existing.ID])
	}
	if t.Enabled {
		s.scheduleTask(t)
	}
	s.mu.Unlock()

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	s.mu.Lock()
	if eid, ok := s.entries[id]; ok {
		s.cron.Remove(eid)
		delete(s.entries, id)
	}
	s.mu.Unlock()

	return s.repo.Delete(ctx, id)
}

func (s *Service) reloadTasks(ctx context.Context) {
	tasks, err := s.repo.ListEnabled(ctx)
	if err != nil {
		s.log.Error("reload scheduled tasks", slog.String("error", err.Error()))
		return
	}

	s.mu.Lock()
	for id, eid := range s.entries {
		s.cron.Remove(eid)
		delete(s.entries, id)
	}

	for i := range tasks {
		s.scheduleTask(&tasks[i])
	}
	s.mu.Unlock()

	s.log.Info("scheduled tasks loaded", slog.Int("count", len(tasks)))
}

func (s *Service) scheduleTask(t *domain.ScheduledTask) {
	handler, ok := s.handlers[t.Action]
	if !ok {
		s.log.Warn("unknown action for scheduled task",
			slog.String("name", t.Name),
			slog.String("action", t.Action))
		return
	}

	var params json.RawMessage
	if t.ParamsJSON != "" {
		params = json.RawMessage(t.ParamsJSON)
	}

	task := t
	eid, err := s.cron.AddFunc(t.CronExpr, func() {
		s.log.Info("running scheduled task",
			slog.String("name", task.Name),
			slog.String("action", task.Action))
		ctx := context.Background()
		if err := handler(ctx, params); err != nil {
			s.log.Error("scheduled task failed",
				slog.String("name", task.Name),
				slog.String("error", err.Error()))
		}
		now := time.Now()
		s.repo.SetLastRun(ctx, task.ID, now)
	})
	if err != nil {
		s.log.Error("schedule task",
			slog.String("name", t.Name),
			slog.String("error", err.Error()))
		return
	}

	s.entries[t.ID] = eid
}
