# GitHub README Stats 📊

Automatically update your GitHub profile README with beautiful metrics about your coding activity! This GitHub Action collects data from your repositories and WakaTime, then displays stunning statistics directly on your profile.

[![GitHub Action](https://img.shields.io/badge/GitHub-Action-blue?logo=github)](https://github.com/features/actions)
[![WakaTime](https://img.shields.io/badge/WakaTime-Supported-green?logo=wakatime)](https://wakatime.com)
[![GitHub stars](https://img.shields.io/github/stars/thanhhaudev/github-stats?style=social)](https://github.com/thanhhaudev/github-stats)

## ✨ Features

- 📅 **Commit Patterns** - Visualize when you code most (time of day, day of week)
- 💻 **Language Statistics** - Track programming languages across your repositories
- ⏱️ **WakaTime Integration** - Display coding time, editors, projects and OS usage
- 🎨 **Customizable** - Choose metrics and customize appearance
- 🔄 **Auto-Updating** - Runs on schedule to keep your profile fresh
- 🚀 **Easy Setup** - Get started in 5 minutes

## 🚀 Quick Start

### Step 1: Create Your Profile Repository

Create a repository with the **same name as your GitHub username** (e.g., `username/username`). This special repository's README will appear on your GitHub profile.

> 💡 **Tip:** Don't have a profile repository yet? [Learn more about GitHub profile READMEs](https://docs.github.com/en/account-and-profile/setting-up-and-managing-your-github-profile/customizing-your-profile/managing-your-profile-readme)

### Step 2: Add Markers to Your README

Add these comments to your `README.md` where you want the metrics to appear:

```markdown
<!--START_SECTION:readme-stats-->
<!--END_SECTION:readme-stats-->
```

> 💡 **Tip:** You can customize the section name using the `SECTION_NAME` variable

### Step 3: Get Your Tokens

1. **GitHub Token** (Required)
   - Go to [GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)](https://github.com/settings/tokens)
   - Click "Generate new token (classic)"
   - Select scopes: `repo` and `user`
   - Generate and copy the token

   > 🔒 **Security Note:** The `repo` scope is only used to read commit metadata (timestamps and line changes). Your code is never accessed or stored

2. **WakaTime API Key** (Optional)
   - Optional, only needed if you want to display coding time statistics
   - Get your key from [WakaTime Settings](https://wakatime.com/settings/api-key)

### Step 4: Add Secrets to Your Repository

1. Go to your profile repository's **Settings → Secrets and variables → Actions**
2. Click **New repository secret**
3. Add these secrets:
   - Name: `GH_TOKEN`, Value: Your GitHub token
   - Name: `WAKATIME_API_KEY`, Value: Your WakaTime key (if using WakaTime)

<img width="1128" alt="image" src="https://github.com/user-attachments/assets/40d8c7aa-2c44-40d5-820c-9e93e8637554">

### Step 5: Create the Workflow

Create `.github/workflows/update-stats.yml` in your profile repository:

```yaml
name: Update README Stats

on:
  schedule:
    - cron: '0 0 * * *'  # Runs daily at midnight UTC
  workflow_dispatch:      # Allows manual trigger

jobs:
  update-readme:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Update Stats
        uses: thanhhaudev/github-stats@master
        env:
          GITHUB_TOKEN: ${{ secrets.GH_TOKEN }}
          SHOW_METRICS: "COMMIT_TIMES_OF_DAY,COMMIT_DAYS_OF_WEEK,LANGUAGE_PER_REPO"
```
### Step 6: Trigger the Action

1. Go to the **Actions** tab in your repository
2. Click on **Update README Stats** workflow
3. Click **Run workflow** → **Run workflow**
4. Wait a few seconds and check your README

---

## 📊 Available Metrics

Choose which metrics to display by setting the `SHOW_METRICS` environment variable with a comma-separated list.

**Example:**
```yaml
SHOW_METRICS: "COMMIT_TIMES_OF_DAY,COMMIT_DAYS_OF_WEEK,LANGUAGE_PER_REPO"
```

### 🕒 `COMMIT_TIMES_OF_DAY`

Shows when you code during the day (morning, daytime, evening, night).

**Example output:**

**🕒 I'm An Early Bird 🐤**
```
🌅 Morning                214 commits         ████░░░░░░░░░░░░░░░░░░░░░   17.33%
🌞 Daytime                444 commits         █████████░░░░░░░░░░░░░░░░   35.95%
🌆 Evening                351 commits         ███████░░░░░░░░░░░░░░░░░░   28.42%
🌙 Night                  226 commits         █████░░░░░░░░░░░░░░░░░░░░   18.30%
```

> 💡 **Tip:** Set `SIMPLIFY_COMMIT_TIMES_TITLE: "true"` to show just "I'm An Early 🐤" or "I'm A Night 🦉"


### 📅 `COMMIT_DAYS_OF_WEEK`

Shows which days of the week you're most productive.

**Example output:**

**📅 I'm Most Productive on Sundays**
```
Sunday                   112 commits         ██████░░░░░░░░░░░░░░░░░░░   24.03%
Monday                   57 commits          ███░░░░░░░░░░░░░░░░░░░░░░   12.23%
Tuesday                  58 commits          ███░░░░░░░░░░░░░░░░░░░░░░   12.45%
Wednesday                73 commits          ████░░░░░░░░░░░░░░░░░░░░░   15.67%
Thursday                 94 commits          █████░░░░░░░░░░░░░░░░░░░░   20.17%
Friday                   31 commits          ██░░░░░░░░░░░░░░░░░░░░░░░   06.65%
Saturday                 41 commits          ██░░░░░░░░░░░░░░░░░░░░░░░   08.80%
```

### 🔥 `LANGUAGE_PER_REPO`

Shows the primary programming language distribution across your repositories.

**Example output:**

**🔥 I Mostly Code in Go**
```
Go                       6 repos             █████████████████████░░░░   85.71%
TypeScript               1 repo              ████░░░░░░░░░░░░░░░░░░░░░   14.29%
```

### 💬 `LANGUAGES_AND_TOOLS`

Displays all languages you use with colorful badges showing percentages.

**Example output:**

**💬 Languages & Tools**

![JavaScript](https://img.shields.io/badge/JavaScript-20.0%25-f1e05a?&logo=JavaScript&labelColor=151b23)
![Python](https://img.shields.io/badge/Python-13.0%25-3572A5?&logo=Python&labelColor=151b23)
![Java](https://img.shields.io/badge/Java-12.0%25-b07219?&logo=Java&labelColor=151b23)
![Go](https://img.shields.io/badge/Go-2.8%25-00ADD8?&logo=Go&labelColor=151b23)

### ⏱️ `WAKATIME_SPENT_TIME`

Shows detailed coding activity from WakaTime (requires WakaTime API key).

**Requirements:**
- Set `WAKATIME_API_KEY` with your WakaTime API key
- Set `WAKATIME_DATA` to choose what to display (comma-separated)

**Available data types:**
- `EDITORS` - Which code editors you use
- `LANGUAGES` - Programming languages you code in
- `PROJECTS` - Projects you work on
- `OPERATING_SYSTEMS` - Operating systems you use

**Example configuration:**
```yaml
WAKATIME_API_KEY: ${{ secrets.WAKATIME_API_KEY }}
WAKATIME_DATA: "EDITORS,LANGUAGES,PROJECTS,OPERATING_SYSTEMS"
WAKATIME_RANGE: "last_7_days"
```

**Example output:**
```
📝 Editors:
PhpStorm                 42 hrs 14 mins      ███████████████████████░░   93.02%
GoLand                   3 hrs 10 mins       ██░░░░░░░░░░░░░░░░░░░░░░░   06.98%

💬 Languages:
Go                       22 hrs 19 mins      ████████████░░░░░░░░░░░░░   49.16%
JavaScript               14 hrs 41 mins      ████████░░░░░░░░░░░░░░░░░   32.34%
Python                   1 hr 53 mins        █░░░░░░░░░░░░░░░░░░░░░░░░   04.18%

📦 Projects:
Project A                6 hrs 47 mins       ███████████████████░░░░░░   77.43%
Project B                1 hr 35 mins        ████░░░░░░░░░░░░░░░░░░░░░   18.07%
Project C                23 mins             █░░░░░░░░░░░░░░░░░░░░░░░░   04.49%

💻 Operating Systems:
Windows                  42 hrs 14 mins      ████████████████████░░░░░   70.00%
Mac                      12 hrs 10 mins      ███████░░░░░░░░░░░░░░░░░░   20.00%
Linux                    6 hrs  3 mins       ████░░░░░░░░░░░░░░░░░░░░░   10.00%
```

**Time range options** (set with `WAKATIME_RANGE`):

| Value           | Title Displayed    |
|-----------------|--------------------|
| `last_7_days`   | 📅 Last 7 Days     |
| `last_30_days`  | 📊 Last 30 Days    |
| `last_6_months` | 📈 Last 6 Months   |
| `last_year`     | 🗓️ Last 12 Months |
| `all_time`      | ⏱️ All Time        |

---

## ⚙️ Configuration

### Environment Variables
| Variable                      | Description                                                                                      | Required               | Default                     |
|-------------------------------|--------------------------------------------------------------------------------------------------|------------------------|-----------------------------|
| `GITHUB_TOKEN`                | GitHub token for API access                                                                      | ✅ Yes                  | -                           |
| `SHOW_METRICS`                | Comma-separated list of metrics to display                                                       | ✅ Yes                  | -                           |
| `WAKATIME_API_KEY`            | WakaTime API key for coding stats                                                                | ❌ No                   | -                           |
| `WAKATIME_DATA`               | WakaTime data to show: `EDITORS`, `LANGUAGES`, `PROJECTS`, `OPERATING_SYSTEMS` (comma-separated) | Only if using WakaTime | -                           |
| `WAKATIME_RANGE`              | Time range: `last_7_days`, `last_30_days`, `last_6_months`, `last_year`, `all_time`              | ❌ No                   | `last_7_days`               |
| `SHOW_LAST_UPDATE`            | Show last update timestamp in README                                                             | ❌ No                   | `false`                     |
| `TIME_ZONE`                   | Timezone for statistics (e.g., `America/New_York`, `Asia/Tokyo`)                                 | ❌ No                   | `UTC`                       |
| `TIME_LAYOUT`                 | Go time format layout for timestamps                                                             | ❌ No                   | `2006-01-02 15:04:05 -0700` |
| `ONLY_MAIN_BRANCH`            | Only count commits from main branch (faster performance)                                         | ❌ No                   | `false`                     |
| `EXCLUDE_FORK_REPOS`          | Exclude forked repositories from metrics                                                         | ❌ No                   | `false`                     |
| `SECTION_NAME`                | Custom section name for README markers                                                           | ❌ No                   | `readme-stats`              |
| `COMMIT_MESSAGE`              | Custom commit message when updating README                                                       | ❌ No                   | `📝 Update README.md`       |
| `COMMIT_USER_NAME`            | Git commit author name                                                                           | ❌ No                   | `GitHub Action`             |
| `COMMIT_USER_EMAIL`           | Git commit author email                                                                          | ❌ No                   | `action@github.com`         |
| `PROGRESS_BAR_VERSION`        | Progress bar style: `1` (blocks) or `2` (emoji squares)                                          | ❌ No                   | `1`                         |
| `SIMPLIFY_COMMIT_TIMES_TITLE` | Show simplified title: "I'm An Early 🐤" or "I'm A Night 🦉"                                     | ❌ No                   | `false`                     |
| `HIDE_REPO_INFO`              | Hide repository information in action logs                                                       | ❌ No                   | `false`                     |

### 🎨 Progress Bar Styles

You can customize the appearance of progress bars using `PROGRESS_BAR_VERSION`:

**Version 1** (Default) - Block style:
```
████░░░░░░░░░░░░░░░░░░░░░
```

**Version 2** - Emoji squares with half-block support:
```
🟩🟩🟩🟩🟩🟩🟩🟩🟨⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜
```

### 🛠 Example Configurations

**Minimal Setup** (GitHub stats only):
```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GH_TOKEN }}
  SHOW_METRICS: "COMMIT_TIMES_OF_DAY,COMMIT_DAYS_OF_WEEK,LANGUAGE_PER_REPO"
```

**Full Setup** (with WakaTime):
```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GH_TOKEN }}
  WAKATIME_API_KEY: ${{ secrets.WAKATIME_API_KEY }}
  WAKATIME_DATA: "EDITORS,LANGUAGES,PROJECTS,OPERATING_SYSTEMS"
  WAKATIME_RANGE: "last_30_days"
  SHOW_METRICS: "COMMIT_TIMES_OF_DAY,COMMIT_DAYS_OF_WEEK,LANGUAGE_PER_REPO,LANGUAGES_AND_TOOLS,WAKATIME_SPENT_TIME"
  SHOW_LAST_UPDATE: "true"
  ONLY_MAIN_BRANCH: "true"
  PROGRESS_BAR_VERSION: "2"
```

**Performance Optimized**:
```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GH_TOKEN }}
  SHOW_METRICS: "COMMIT_TIMES_OF_DAY,LANGUAGE_PER_REPO"
  ONLY_MAIN_BRANCH: "true"  # Faster - only scans main branch
  EXCLUDE_FORK_REPOS: "true"  # Skip forked repositories
  HIDE_REPO_INFO: "true"  # Cleaner logs
```
---

## 📝 FAQ

<details>
<summary><b>Can I use this on a regular repository (not my profile)?</b></summary>

Yes! You can use this action on any repository. Just add the markers to any markdown file and configure the workflow accordingly.

</details>

<details>
<summary><b>How often does it update?</b></summary>

By default, the workflow runs daily at midnight UTC (configured with `cron: '0 0 * * *'`). You can change this schedule or trigger it manually anytime.

</details>

<details>
<summary><b>Does this count private repository commits?</b></summary>

Yes, if your GitHub token has access to private repositories (which it does with the `repo` scope), it will count commits from private repos too.

</details>

<details>
<summary><b>Can I customize the appearance?</b></summary>

Yes! You can:
- Choose which metrics to display with `SHOW_METRICS`
- Change progress bar style with `PROGRESS_BAR_VERSION`
- Simplify titles with `SIMPLIFY_COMMIT_TIMES_TITLE`
- Set custom timezone with `TIME_ZONE`

</details>

<details>
<summary><b>Is my data safe?</b></summary>

Absolutely! This action:
- Only reads commit metadata (timestamps, line counts)
- Never accesses your actual code
- Runs in your own GitHub Actions environment
- Doesn't send data to any external services (except WakaTime API if you enable it)

</details>
