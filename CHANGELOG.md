# Changelog

## [0.2.0] — 2026-08-31

### Added
- `SearchType` enum and documented canonical `type` tokens (`movie`, `tv`, `music`, `book`, `audiobook`, `comic`).
- `SearchRequest` ID fields: `imdb_id`, `tmdb_id`, `tvdb_id`.
- `SearchResult.infohash` for torrent info hash without magnet parsing.
- `GetCapabilitiesResponse`: `supports_music_search`, `supports_book_search`, `supports_id_search`, `supports_season_pack`.
- Paging contract: default limit 50, max 100; `total` is unpaged hit count.
- Error/privacy contract: `ResourceExhausted` for rate limits; empty results when unconfigured; no API keys in `download_url`.
- `contract_test.go` freezing RPC names and field numbers.
- Forgejo CI `proto-check` job.

### Fixed
- `go_package` aligned with checked-in stubs (`muxcore/indexer/v1`, not `gen/`).
- Makefile `clean` targets `muxcore/indexer/v1/*.go`.

### Compatibility
- Additive **v1** proto changes; implementers and callers may adopt new fields incrementally.
- Capability **`indexer`** (implementers) vs **`contracts.indexer`** (catalog).

## [0.1.0] — 2026-08-08

### Added
- Initial `IndexerService` contract (`Search`, `GetCapabilities`, `ListIndexers`).
- Generated Go stubs under `muxcore/indexer/v1`.
- Documented TV season-pack (`episode=0`) and anime `absolute` search fields.

### Compatibility
- Contract interface version **v1**; implementers advertise capability `indexer`.
- Consumers: `indexer-piratebay`, `media-automation` (fan-out Search).
