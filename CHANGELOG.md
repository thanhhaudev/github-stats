# Changelog

All notable changes to this project will be documented in this file. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) starting at `v1.5.3`.

> Earlier releases (`v1.0` … `v1.4.0`, `1.5.0` … `1.5.2`) predate this changelog and are reconstructable from the git history. From `v1.5.3` onward, every release pins a tagged entry below.

## [Unreleased]

## [1.5.5] - 2026-05-18

### Added
- Root `Makefile` targets for local development, including GitHub CLI-backed runs that avoid storing `GITHUB_TOKEN` in `cmd/.env`.
- Focused regression coverage for commit fetching, WakaTime cache fallback, README marker replacement, pagination, and config metadata parity.

### Changed
- Upgrade the Go module, Docker builder image, and CI setup to Go 1.26.3.
- Make `pkg/container` receive its clock dependency explicitly instead of requiring a context value.
- Make commit fetching return deterministic repository-level results while branch fetches run concurrently.

### Fixed
- Redact sensitive runtime logs.
- Compact AI footprint line counts.
- Preserve valid UTF-8 when truncating Unicode stats labels.
- Return branch fetch failures with repo and branch context instead of sharing a top-level error channel across repo and branch workers.

## [1.5.4] - 2026-05-18

### Changed
- Cache successful WakaTime stats when `ENABLE_CACHE=true`, and reuse cached WakaTime data when the API is still processing or returns stale stats.
- README now warns users to fork the Action or pin a specific release/SHA instead of relying on the floating `v1` reference.

### Fixed
- GitHub metrics continue updating when WakaTime stats are not ready, instead of skipping the entire README update.

## [1.5.3] - 2026-05-17

### Added
- Automated release workflow: pushing a `vX.Y.Z` tag publishes a GitHub Release and updates the floating `v1` tag.
- `CHANGELOG.md` (this file).
- `docs/` split: `docs/metrics.md`, `docs/configuration.md`, `docs/caching.md`, `docs/scheduling.md`.

### Changed
- README rewritten in concise format (513 → ~85 lines); advanced topics moved to `docs/`.
- `pkg/container.metrics()` now references `config.Metric*` constants instead of bare string literals.
- README + docs reference `thanhhaudev/github-stats@v1` instead of `@master` so users pin a stable major.

### Fixed
- AI footprint row label: `Total Prompt Chars` (raw `ai_prompt_length`) replaces the misleading `Average Prompt` while WakaTime omits `ai_average_prompt_length` from `/stats`. Reverts to `Average Prompt` automatically once the field appears.

[Unreleased]: https://github.com/thanhhaudev/github-stats/compare/v1.5.5...HEAD
[1.5.5]: https://github.com/thanhhaudev/github-stats/releases/tag/v1.5.5
[1.5.4]: https://github.com/thanhhaudev/github-stats/releases/tag/v1.5.4
[1.5.3]: https://github.com/thanhhaudev/github-stats/releases/tag/v1.5.3
