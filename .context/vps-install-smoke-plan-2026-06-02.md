# VPS Install And Smoke Plan - 2026-06-02

## Purpose

This document records the no-downtime VPS installation and smoke-test plan for
Shilka. The goal is to validate the current release installer, embedded panel,
sing-box lifecycle wiring, API tokens, and node sync flow on the target server
without interrupting existing services.

## Current Context

- Target host: `2.26.22.151`.
- SSH port: `5367`.
- Login mode: root SSH is available. Do not store credentials in repo files.
- OS: Ubuntu 24.04.3 LTS.
- CPU architecture: `x86_64`, so the installer should select `amd64` assets.
- Existing services:
  - `nginx.service` is active and owns ports `80` and `443`.
  - Docker is active.
  - Docker container `dataset-agent-app` owns public port `3005`.
  - SSH is exposed on port `5367`.
- Existing Shilka state:
  - No `shilka.service` or `sing-box.service` unit files were present during
    read-only inspection.
  - `/opt/shilka`, `/etc/shilka`, `/var/lib/shilka`, `/var/log/shilka`, and
    `/usr/local/bin/shilka` did not exist during read-only inspection.

## Chosen Install Mode

- Use no-downtime installation.
- Do not stop or reconfigure nginx.
- Do not stop or reconfigure Docker or the existing container on port `3005`.
- Do not use Let's Encrypt standalone during the initial smoke run because
  ports `80` and `443` are already occupied by nginx.
- Install Shilka on high port `18443`.
- Use self-signed panel TLS for the initial smoke run.
- Accept the installer-generated random base path and save it as `PANEL_PATH`.
- Pin release versions for reproducibility:
  - `GITHUB_REPO=Web-Panel-Sing-Box/shilka-web-panel`
  - `PANEL_VERSION=1.7.0`
  - `SING_BOX_VERSION=1.13.12`

## Installation Commands

Connect as root:

```bash
ssh -p 5367 root@2.26.22.151
```

Run a final preflight:

```bash
ss -tulpen
systemctl status nginx --no-pager
docker ps
ufw status verbose
```

Download and run the pinned installer:

```bash
curl -fsSL https://raw.githubusercontent.com/Web-Panel-Sing-Box/shilka-web-panel/main/scripts/install.sh -o /tmp/shilka-install.sh
chmod +x /tmp/shilka-install.sh
GITHUB_REPO=Web-Panel-Sing-Box/shilka-web-panel PANEL_VERSION=1.7.0 SING_BOX_VERSION=1.13.12 bash /tmp/shilka-install.sh
```

Installer answers:

- Domain/IP: `2.26.22.151`
- Use Let's Encrypt for this IP: `n`
- Panel port: `18443`
- Panel path: accept generated random path and store it as `PANEL_PATH`
- Admin username: accept generated value or enter an operator-approved value
- Admin password: accept generated value or enter an operator-approved value
- Proceed: `y`

If UFW is active and blocks the panel port:

```bash
ufw allow 18443/tcp
```

## Immediate Post-Install Fix

The current installer writes `subscription.public_url` as
`https://host:port` and does not include `http.base_path`. Until the installer
is fixed, update it manually after installation:

```bash
export PANEL_PATH="/generated-installer-path"
shilka setting -public-url "https://2.26.22.151:18443${PANEL_PATH}"
systemctl restart shilka.service
```

## Health Checks

Check systemd and recent logs:

```bash
systemctl status shilka.service --no-pager
journalctl -u shilka.service -n 120 --no-pager
```

Check Shilka settings and sing-box status:

```bash
shilka setting -show
shilka core status
shilka core config-check
```

Check local HTTPS endpoints:

```bash
export BASE="https://127.0.0.1:18443${PANEL_PATH}"
curl -sk "$BASE/api/health"
curl -sk "$BASE/api"
curl -sk -o /dev/null -w "%{http_code}\n" "$BASE/"
```

## API Smoke Flow

Set smoke variables:

```bash
export BASE="https://127.0.0.1:18443${PANEL_PATH}"
export ADMIN_USER="<installer-admin-user>"
export ADMIN_PASSWORD="<installer-admin-password>"
```

Login and verify the admin profile:

```bash
export TOKEN="$(curl -sk -X POST "$BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"${ADMIN_USER}\",\"password\":\"${ADMIN_PASSWORD}\"}" \
  | jq -r '.token')"

test -n "$TOKEN"
curl -sk "$BASE/api/auth/me" -H "Authorization: Bearer ${TOKEN}"
```

Create a temporary VLESS TCP inbound on a high test port:

```bash
export INBOUND_ID="$(curl -sk -X POST "$BASE/api/inbounds" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"remark":"smoke-vless-tcp","protocol":"vless","port":28443,"transmission":"tcp","tls":"none"}' \
  | jq -r '.id')"

test -n "$INBOUND_ID"
```

Create a temporary client:

```bash
export CLIENT_ID="$(curl -sk -X POST "$BASE/api/clients" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"smoke-client\",\"inboundId\":\"${INBOUND_ID}\",\"totalQuota\":0}" \
  | jq -r '.id')"

test -n "$CLIENT_ID"
```

Verify client list, client links, subscriptions, and core config:

```bash
curl -sk "$BASE/api/clients?inboundId=${INBOUND_ID}" -H "Authorization: Bearer ${TOKEN}"
curl -sk "$BASE/api/clients/${CLIENT_ID}/links" -H "Authorization: Bearer ${TOKEN}"
shilka core config-check
```

The subscription URL must include `${PANEL_PATH}`. If it does not, the manual
`subscription.public_url` fix was missed or did not persist.

## API Token And Node Smoke Flow

Create a temporary API token:

```bash
export NODE_TOKEN="$(shilka api-token create -name smoke-node -scopes node | jq -r '.token')"
test -n "$NODE_TOKEN"
```

Verify node API accepts the API token:

```bash
curl -sk "$BASE/api/node/v1/status" -H "Authorization: Bearer ${NODE_TOKEN}"
```

Verify browser-admin endpoints do not accept API-token bearer credentials:

```bash
curl -sk -o /dev/null -w "%{http_code}\n" "$BASE/api/inbounds" -H "Authorization: Bearer ${NODE_TOKEN}"
```

Expected result: `401`.

Add a temporary self-node with private-address access explicitly enabled:

```bash
export NODE_ID="$(shilka node add \
  -name smoke-self \
  -scheme https \
  -address 127.0.0.1 \
  -port 18443 \
  -base-path "${PANEL_PATH}" \
  -token "${NODE_TOKEN}" \
  -allow-private \
  | jq -r '.id')"

test -n "$NODE_ID"
shilka node probe "$NODE_ID"
shilka node sync "$NODE_ID"
```

## Cleanup

Remove temporary smoke resources:

```bash
curl -sk -X DELETE "$BASE/api/clients/${CLIENT_ID}" -H "Authorization: Bearer ${TOKEN}"
curl -sk -X DELETE "$BASE/api/inbounds/${INBOUND_ID}" -H "Authorization: Bearer ${TOKEN}"
shilka node delete "$NODE_ID"
```

Find and revoke the smoke API token:

```bash
shilka api-token list
shilka api-token revoke "<smoke-token-id>"
```

Run final checks:

```bash
shilka core config-check
systemctl status shilka.service --no-pager
journalctl -u shilka.service -n 120 --no-pager
```

## Acceptance Criteria

- Existing nginx, Docker, SSH, and the public app on port `3005` remain active.
- `shilka.service` is active and enabled.
- `/opt/shilka`, `/etc/shilka`, `/var/lib/shilka`, and `/var/log/shilka`
  exist with expected Shilka ownership and permissions.
- Panel is reachable at `https://2.26.22.151:18443${PANEL_PATH}`.
- `/api/health`, `/api`, `/api/auth/login`, `/api/auth/me`, `/api/core/status`,
  and `/api/core/config` respond as expected.
- `shilka core config-check` passes before and after smoke CRUD.
- Temporary inbound/client CRUD works.
- Client subscription URL includes the generated base path.
- API-token auth works only on `/api/node/v1/*`.
- Temporary self-node `probe` and `sync` succeed with `allow-private`.
- Smoke resources are removed after testing.
- Journal and panel logs do not contain admin passwords, JWT secrets, API token
  raw values, subscription tokens, client UUID lists, or private keys.

## Follow-Up Linear Work

Create separate implementation issues for:

- Running the VPS no-downtime install smoke.
- Fixing installer `subscription.public_url` base-path handling.
- Hardening installer behavior for occupied `80`/`443` and reverse-proxy setups.
- Adding checksum verification for downloaded Shilka release assets.
- Updating README install and development commands.
