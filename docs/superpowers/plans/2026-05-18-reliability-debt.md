# Reliability Debt Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove concurrency, clock, Unicode, coverage, and config drift risks identified in review.

**Architecture:** Apply the fixes in small TDD loops. Keep repository-level concurrency as the public unit in `InitCommits`, make `clock.Clock` an explicit `DataContainer` dependency, fix writer truncation at the helper level, then add guard coverage around production paths and config parity.

**Tech Stack:** Go 1.24, standard `testing`, optional `golang.org/x/sync/errgroup`, existing project packages.

---

### Task 1: Make InitCommits Repository Workers Return One Result

**Files:**
- Modify: `pkg/container/container.go`
- Test: `pkg/container/container_test.go`
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: Write failing test for branch error handling**

Add a test that sets up one repository with multiple branches and a commit fetch error on one branch. The test should call `InitCommits` and assert it returns the branch error instead of blocking or swallowing it.

- [ ] **Step 2: Run the narrow test**

Run: `go test ./pkg/container -run TestDataContainerInitCommitsReturnsBranchError`

Expected: FAIL before implementation.

- [ ] **Step 3: Implement repository-scoped result handling**

Refactor `InitCommits` so each repo worker sends one `commitResult{commits []github.Commit, err error}`. Use `errgroup.WithContext` inside all-branches mode so branch fetches cancel on first branch error and the repo worker returns that error.

- [ ] **Step 4: Verify**

Run: `go test ./pkg/container -run TestDataContainerInitCommitsReturnsBranchError`

Expected: PASS.

### Task 2: Make Clock Dependency Explicit

**Files:**
- Modify: `pkg/container/container.go`
- Modify: `cmd/main.go`
- Test: `pkg/container/container_test.go`

- [ ] **Step 1: Write failing test for missing context clock**

Add a test that calls `InitCommits(context.Background())` with a successful commit fetch and asserts it does not panic.

- [ ] **Step 2: Run the narrow test**

Run: `go test ./pkg/container -run TestDataContainerInitCommitsDoesNotRequireContextClock`

Expected: FAIL before implementation.

- [ ] **Step 3: Add explicit clock field**

Add `Clock clock.Clock` to `DataContainer`, default it to `clock.NewClock()` in `NewDataContainer`, add `SetClock(cl clock.Clock)`, and update `cmd/main.go` to inject the configured clock. Use `d.Clock.ToClockTz` in `InitCommits`.

- [ ] **Step 4: Verify**

Run: `go test ./pkg/container -run TestDataContainerInitCommitsDoesNotRequireContextClock`

Expected: PASS.

### Task 3: Fix Unicode Truncation

**Files:**
- Modify: `pkg/writer/writer.go`
- Test: `pkg/writer/writer_test.go`

- [ ] **Step 1: Write failing Unicode test**

Add a test for `truncateString` with a long accented/CJK/emoji string. Assert the result is valid UTF-8 and `displayWidth(result) <= limit`.

- [ ] **Step 2: Run the narrow test**

Run: `go test ./pkg/writer -run TestTruncateStringPreservesUTF8`

Expected: FAIL before implementation.

- [ ] **Step 3: Implement display-width truncation**

Update `truncateString` to iterate over runes and stop before exceeding the requested display width.

- [ ] **Step 4: Verify**

Run: `go test ./pkg/writer -run TestTruncateStringPreservesUTF8`

Expected: PASS.

### Task 4: Add Production Path Coverage

**Files:**
- Modify: `pkg/container/manager_test.go`
- Modify: `pkg/container/wakatime_cache_test.go`
- Modify: `cmd/main_test.go`
- Modify: `cmd/git_test.go`

- [ ] **Step 1: Add focused tests**

Cover pagination cursor behavior, WakaTime pending cache fallback, README marker replacement, and git/cache safety behavior that is not already covered.

- [ ] **Step 2: Verify**

Run: `go test ./pkg/container ./cmd`

Expected: PASS.

### Task 5: Add Config/Action/Docs Parity Guard

**Files:**
- Modify: `pkg/config/config_test.go`

- [ ] **Step 1: Add parity tests**

Add tests that scan `action.yml`, `docs/configuration.md`, and `docs/metrics.md` for required env/metric keys already defined in `pkg/config`.

- [ ] **Step 2: Verify**

Run: `go test ./pkg/config`

Expected: PASS.

### Task 6: Final Verification

**Files:**
- No code edits.

- [ ] **Step 1: Run full tests**

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 2: Run race tests**

Run: `go test -race ./...`

Expected: PASS.
