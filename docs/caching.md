# Caching

If you have many repos, runs may approach the GitHub API rate limit. Caching skips re-fetching commits for repos that haven't been pushed to since the last run.

## How to enable

Add an `actions/cache@v4` step **before** the action and set `ENABLE_CACHE: "true"`:

```yaml
jobs:
  update:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4

      - name: Restore cache
        uses: actions/cache@v4
        with:
          path: .github-stats-cache.json
          key: github-stats-${{ github.run_id }}
          restore-keys: github-stats-

      - uses: thanhhaudev/github-stats@master
        env:
          GITHUB_TOKEN: ${{ secrets.GH_TOKEN }}
          SHOW_METRICS: "COMMIT_TIMES_OF_DAY,LANGUAGE_PER_REPO"
          ENABLE_CACHE: "true"
```

> ⚠️ Add `.github-stats-cache.json` to your profile repo's `.gitignore`. The cache holds repo URLs (including private) and commit metadata — fine inside GitHub Actions cache, not fine inside your public profile repo.

## How it works

- The action stores fetched repo metadata + commits in `CACHE_FILE` (`.github-stats-cache.json` by default).
- Each run queries every repo's `pushedAt`. Unchanged repos reuse cached commits and skip the API calls.
- Cached repos that no longer exist (deleted, transferred) are pruned automatically.
- The cache schema is versioned. Schema upgrades invalidate the cache.
- Toggling `ONLY_MAIN_BRANCH` invalidates the cache (the two modes return different commit sets).

## Trade-offs

- GitHub evicts caches after 7 days of inactivity.
- The first run after a cache miss is as slow as today — caching only helps subsequent runs.
- WakaTime stats are not cached (they are aggregates over a time range, not incrementally fetchable).

## Security

GitHub Actions cache is scoped to the repo and requires authenticated access. It is not publicly readable, even on public repos. Workflows from forked PRs cannot access the cache (GitHub enforces this).
