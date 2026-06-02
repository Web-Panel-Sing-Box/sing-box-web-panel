package settings

import (
	"context"
	"fmt"
)

// Repository reads and writes panel settings.
type Repository interface {
	All(ctx context.Context) (map[string]string, error)
	Set(ctx context.Context, key, value string) error
}

// ConfigTrigger schedules a debounced sing-box config rebuild.
type ConfigTrigger interface {
	Trigger()
}

// Service provides read/update access to panel settings, stored as plain key/value
// rows. It validates well-known keys and triggers a config rebuild when settings
// that affect sing-box generation are modified.
type Service struct {
	repo    Repository
	trigger ConfigTrigger
}

func New(repo Repository, trigger ConfigTrigger) *Service {
	return &Service{repo: repo, trigger: trigger}
}

// All returns every persisted setting as a key → value map.
func (s *Service) All(ctx context.Context) (map[string]string, error) {
	m, err := s.repo.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings all: %w", err)
	}
	if m == nil {
		m = make(map[string]string)
	}
	return m, nil
}

// Patch upserts the provided key/value pairs. Unknown keys are silently accepted;
// the caller is expected to validate on its own side. Returns the set of keys that
// were actually changed and might require a config rebuild.
func (s *Service) Patch(ctx context.Context, patch map[string]string) ([]string, error) {
	var changed []string
	for k, v := range patch {
		if v == "" {
			continue
		}
		if err := s.repo.Set(ctx, k, v); err != nil {
			return nil, fmt.Errorf("settings patch %q: %w", k, err)
		}
		changed = append(changed, k)
	}
	return changed, nil
}

// PatchAndRebuild saves settings and triggers a sing-box config rebuild when any
// generator-affecting key is among the changed ones.
func (s *Service) PatchAndRebuild(ctx context.Context, patch map[string]string) error {
	changed, err := s.Patch(ctx, patch)
	if err != nil {
		return err
	}
	if len(changed) > 0 && s.trigger != nil {
		s.trigger.Trigger()
	}
	return nil
}
