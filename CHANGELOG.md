# Changelog

## [0.1.0] — 2026-08-08

### Added
- Initial `IndexerService` contract (`Search`, `GetCapabilities`, `ListIndexers`).
- Generated Go stubs under `muxcore/indexer/v1`.
- Documented TV season-pack (`episode=0`) and anime `absolute` search fields.

### Compatibility
- Contract interface version **v1**; implementers advertise capability `indexer`.
- Consumers: `indexer-piratebay`, `media-automation` (fan-out Search).
