package apitoken_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/repo"
	"sing-box-web-panel/internal/services/apitoken"
)

type fakeRepo struct {
	tokens []domain.APIToken
}

func (r *fakeRepo) Create(_ context.Context, t *domain.APIToken) error {
	for i := range r.tokens {
		if r.tokens[i].Name == t.Name {
			return repo.ErrExist
		}
	}
	t.ID = int64(len(r.tokens) + 1)
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = t.CreatedAt
	r.tokens = append(r.tokens, *t)
	return nil
}

func (r *fakeRepo) List(context.Context) ([]domain.APIToken, error) {
	return append([]domain.APIToken(nil), r.tokens...), nil
}

func (r *fakeRepo) ListEnabled(context.Context) ([]domain.APIToken, error) {
	var out []domain.APIToken
	for i := range r.tokens {
		if r.tokens[i].Enabled {
			out = append(out, r.tokens[i])
		}
	}
	return out, nil
}

func (r *fakeRepo) SetEnabled(_ context.Context, id int64, enabled bool) error {
	for i := range r.tokens {
		if r.tokens[i].ID == id {
			r.tokens[i].Enabled = enabled
			return nil
		}
	}
	return repo.ErrNotFound
}

func (r *fakeRepo) Delete(_ context.Context, id int64) error {
	for i := range r.tokens {
		if r.tokens[i].ID == id {
			r.tokens = append(r.tokens[:i], r.tokens[i+1:]...)
			return nil
		}
	}
	return repo.ErrNotFound
}

func (r *fakeRepo) Touch(_ context.Context, id int64, at time.Time) error {
	for i := range r.tokens {
		if r.tokens[i].ID == id {
			r.tokens[i].LastUsedAt = &at
			return nil
		}
	}
	return repo.ErrNotFound
}

func TestCreateReturnsRawTokenOnceAndStoresHash(t *testing.T) {
	r := &fakeRepo{}
	svc := apitoken.NewService(r)

	created, err := svc.Create(context.Background(), "node-a", "node")
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if created.Raw == "" {
		t.Fatal("expected raw token")
	}
	if created.Token.TokenHash == "" || created.Token.TokenHash == created.Raw {
		t.Fatalf("expected stored hash, got %q", created.Token.TokenHash)
	}
	if created.Token.TokenPrefix == "" {
		t.Fatal("expected token prefix")
	}
}

func TestVerifyHonorsEnabledAndScope(t *testing.T) {
	r := &fakeRepo{}
	svc := apitoken.NewService(r)
	created, err := svc.Create(context.Background(), "node-a", "node")
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	if _, err := svc.Verify(context.Background(), created.Raw, "node"); err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if _, err := svc.Verify(context.Background(), created.Raw, "admin"); !errors.Is(err, apitoken.ErrUnauthorized) {
		t.Fatalf("expected scope rejection, got %v", err)
	}
	if err := svc.SetEnabled(context.Background(), created.Token.ID, false); err != nil {
		t.Fatalf("disable token: %v", err)
	}
	if _, err := svc.Verify(context.Background(), created.Raw, "node"); !errors.Is(err, apitoken.ErrUnauthorized) {
		t.Fatalf("expected disabled rejection, got %v", err)
	}
}
