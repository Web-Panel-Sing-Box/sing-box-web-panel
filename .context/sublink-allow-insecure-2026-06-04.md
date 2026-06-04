# Sub-link Allow Insecure Fix

## Task

Implemented Linear issue SIN-53: generated subscription links could not connect to Hysteria2, Naive, or VLESS TLS inbounds that use self-signed or otherwise untrusted certificates because client-side certificate verification was always strict.

## What Changed

- Added a nullable `allowInsecure` setting to inbound settings JSON. `nil` keeps existing inbounds in automatic mode, while explicit `true` or `false` stores the administrator's override.
- Added a shared effective rule on `domain.Inbound`: the generated client output allows insecure TLS only for `tls` mode. If no override is stored, it defaults to enabled when no ACME domain and no certificate/key paths are configured.
- Exposed effective `settings.allowInsecure` through inbound API responses and accepted `allowInsecure` on inbound create/update requests.
- Updated generated share links:
  - Hysteria2 now renders `insecure=1` when effective and `insecure=0` otherwise.
  - VLESS TLS and Naive links render `allowInsecure=1` when effective.
  - Reality and non-TLS links do not receive insecure flags.
- Updated sing-box JSON subscription output so outbound TLS includes `insecure: true` when effective.
- Added an `Allow insecure` toggle to the inbound TLS form and wired it into the frontend payload for VLESS TLS, Naive, and Hysteria2.
- Regenerated Swagger docs so the new API field is documented.

## Files Touched

- Backend domain/service/API: `internal/domain/inbound.go`, `internal/services/inbound/service.go`, `internal/transport/handler/inbound_handler.go`
- Subscription generation: `internal/services/sublink/builder.go`, `internal/services/sublink/subscription.go`
- Frontend API/form/i18n: `frontend/src/api/inbounds.ts`, `frontend/src/hooks/useInboundForm.ts`, `frontend/src/components/inbounds/inbound-form-modal.tsx`, `frontend/src/lib/i18n.tsx`
- Tests: `tests/services/sublink/builder_test.go`, `tests/services/inbound/service_test.go`, `tests/transport/handler/inbound_handler_test.go`, `frontend/src/components/inbounds/inbound-form-modal.test.tsx`
- Generated docs: `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`

## Verification

- `go test ./tests/services/sublink ./tests/services/inbound ./tests/transport/handler`
- `cd frontend && pnpm typecheck`
- `cd frontend && pnpm test`
- `cd frontend && pnpm build`
- `go test ./tests/...`
- `go build ./...`
- `go vet ./...`
- `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/main.go -o docs --parseDependency --parseInternal`
