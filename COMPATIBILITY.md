# Compatibility

| Component | Requirement |
|-----------|-------------|
| Contract interface | `IndexerService` **v1** |
| Go module path | `github.com/Muxcore-Media/contracts-indexer` |
| Generated package | `github.com/Muxcore-Media/contracts-indexer/muxcore/indexer/v1` |
| MuxCore core | ≥ 0.4.0 (consumers) |
| Capability ID (contract repo) | `contracts.indexer` |
| Capability ID (implementers / callers) | `indexer` |

Implementers register capability **`indexer`** in `muxcore.json` and implement gRPC service **`IndexerService`**. The contract repository capability id **`contracts.indexer`** identifies this proto module in the umbrella catalog.

## Search type vocabulary

`SearchRequest.type` accepts lowercase string tokens (see `SearchType` enum in proto):

| Token | Meaning |
|-------|---------|
| `movie` | Theatrical / feature release |
| `tv` | Series episode or season pack |
| `music` | Audio release |
| `book` | E-book |
| `audiobook` | Spoken-word release |
| `comic` | Comic / graphic novel |

Implementers may accept legacy aliases but **must** accept these canonical tokens. `categories` carries indexer-specific Newznab/Torznab category IDs when callers know them.

## TV and anime query shaping

| Pattern | Fields | Meaning |
|---------|--------|---------|
| Season pack | `season>0`, `episode=0` | Full-season search; omit Exx from query |
| Episode | `season>0`, `episode>0` | Single episode |
| Anime absolute | `absolute>0` | Absolute episode number (field tag 10) |

Callers check `GetCapabilitiesResponse.supports_season_pack` before relying on season-pack semantics.

## Paging

| Field | Rule |
|-------|------|
| `SearchRequest.limit` | Default **50** when unset or zero; maximum **100** |
| `SearchRequest.offset` | Zero-based page start |
| `SearchResponse.total` | **Unpaged** hit count across all matches — not `len(results)` |

Implementers must set `total` to the full match count before applying limit/offset to `results`.

## ID search

When `GetCapabilitiesResponse.supports_id_search` is true, implementers should honor optional `SearchRequest.imdb_id`, `tmdb_id`, and `tvdb_id` for Newznab/Torznab ID queries.

## Info hash

`SearchResult.infohash` carries the torrent info hash when known. Callers should prefer it over parsing `download_url` magnet links.

## Indexer CRUD

`GetIndexer` / `CreateIndexer` / `UpdateIndexer` / `DeleteIndexer` are additive v1 RPCs for household Prowlarr catalog edits.

| Situation | Behavior |
|-----------|----------|
| Prowlarr mode (`PROWLARR_URL`) | Proxy `/api/v1/indexer` for Torznab/Newznab feeds |
| Direct Torznab / Jackett / piratebay | `FailedPrecondition` or `Unimplemented` |
| `IndexerSpec.api_key` | Write-only — never returned on get/list |

## Error and privacy contract

| Situation | Behavior |
|-----------|----------|
| Rate limited | Return gRPC `ResourceExhausted` (`media-automation` holds searches 15 minutes) |
| Unconfigured indexer | Return **empty** `SearchResponse` — not an error |
| `SearchResult.download_url` | **Must not** embed indexer API keys or credentials (strip Torznab `apikey=`, etc.) |

Breaking proto changes require a new major contract version and coordinated consumer updates.
