## Quick start

1. Use a profile repo: a repo named exactly `<your-username>/<your-username>`. Its README shows on your profile page.
2. Add markers in that README where stats should land:
   ```markdown
   <!--START_SECTION:readme-stats-->
   <!--END_SECTION:readme-stats-->
   ```
3. Add secrets at **Settings → Secrets and variables → Actions**:
   - `GH_TOKEN` — Personal access token (classic), scopes `repo` + `user`. ([create one](https://github.com/settings/tokens))
   - `WAKATIME_API_KEY` — optional, only for WakaTime metrics. ([get key](https://wakatime.com/settings/api-key))
4. Add `.github/workflows/update-stats.yml`:
   ```yaml
   name: Update README Stats
   on:
     schedule:
       - cron: '0 0 * * *'    # daily at midnight UTC
     workflow_dispatch:
   jobs:
     update:
       runs-on: ubuntu-latest
       permissions:
         contents: write
       steps:
         - uses: actions/checkout@v4
         - uses: thanhhaudev/github-stats@master
           env:
             GITHUB_TOKEN: ${{ secrets.GH_TOKEN }}
             WAKATIME_API_KEY: ${{ secrets.WAKATIME_API_KEY }}
             SHOW_METRICS: "CODING_STREAK"
   ```
5. Trigger once: **Actions → Update README Stats → Run workflow**.

## Sample output

**📈 Coding Streak**
```
🔥 Current Streak:        14 days
🏆 Longest Streak:        45 days
📊 Daily Average:         3 hrs 44 mins
💪 Total Coding Time:     1,383 hrs 16 mins
🎯 Coding Consistency:    87.5%
📅 Active Days:           128 days
```

Every other metric: [docs/metrics.md](docs/metrics.md).

## Metrics

Set `SHOW_METRICS` to a comma-separated list. Output appears in the order you list.

| Key                   | Shows                                                        |
|-----------------------|--------------------------------------------------------------|
| `CODING_STREAK`       | Streak + (with WakaTime) daily-average totals                |
| `COMMIT_TIMES_OF_DAY` | Morning / Daytime / Evening / Night split                    |
| `COMMIT_DAYS_OF_WEEK` | Commits per weekday                                          |
| `LANGUAGE_PER_REPO`   | Primary language per repo                                    |
| `LANGUAGES_AND_TOOLS` | Per-language badges                                          |
| `WAKATIME_AI_STATS`   | AI vs human attribution (needs WakaTime + GenAI integration) |
| `WAKATIME_SPENT_TIME` | Editors / Languages / Projects / OS time                     |

## Required env vars

| Variable           | Purpose                                                                                                                 |
|--------------------|-------------------------------------------------------------------------------------------------------------------------|
| `GITHUB_TOKEN`     | API access. Scopes `repo` + `user`.                                                                                     |
| `SHOW_METRICS`     | Which metrics to render.                                                                                                |
| `WAKATIME_API_KEY` | Required for any `WAKATIME_*` metric and for time fields in `CODING_STREAK`.                                            |
| `WAKATIME_DATA`    | Required if `WAKATIME_SPENT_TIME` is in `SHOW_METRICS`. Any of `EDITORS`, `LANGUAGES`, `PROJECTS`, `OPERATING_SYSTEMS`. |

Full env var list (timezone, cache, commit author, progress-bar style, etc.): [docs/configuration.md](docs/configuration.md).

## More docs

- [Configuration reference](docs/configuration.md) — every env var, progress-bar styles, ready-made configs.
- [Caching](docs/caching.md) — skip API calls for unchanged repos.
- [Running every few minutes](docs/scheduling.md) — cron limits, commit spam, rate budgets.

## Notes

- Reads commit metadata only (timestamps, line counts). Never reads file contents.
- Counts private-repo commits when the token has `repo` scope.
- Works on any repo, not just your profile repo. Override the marker name with `SECTION_NAME`.
