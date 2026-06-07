---
type: "query"
date: "2026-06-07T08:04:38.083349+00:00"
question: "SIN-68 Naive Proxy does not work"
contributor: "graphify"
source_nodes: ["ErrNaiveJSONRequiresTrustedTLS", "BuildClientConfig()", "SubscriptionHandler.Serve()", "stage_binaries", "commit_binaries"]
---

# Q: SIN-68 Naive Proxy does not work

## Answer

Fixed Naive JSON subscription and installer support. Naive JSON no longer emits unsupported tls.insecure. Self-signed Naive JSON returns HTTP 400 explaining trusted TLS is required; plain/base64 links still work. Installer stages and commits libcronet.so from the official sing-box tarball next to the sing-box binary. Local checks passed: go test ./tests/services/sublink, go test ./tests/transport/handler, go test ./tests/..., go build ./..., go vet ./..., bash -n scripts/install.sh. RU node verified with shilka v1.11.1-sin68, active service, libcronet.so present, Naive outbound sing-box check passing, JSON self-signed subscription returning 400, plain subscription returning 200.

## Source Nodes

- ErrNaiveJSONRequiresTrustedTLS
- BuildClientConfig()
- SubscriptionHandler.Serve()
- stage_binaries
- commit_binaries