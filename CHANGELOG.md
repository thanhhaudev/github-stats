# Changelog

All notable changes to this project will be documented in this file. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) starting at `v1.6.0`.

> Earlier releases (`v1.0` … `v1.4.0`, `1.5.0` … `1.5.2`) predate this changelog and are reconstructable from the git history. From `v1.6.0` onward, every release pins a tagged entry below.

## [Unreleased]

## [1.6.0] - 2026-04-26

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

[Unreleased]: https://github.com/thanhhaudev/github-stats/compare/v1.6.0...HEAD
[1.6.0]: https://github.com/thanhhaudev/github-stats/releases/tag/v1.6.0
