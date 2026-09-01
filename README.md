# contracts-indexer

Protobuf/gRPC contract interfaces for search indexer modules in MuxCore.

## Proto Service

- **`IndexerService`** (`proto/muxcore/indexer/v1/indexer.proto`)
  - `Search` — search for media releases across configured indexers
  - `GetCapabilities` — discover indexer search features and supported categories
  - `ListIndexers` — enumerate configured indexers with their status

### `SearchRequest` notes

- **type** — canonical tokens: `movie`, `tv`, `music`, `book`, `audiobook`, `comic` (see `SearchType` enum)
- **categories** — indexer-specific Newznab/Torznab category IDs when known
- TV season-pack: `episode=0` with `season>0`
- **absolute** — anime absolute-number field (tag 10)
- **imdb_id** / **tmdb_id** / **tvdb_id** — optional ID search when implementer supports it
- **limit** — default 50, max 100; **total** in the response is unpaged hit count

### `SearchResult` notes

- **infohash** — torrent info hash when known; prefer over parsing magnets
- **download_url** — must not embed indexer API keys

## Version

**v0.2.0** — see [CHANGELOG.md](CHANGELOG.md) and [COMPATIBILITY.md](COMPATIBILITY.md).

## Go import

Generated stubs: `github.com/Muxcore-Media/contracts-indexer/muxcore/indexer/v1`

## Implementing Modules

Any module advertising capability **`indexer`** may implement this service. Valid shapes:

- **Single-site** — one tracker or API (e.g. a Torznab endpoint or native site module)
- **Aggregator** — one module that wraps many configured sites and returns results with distinct `indexer_name` / `indexer_id`

Callers such as `media-automation` discover **all** `indexer` modules and may fan out `Search` in parallel; aggregators are optional, not required hubs.

### Known implementers

- [indexer-piratebay](https://github.com/Muxcore-Media/indexer-piratebay) — single-site The Pirate Bay (Apibay-compatible JSON)
- [indexer-torznab](https://github.com/Muxcore-Media/indexer-torznab) — Torznab/Newznab adapter (Prowlarr, Jackett, etc.)

`indexer-prowlarr` was planned as a dedicated aggregator module; **indexer-torznab** supersedes that approach via standard Torznab feeds.

## Regenerate protos

```bash
make proto
```

CI runs `make proto` and fails if generated stubs drift.
