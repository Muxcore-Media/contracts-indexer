# contracts-indexer

Protobuf/gRPC contract interfaces for search indexer modules in MuxCore.

## Proto Service

- **`IndexerService`** (`proto/muxcore/indexer/v1/indexer.proto`)
  - `Search` — search for media releases across configured indexers
  - `GetCapabilities` — discover indexer search features and supported categories
  - `ListIndexers` — enumerate configured indexers with their status

### `SearchRequest` notes

- TV: `episode=0` with `season>0` means season-pack search (no Exx in query)
- `absolute` — anime absolute-number field; indexers may ignore it while still matching an absolute token in `query`

## Go import

Generated stubs: `github.com/Muxcore-Media/contracts-indexer/muxcore/indexer/v1`

## Implementing Modules

Any module advertising capability `indexer` may implement this service. Valid shapes:

- **Single-site** — one tracker or API (e.g. a Torznab endpoint or native site module)
- **Aggregator** — one module that wraps many configured sites (e.g. Prowlarr) and returns results with distinct `indexer_name` / `indexer_id`

Callers such as `media-automation` discover **all** `indexer` modules and may fan out `Search` in parallel; aggregators are optional, not required hubs.

### Known / planned implementers

- [indexer-piratebay](https://github.com/Muxcore-Media/indexer-piratebay) — single-site The Pirate Bay (Apibay-compatible JSON)
- `indexer-prowlarr` — Prowlarr-backed aggregator (planned; repo not published)
