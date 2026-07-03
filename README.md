# contracts-indexer

Protobuf/gRPC contract interfaces for search indexer modules in MuxCore.

## Proto Service

- **`IndexerService`** (`proto/muxcore/indexer/v1/indexer.proto`)
  - `Search` — search for media releases across configured indexers
  - `GetCapabilities` — discover indexer search features and supported categories
  - `ListIndexers` — enumerate configured indexers with their status

## Implementing Modules

- [indexer-prowlarr](https://github.com/Muxcore-Media/indexer-prowlarr) — Prowlarr-backed search indexer
