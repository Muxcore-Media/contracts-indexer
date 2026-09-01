# AGENTS.md — contracts-indexer

Protobuf/gRPC contract library — **not** a runnable sidecar. Workspace context: [`../AGENTS.md`](../AGENTS.md).

## Module identity

| Field | Value |
|-------|-------|
| Directory | `contracts-indexer` |
| Go module | `github.com/Muxcore-Media/contracts-indexer` |
| Generated import | `github.com/Muxcore-Media/contracts-indexer/muxcore/indexer/v1` |
| Proto source | `proto/muxcore/indexer/v1/indexer.proto` |
| Catalog capability | `contracts.indexer` (`muxcore.json`) |
| Implementer capability | `indexer` (mesh peers, callers) |

## Agent rules

- **Proto changes are the product.** Edit `.proto` first, then regenerate stubs.
- Run `make proto` after any `.proto` edit; commit generated `muxcore/indexer/v1/*.go` with the proto change.
- CI (`proto-check` in `.forgejo/workflows/ci.yml` and `.github/workflows/ci.yml`) fails if stubs drift.
- Additive field changes only on **v1** — new fields get new tag numbers; never renumber or change wire types.
- Breaking changes require a new major package (`muxcore/indexer/v2`) and coordinated consumer bumps.
- Extend `muxcore/indexer/v1/contract_test.go` when adding RPCs or freezing new field numbers.
- Document compatibility rules in `COMPATIBILITY.md` and proto comments (search types, paging, errors, privacy).
- Do not edit polluted workspace dumps (see `MASTER-ROADMAP.md` Appendix H).

## Regenerate protos

Use `nix-shell` when protoc/plugins are not on PATH:

```bash
cd contracts-indexer
nix-shell -p protobuf go --run 'go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1 && export PATH="$(go env GOPATH)/bin:$PATH" && make proto'
```

## Build / test

```bash
cd contracts-indexer
nix-shell -p go --run 'go test ./...'
```

`make clean` removes generated stubs under `muxcore/indexer/v1/` — use only when intentionally regenerating from scratch.

## Compatibility checklist (proto edits)

1. Canonical search types documented (`SearchType` enum + `type` string tokens).
2. Paging: default limit 50, max 100; `total` is unpaged count.
3. Error contract: `ResourceExhausted` = rate limited; unconfigured = empty results.
4. Privacy: `download_url` must not leak API keys.
5. Update `CHANGELOG.md`, `COMPATIBILITY.md`, and README when behavior is caller-visible.
