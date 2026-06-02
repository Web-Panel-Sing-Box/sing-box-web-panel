package settings_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	svcsettings "sing-box-web-panel/internal/services/settings"
)

type fakeRepo struct {
	mu   sync.RWMutex
	data map[string]string
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{data: make(map[string]string)}
}

func (r *fakeRepo) All(_ context.Context) (map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]string, len(r.data))
	for k, v := range r.data {
		out[k] = v
	}
	return out, nil
}

func (r *fakeRepo) Set(_ context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[key] = value
	return nil
}

type fakeTrigger struct {
	mu    sync.RWMutex
	count int
}

func (t *fakeTrigger) Trigger() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.count++
}

func (t *fakeTrigger) triggered() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.count
}

func newService(repo *fakeRepo, trigger svcsettings.ConfigTrigger) *svcsettings.Service {
	return svcsettings.New(repo, trigger)
}

func TestAll_Empty(t *testing.T) {
	svc := newService(newFakeRepo(), nil)
	m, err := svc.All(context.Background())
	if err != nil {
		t.Fatalf("all: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("want empty map, got %v", m)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	repo := newFakeRepo()
	repo.Set(context.Background(), "key", "val")
	svc := newService(repo, nil)

	m, err := svc.All(context.Background())
	if err != nil {
		t.Fatalf("all: %v", err)
	}
	if m["key"] != "val" {
		t.Errorf("key = %q, want val", m["key"])
	}
	m["key"] = "mutated"
	m2, _ := svc.All(context.Background())
	if m2["key"] != "val" {
		t.Error("All should return a copy of the internal map")
	}
}

func TestPatch_SavesValues(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(repo, nil)

	patch := map[string]string{"a": "1", "b": "2"}
	changed, err := svc.Patch(context.Background(), patch)
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	if len(changed) != 2 {
		t.Errorf("want 2 changed, got %d", len(changed))
	}

	m, _ := svc.All(context.Background())
	if m["a"] != "1" || m["b"] != "2" {
		t.Errorf("unexpected values: %v", m)
	}
}

func TestPatch_OverwritesExisting(t *testing.T) {
	repo := newFakeRepo()
	repo.Set(context.Background(), "key", "old")
	svc := newService(repo, nil)

	changed, err := svc.Patch(context.Background(), map[string]string{"key": "new"})
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	if len(changed) != 1 || changed[0] != "key" {
		t.Errorf("unexpected changed: %v", changed)
	}

	m, _ := svc.All(context.Background())
	if m["key"] != "new" {
		t.Errorf("key = %q, want new", m["key"])
	}
}

func TestPatch_SkipsEmptyValues(t *testing.T) {
	repo := newFakeRepo()
	repo.Set(context.Background(), "keep", "val")
	svc := newService(repo, nil)

	changed, err := svc.Patch(context.Background(), map[string]string{"keep": ""})
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	if len(changed) != 0 {
		t.Errorf("empty values must be skipped, got %v", changed)
	}

	m, _ := svc.All(context.Background())
	if m["keep"] != "val" {
		t.Error("existing value must be preserved when patching with empty")
	}
}

func TestPatchAndRebuild_TriggersOnChange(t *testing.T) {
	repo := newFakeRepo()
	trig := &fakeTrigger{}
	svc := newService(repo, trig)

	err := svc.PatchAndRebuild(context.Background(), map[string]string{"log_level": "debug"})
	if err != nil {
		t.Fatalf("PatchAndRebuild: %v", err)
	}
	if trig.triggered() != 1 {
		t.Errorf("trigger count = %d, want 1", trig.triggered())
	}
}

func TestPatchAndRebuild_NoTriggerOnEmpty(t *testing.T) {
	repo := newFakeRepo()
	repo.Set(context.Background(), "key", "val")
	trig := &fakeTrigger{}
	svc := newService(repo, trig)

	err := svc.PatchAndRebuild(context.Background(), map[string]string{"key": ""})
	if err != nil {
		t.Fatalf("PatchAndRebuild: %v", err)
	}
	if trig.triggered() != 0 {
		t.Errorf("trigger must NOT fire when nothing changed, got %d", trig.triggered())
	}
}

func TestPatchAndRebuild_NilTriggerDoesNotPanic(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(repo, nil)

	err := svc.PatchAndRebuild(context.Background(), map[string]string{"a": "1"})
	if err != nil {
		t.Fatalf("PatchAndRebuild with nil trigger: %v", err)
	}
}

type faultRepo struct{}

func (faultRepo) All(context.Context) (map[string]string, error) { return nil, errors.New("db down") }
func (faultRepo) Set(context.Context, string, string) error       { return errors.New("db down") }

func TestAll_PropagatesRepoError(t *testing.T) {
	svc := svcsettings.New(faultRepo{}, nil)
	_, err := svc.All(context.Background())
	if err == nil {
		t.Error("All must return an error when the repo fails")
	}
}

func TestPatch_PropagatesRepoError(t *testing.T) {
	svc := svcsettings.New(faultRepo{}, nil)
	_, err := svc.Patch(context.Background(), map[string]string{"a": "1"})
	if err == nil {
		t.Error("Patch must return an error when the repo fails")
	}
}
