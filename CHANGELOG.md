# Changelog

All notable changes to this project will be documented in this file. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) starting at `v1.5.3`.

> Earlier releases (`v1.0` … `v1.4.0`, `1.5.0` … `1.5.2`) predate this changelog and are reconstructable from the git history. From `v1.5.3` onward, every release pins a tagged entry below.

## [Unreleased]

### Changed
- **Language statistics now reflect the lines you actually wrote.** `LANGUAGES_AND_TOOLS` and `LANGUAGE_PER_REPO` are computed from the files changed in your own commits (vendored/generated files excluded) instead of the whole-repo language bar of every repository you touched. Reported languages and percentages will change. The first run after upgrading is slower because each commit is analyzed once; results are cached per commit, so later runs stay fast.

## [1.5.9] - 2026-05-28

### Fixed
- Push regression from the hardened-auth change in this release. `actions/checkout` (default `persist-credentials: true`) leaves an `http.https://github.com/.extraheader` entry in the workspace `.git/config`. Git treats `extraheader` as multi-valued, so the new entry injected via `GIT_CONFIG_VALUE_0` was sent **in addition to** the existing one. GitHub then rejected the push with HTTP 400 `Duplicate header: "Authorization"` before validating either token — even when both tokens were identical. `pushReadme` now runs `git config --local --unset-all http.https://github.com/.extraheader` before injecting its own, tolerating exit 5 when no prior entry exists.
- `HIDE_REPO_INFO=true` no longer collapses every git failure to a bare `exit status N`. The suppressed branch of `runGitCommandEnv` was capturing stdout/stderr into buffers and then discarding them. Sanitized stdout/stderr (still scrubbed of token, owner, URLs, and github-token patterns) are now included in the returned error so push failures are debuggable.

### Security
- Stop persisting `GITHUB_TOKEN` in `.git/config`. The `origin` URL is now set without credentials; auth is supplied per-push via `http.extraheader` injected through `GIT_CONFIG_COUNT/KEY_0/VALUE_0` env vars (same pattern as `actions/checkout`).
- Pin `Dockerfile` runtime base to `alpine:3.21` instead of the mutable `alpine:latest`.
- Pass `BRANCH_NAME` after a `--` separator to `git push` so a value starting with `-` cannot be reinterpreted as a flag.
- Strip backticks from WakaTime project/editor/language names before rendering inside the surrounding `text` code fence to prevent markdown-fence injection from upstream API data.

## [1.5.8] - 2026-05-23

### Added
- `STRIP_EMOJI` env var to suppress emoji from the rendered README block. Values: `NONE` (default, current output), `TITLE` (strip bold-section titles), `CONTENT` (strip stat-line labels inside code fences), `ALL` (strip everywhere including the `Last updated on …` line). Progress-bar glyphs in `PROGRESS_BAR_VERSION=2` are intentionally preserved.

## [1.5.7] - 2026-05-21

### Fixed
- Version the cached WakaTime snapshot independently of the repo-commit cache, so a commit-cache schema upgrade or an `ONLY_MAIN_BRANCH` toggle no longer discards a still-valid WakaTime snapshot.
- Keep freshly fetched WakaTime stats when only the all-time endpoint is still processing, instead of discarding them and falling back entirely to the cached snapshot.

## [1.5.6] - 2026-05-19

### Changed
- Align fixed-width stats table labels with stat block labels so `COMMIT_TIMES_OF_DAY` and other table metrics share the same value column.

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

[Unreleased]: https://github.com/thanhhaudev/github-stats/compare/v1.5.9...HEAD
[1.5.9]: https://github.com/thanhhaudev/github-stats/releases/tag/v1.5.9
[1.5.8]: https://github.com/thanhhaudev/github-stats/releases/tag/v1.5.8
[1.5.7]: https://github.com/thanhhaudev/github-stats/releases/tag/v1.5.7
[1.5.6]: https://github.com/thanhhaudev/github-stats/releases/tag/v1.5.6
[1.5.5]: https://github.com/thanhhaudev/github-stats/releases/tag/v1.5.5
[1.5.4]: https://github.com/thanhhaudev/github-stats/releases/tag/v1.5.4
[1.5.3]: https://github.com/thanhhaudev/github-stats/releases/tag/v1.5.3
