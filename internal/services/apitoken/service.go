package apitoken

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/lib/keys"
)

type Repo interface {
	Create(ctx context.Context, t *domain.APIToken) error
	List(ctx context.Context) ([]domain.APIToken, error)
	ListEnabled(ctx context.Context) ([]domain.APIToken, error)
	SetEnabled(ctx context.Context, id int64, enabled bool) error
	Delete(ctx context.Context, id int64) error
	Touch(ctx context.Context, id int64, at time.Time) error
}

var ErrUnauthorized = errors.New("api token unauthorized")

type CreatedToken struct {
	Token domain.APIToken
	Raw   string
}

type Service struct {
	repo Repo
	now  func() time.Time
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo, now: time.Now}
}

func (s *Service) Create(ctx context.Context, name, scopes string) (*CreatedToken, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if scopes == "" {
		scopes = "node"
	}
	raw, err := keys.GenerateToken(36)
	if err != nil {
		return nil, err
	}
	token := &domain.APIToken{
		Name:        name,
		TokenHash:   hash(raw),
		TokenPrefix: prefix(raw),
		Scopes:      scopes,
		Enabled:     true,
	}
	if err := s.repo.Create(ctx, token); err != nil {
		return nil, err
	}
	return &CreatedToken{Token: *token, Raw: raw}, nil
}

func (s *Service) List(ctx context.Context) ([]domain.APIToken, error) {
	return s.repo.List(ctx)
}

func (s *Service) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	return s.repo.SetEnabled(ctx, id, enabled)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Verify(ctx context.Context, raw, requiredScope string) (*domain.APIToken, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrUnauthorized
	}
	want := hash(raw)
	tokens, err := s.repo.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	for i := range tokens {
		if subtle.ConstantTimeCompare([]byte(tokens[i].TokenHash), []byte(want)) != 1 {
			continue
		}
		if requiredScope != "" && !hasScope(tokens[i].Scopes, requiredScope) {
			return nil, ErrUnauthorized
		}
		_ = s.repo.Touch(ctx, tokens[i].ID, s.now().UTC())
		return &tokens[i], nil
	}
	return nil, ErrUnauthorized
}

func hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func prefix(raw string) string {
	if len(raw) <= 8 {
		return raw
	}
	return raw[:8]
}

func hasScope(scopes, required string) bool {
	for _, s := range strings.Split(scopes, ",") {
		s = strings.TrimSpace(s)
		if s == "*" || s == required {
			return true
		}
	}
	return false
}
