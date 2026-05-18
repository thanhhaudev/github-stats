# Reliability Debt Design

## Goal

Fix the highest-risk reliability debt in the current branch in a sequence that keeps each change reviewable and testable.

## Scope

This work covers five areas:

1. Make `pkg/container.DataContainer.InitCommits` concurrency deterministic.
2. Make the `clock.Clock` dependency explicit instead of requiring a context value.
3. Fix Unicode-safe truncation in `pkg/writer`.
4. Add focused tests for production paths in `pkg/container`, `pkg/github`, and `cmd`.
5. Add parity tests that catch drift between config constants, `action.yml`, and docs.

## Design

### InitCommits Concurrency

`InitCommits` should treat a repository as the unit of concurrency. Each repository worker returns exactly one result: either a commit slice or an error. If all branches are fetched, branch workers are internal to the repository worker and their errors are collected there. Branch goroutines must not write to the top-level result channel directly.

The implementation may use `golang.org/x/sync/errgroup` because it models "cancel siblings on first error" cleanly. The top-level function should return the first actionable error with enough repo/branch context to debug the failure.

### Clock Dependency

`DataContainer` should own a `clock.Clock` field. `NewDataContainer` should continue to be usable from tests without extra arguments by defaulting to UTC, and a setter or constructor option should allow `cmd/main.go` to inject the configured clock.

`InitCommits` should call `d.Clock.ToClockTz(...)` instead of reading `clock.ClockKey{}` from context. `withClock` can remain for now if other code still uses it, but `InitCommits` must not depend on it.

### Unicode Truncation

`truncateString` should never return invalid UTF-8. Because the writer already computes display width for alignment, truncation should cap display width rather than byte length. This preserves table layout better for CJK, emoji, and accented text than rune-count truncation alone.

### Coverage

Coverage work should focus on behavior with production risk:

- Pagination continues until `HasNextPage` is false and passes `afterCursor`.
- API and repository errors return useful errors without losing context.
- WakaTime pending/not-ready responses restore cached data instead of failing the run.
- README update replaces only the configured marker section.
- Git helpers sanitize credentials and do not push unsafe cache files.

The target is meaningful behavioral coverage, not a fixed percentage.

### Config Parity

Add tests that fail clearly when a config key or metric key appears in code but is missing from user-facing metadata/docs. The first pass should be guard tests, not schema generation. Schema generation can be a later refactor if drift remains painful.

## Verification

Each task should run the narrow test that proves the behavior and then `go test ./...`. Concurrency changes should also be checked with `go test -race ./...` before final completion.
