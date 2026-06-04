#!/usr/bin/env bash
set -Eeuo pipefail

VERSION="${1:-}"

if [[ -z "${VERSION}" ]]; then
  echo "usage: $0 vX.Y.Z" >&2
  exit 2
fi

if [[ "${VERSION}" != v* || "${VERSION}" =~ [[:space:]] ]]; then
  echo "ERROR: release version must be a tag like v1.9.2" >&2
  exit 2
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

write_checksums() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum shilka-linux-amd64 shilka-linux-arm64 > checksums.txt
    return
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 shilka-linux-amd64 shilka-linux-arm64 > checksums.txt
    return
  fi
  echo "ERROR: sha256sum or shasum is required" >&2
  exit 1
}

echo "Building Shilka release assets for ${VERSION}"

pnpm --dir frontend install --frozen-lockfile
pnpm --dir frontend build

install -d cmd/frontend/dist dist
if command -v rsync >/dev/null 2>&1; then
  rsync -a --delete --exclude=PLACEHOLDER frontend/dist/ cmd/frontend/dist/
else
  find cmd/frontend/dist -mindepth 1 ! -name PLACEHOLDER -exec rm -rf {} +
  cp -R frontend/dist/. cmd/frontend/dist/
fi
touch cmd/frontend/dist/PLACEHOLDER

rm -f dist/shilka-linux-amd64 dist/shilka-linux-arm64 dist/checksums.txt

LDFLAGS="-s -w -X sing-box-web-panel/internal/version.Version=${VERSION}"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="${LDFLAGS}" -o dist/shilka-linux-amd64 ./cmd/
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="${LDFLAGS}" -o dist/shilka-linux-arm64 ./cmd/

(
  cd dist
  write_checksums
)

if [[ "$(go env GOOS)" == "linux" && "$(go env GOARCH)" == "amd64" ]]; then
  actual="$(./dist/shilka-linux-amd64 version)"
  expected="shilka ${VERSION}"
  if [[ "${actual}" != "${expected}" ]]; then
    echo "ERROR: version smoke check failed: got ${actual}, want ${expected}" >&2
    exit 1
  fi
fi
