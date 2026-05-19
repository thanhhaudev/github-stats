# Configuration

## Environment variables

| Variable                      | Description                                                                                                                    | Default                     |
|-------------------------------|--------------------------------------------------------------------------------------------------------------------------------|-----------------------------|
| `GITHUB_TOKEN`                | **Required.** GitHub API token. Scope `repo`.                                                                                  | —                           |
| `SHOW_METRICS`                | **Required.** Comma-separated list of metrics. See [metrics.md](metrics.md).                                                   | —                           |
| `WAKATIME_API_KEY`            | Required for `WAKATIME_*` metrics and time fields in `CODING_STREAK`.                                                          | —                           |
| `WAKATIME_DATA`               | Required if `WAKATIME_SPENT_TIME` is in `SHOW_METRICS`. Comma list of `EDITORS`, `LANGUAGES`, `PROJECTS`, `OPERATING_SYSTEMS`. | —                           |
| `WAKATIME_RANGE`              | `last_7_days`, `last_30_days`, `last_6_months`, `last_year`, `all_time`.                                                       | `last_7_days`               |
| `TIME_ZONE`                   | IANA timezone (e.g. `Asia/Ho_Chi_Minh`). Used for streak day boundaries and `SHOW_LAST_UPDATE`.                                | `UTC`                       |
| `TIME_LAYOUT`                 | Go time layout for `SHOW_LAST_UPDATE`.                                                                                         | `2006-01-02 15:04:05 -0700` |
| `SHOW_LAST_UPDATE`            | Append a timestamp line to the rendered block.                                                                                 | `false`                     |
| `ONLY_MAIN_BRANCH`            | Count commits only from each repo's default branch. Faster.                                                                    | `false`                     |
| `EXCLUDE_FORK_REPOS`          | Skip forked repos.                                                                                                             | `false`                     |
| `BRANCH_NAME`                 | Branch to push README updates to.                                                                                              | `main`                      |
| `SECTION_NAME`                | Marker name. Markers become `<!--START_SECTION:<name>-->` and `<!--END_SECTION:<name>-->`.                                     | `readme-stats`              |
| `PROGRESS_BAR_VERSION`        | `1` (block chars) or `2` (emoji squares).                                                                                      | `1`                         |
| `SIMPLIFY_COMMIT_TIMES_TITLE` | Shorten `COMMIT_TIMES_OF_DAY` title.                                                                                           | `false`                     |
| `SIMPLE_LOGS`                 | Show only high-level step logs. Useful for public repos where you want less noisy action output.                               | `false`                     |
| `COMMIT_MESSAGE`              | Commit message used when pushing the README.                                                                                   | `📝 Update README.md`       |
| `COMMIT_USER_NAME`            | Git author name.                                                                                                               | `GitHub Action`             |
| `COMMIT_USER_EMAIL`           | Git author email.                                                                                                              | `action@github.com`         |
| `HIDE_REPO_INFO`              | Strip repo names and tokens from action logs.                                                                                  | `false`                     |
| `DRY_RUN`                     | Update the README file without committing or pushing changes.                                                                  | `false`                     |
| `DEBUG`                       | Verbose logs (full GraphQL errors).                                                                                            | `false`                     |
| `ENABLE_CACHE`                | Reuse cached commits between runs. See [caching.md](caching.md).                                                               | `false`                     |
| `CACHE_FILE`                  | Cache file path. Must match the `path` in `actions/cache@v4`.                                                                  | `.github-stats-cache.json`  |

## Progress bar styles

`PROGRESS_BAR_VERSION: "1"` (default):
```
████░░░░░░░░░░░░░░░░░░░░░
```

`PROGRESS_BAR_VERSION: "2"` (emoji squares with half-block support):
```
🟩🟩🟩🟩🟩🟩🟩🟩🟨⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜
```

## Ready-made configs

**Minimal** (GitHub-only):
```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GH_TOKEN }}
  SHOW_METRICS: "COMMIT_TIMES_OF_DAY,COMMIT_DAYS_OF_WEEK,LANGUAGE_PER_REPO"
```

**Full** (with WakaTime):
```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GH_TOKEN }}
  WAKATIME_API_KEY: ${{ secrets.WAKATIME_API_KEY }}
  WAKATIME_DATA: "EDITORS,LANGUAGES,PROJECTS,OPERATING_SYSTEMS"
  WAKATIME_RANGE: "last_30_days"
  SHOW_METRICS: "COMMIT_TIMES_OF_DAY,COMMIT_DAYS_OF_WEEK,LANGUAGE_PER_REPO,LANGUAGES_AND_TOOLS,WAKATIME_SPENT_TIME,CODING_STREAK,WAKATIME_AI_STATS"
  SHOW_LAST_UPDATE: "true"
  ONLY_MAIN_BRANCH: "true"
  PROGRESS_BAR_VERSION: "2"
```

**Performance-optimized** (for large repo lists):
```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GH_TOKEN }}
  SHOW_METRICS: "COMMIT_TIMES_OF_DAY,LANGUAGE_PER_REPO"
  ONLY_MAIN_BRANCH: "true"
  EXCLUDE_FORK_REPOS: "true"
  HIDE_REPO_INFO: "true"
```

**Public-friendly logs** (less noisy output):
```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GH_TOKEN }}
  SHOW_METRICS: "COMMIT_TIMES_OF_DAY,LANGUAGE_PER_REPO"
  SIMPLE_LOGS: "true"
  HIDE_REPO_INFO: "true"
```
