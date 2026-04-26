# Running every few minutes

For runs more frequent than once a day, the bottlenecks aren't the GitHub API (caching handles that) — they are README commit spam, WakaTime rate limits, and overlapping runs racing on git push.

## Recommended workflow

```yaml
name: Update README Stats

on:
  schedule:
    - cron: '0 * * * *'   # hourly
  workflow_dispatch:

concurrency:
  group: github-stats
  cancel-in-progress: false   # serialize overlapping runs instead of racing

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

      - uses: thanhhaudev/github-stats@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GH_TOKEN }}
          WAKATIME_API_KEY: ${{ secrets.WAKATIME_API_KEY }}
          SHOW_METRICS: "COMMIT_TIMES_OF_DAY,LANGUAGE_PER_REPO,CODING_STREAK"
          ENABLE_CACHE: "true"
          SHOW_LAST_UPDATE: "false"   # critical — see below
          HIDE_REPO_INFO: "true"
```

## Why each setting matters

| Setting                     | Why                                                                                                                                                                                                                           |
|-----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `concurrency.group`         | If two runs overlap, GitHub queues the second. Avoids races on cache writes and git push conflicts.                                                                                                                           |
| `cancel-in-progress: false` | Lets in-flight runs finish so their cache updates aren't wasted. Use `true` if you'd rather always run the latest.                                                                                                            |
| `ENABLE_CACHE: "true"`      | Reuses commits from previous runs. Without it, every run re-fetches everything.                                                                                                                                               |
| `SHOW_LAST_UPDATE: "false"` | **Most important.** With it on, the timestamp changes every run, so the action commits + pushes every run. Hourly = 24 commits/day of `📝 Update README.md`. With it off, the action only commits when stats actually change. |

## Cadence guide

| Cadence           | Cron           | Verdict                                                     |
|-------------------|----------------|-------------------------------------------------------------|
| Daily             | `0 0 * * *`    | Default. No concerns.                                       |
| Hourly            | `0 * * * *`    | Sweet spot.                                                 |
| Every 15 min      | `*/15 * * * *` | Fine with the config above.                                 |
| Every 5 min       | `*/5 * * * *`  | GitHub's cron minimum. May be delayed under load.           |
| Faster than 5 min | n/a            | Not supported by GitHub cron. Trigger externally if needed. |

## Limits

- **GitHub cron**: best-effort, ~5 min minimum, may be delayed under heavy load.
- **WakaTime API**: ~60 req/min. Each run uses 2 requests. Per-minute runs (via external trigger) stay safe.
- **GitHub primary rate limit**: 5,000 GraphQL points/hour. With cache warm, each run uses ~10–20 points.
- **Actions cache storage**: 10 GB per repo, LRU-evicted automatically.
