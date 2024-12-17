# GitHub README.md Metrics 📊

This GitHub Action collects your GitHub data and coding activity from WakaTime. It updates a part of your README.md file with cool metrics from this data and then commits and pushes the changes to your repository.

## Prerequisites

1. **Special Repository**: To make this Action work, you need a **special repository** with the same name as your GitHub username (e.g., `username/username`). This is a special repository on GitHub where the README will show up on your profile.


2. **Update the Markdown File**: You need to add two special comments to your Markdown (.md) file. These comments will be used to update the file with the metrics. You can add these comments anywhere in the file, but it's recommended to add them at the end of the file.

   ```markdown
   <!--START_SECTION:readme-stats-->
   <!--END_SECTION:readme-stats-->
   ```

   The Action will replace everything between these two comments with the metrics. You can also specify a section name in the `SECTION_NAME` environment variable.


3. **GitHub Access Token**: To get commit information, you need a GitHub Access Token with `repo` and `user` permissions, available [here](https://github.com/settings/tokens). 
   >Although giving `repo` access might seem **risky**, this Action only accesses commit timestamps and lines of code added or deleted in the repositories you contributed to, which is completely safe.

4. **WakaTime API Key (Optional)**: If you want to use the `WAKATIME_SPENT_TIME` metric, you will need a WakaTime API Key. You can get this from your [WakaTime Account Settings](https://wakatime.com/settings/api-key).


5. **Store Access Keys in Secrets**: You need to store the WakaTime API Key and GitHub Access Token in your repository secrets. You can find this option in the **Settings** of your GitHub repository. Make sure to save them under the following names:

   + **WAKATIME_API_KEY**: Your WakaTime API Key.
   + **GH_TOKEN**: Your GitHub Access Token.

<img width="1128" alt="image" src="https://github.com/user-attachments/assets/40d8c7aa-2c44-40d5-820c-9e93e8637554">


## Usage

### Environment Variables
| Name                          | Description                                                                                                                                                                                      | Required                          | Default                   |
|-------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------|---------------------------|
| `GITHUB_TOKEN`                | The GitHub token to authenticate API requests.                                                                                                                                                   | Yes                               | -                         |
| `SHOW_METRICS`                | The metrics to show in the `README.md` file.                                                                                                                                                     | Yes                               | -                         |
| `WAKATIME_API_KEY`            | The WakaTime API key to fetch coding activity statistics.                                                                                                                                        | No                                | -                         |
| `WAKATIME_RANGE`              | The range for WakaTime statistics (e.g., `last_7_days`, `last_30_days`, `last_6_months`, `last_year`, `all_time`).                                                                               | No                                | last_7_days               |
| `WAKATIME_DATA`               | The data to show from WakaTime statistics.                                                                                                                                                       | If `WAKATIME_API_KEY` is provided | -                         |
| `TIME_ZONE`                   | The timezone to use for statistics.                                                                                                                                                              | No                                | UTC                       |
| `TIME_LAYOUT`                 | The layout of the time to show in the last update time.                                                                                                                                          | No                                | 2006-01-02 15:04:05 -0700 |
| `SHOW_LAST_UPDATE`            | Whether to show the last update time in the `README.md` file.                                                                                                                                    | No                                | -                         |
| `ONLY_MAIN_BRANCH`            | Whether to fetch data only from the main branch. If you don’t set this, it will search for commits in all branches of the repository to count the number of commits, which might take more time. | No                                | -                         |
| `COMMIT_MESSAGE`              | The commit message to use when updating the `README.md`.                                                                                                                                         | No                                | 📝 Update README.md       |
| `COMMIT_USER_NAME`            | The name to use for the commit.                                                                                                                                                                  | No                                | GitHub Action             |
| `COMMIT_USER_EMAIL`           | The email to use for the commit.                                                                                                                                                                 | No                                | action@github.com         |
| `SECTION_NAME`                | The section name in the `README.md` to update.                                                                                                                                                   | No                                | readme-stats              |
| `HIDE_REPO_INFO`              | Whether to hide the repository information in action logs.                                                                                                                                       | No                                | -                         |
| `PROGRESS_BAR_VERSION`        | The version of the progress bar to use.                                                                                                                                                          | No                                | 1                         |
| `EXCLUDE_FORK_REPOS`          | Whether to exclude fork repositories from the metrics.                                                                                                                                           | No                                | -                         |
| `LANGUAGES_AND_TOOLS`         | The languages and tools that you used in your repositories.                                                                                                                                      | No                                | -                         |
| `SIMPLIFY_COMMIT_TIMES_TITLE` | If you want to display a simplified title when using `COMMIT_TIMES_OF_DAY`, enable this option to show either "I'm An Early 🐤" or "I'm A Night 🦉" based on the commit time.                    | No                                | -                         |

### Metrics
The `SHOW_METRICS` environment variable is used to specify the metrics to show in the `README.md` file. You can choose from the following metrics:

**COMMIT_TIMES_OF_DAY**: The distribution of your commits across different times of the day, such as morning, daytime, evening, and night.

   **🕒 I'm An Afternoon Warrior 🥷🏻**
   ```
    🌅 Morning                136 commits         ███████░░░░░░░░░░░░░░░░░░   29.18%
    🌞 Daytime                265 commits         ██████████████░░░░░░░░░░░   56.87%
    🌆 Evening                1 commit            ░░░░░░░░░░░░░░░░░░░░░░░░░   00.21%
    🌙 Night                  64 commits          ███░░░░░░░░░░░░░░░░░░░░░░   13.73%
   ```

**COMMIT_DAYS_OF_WEEK**: The days of the week you make commits.

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

**LANGUAGE_PER_REPO**: The languages you use in each repository.

   **🔥 I Mostly Code in Go**
   ```
   Go                       6 repos             █████████████████████░░░░   85.71%
   TypeScript               1 repo              ████░░░░░░░░░░░░░░░░░░░░░   14.29%
   ```

**LANGUAGES_AND_TOOLS**: The languages and tools you used on your projects.

   **💬 Languages & Tools** 

![JavaScript](https://img.shields.io/badge/JavaScript-20.0%25-f1e05a?&logo=JavaScript&labelColor=151b23)
![Python](https://img.shields.io/badge/Python-13.0%25-3572A5?&logo=Python&labelColor=151b23)
![Java](https://img.shields.io/badge/Java-12.0%25-b07219?&logo=Java&labelColor=151b23)
![C#](https://img.shields.io/badge/C%23-9.4%25-178600?&logo=CSharp&labelColor=151b23)
![PHP](https://img.shields.io/badge/PHP-7.8%25-4F5D95?&logo=PHP&labelColor=151b23)
![C++](https://img.shields.io/badge/C++-7.0%25-00599C?&logo=Cplusplus&labelColor=151b23)
![TypeScript](https://img.shields.io/badge/TypeScript-6.3%25-3178C6?&logo=TypeScript&labelColor=151b23)
![Ruby](https://img.shields.io/badge/Ruby-3.1%25-701516?&logo=Ruby&labelColor=151b23)
![Swift](https://img.shields.io/badge/Swift-2.6%25-FA7343?&logo=Swift&labelColor=151b23)
![Go](https://img.shields.io/badge/Go-2.8%25-00ADD8?&logo=Go&labelColor=151b23)

**WAKATIME_SPENT_TIME**: The time you spent coding on WakaTime. 

Use the `WAKATIME_DATA` environment variable to specify the data to show.
+ **EDITORS**: The editors you use.
   ```
   📝 Editors:
   PhpStorm                 42 hrs 14 mins      ███████████████████████░░   93.02%
   GoLand                   3 hrs 10 mins       ██░░░░░░░░░░░░░░░░░░░░░░░   06.98%
   ```
+ **LANGUAGES**: The languages you code in.
   ```
  💬 Languages:
   Go                       22 hrs 19 mins      ████████████░░░░░░░░░░░░░   49.16%
   JavaScript               14 hrs 41 mins      ████████░░░░░░░░░░░░░░░░░   32.34%
   Python                   1 hr 53 mins        █░░░░░░░░░░░░░░░░░░░░░░░░   04.18%
   Java                     1 hr 27 mins        █░░░░░░░░░░░░░░░░░░░░░░░░   03.20%
   C++                      1 hr 15 mins        █░░░░░░░░░░░░░░░░░░░░░░░░   02.78%
   Ruby                     1 hr 12 mins        █░░░░░░░░░░░░░░░░░░░░░░░░   02.66%
   PHP                      52 mins             ░░░░░░░░░░░░░░░░░░░░░░░░░   01.92%
   TypeScript               43 mins             ░░░░░░░░░░░░░░░░░░░░░░░░░   01.58%
   Swift                    22 mins             ░░░░░░░░░░░░░░░░░░░░░░░░░   00.81%
   Rust                     15 mins             ░░░░░░░░░░░░░░░░░░░░░░░░░   00.56%
   Others                   16 mins             ░░░░░░░░░░░░░░░░░░░░░░░░░   00.80%
   ```
+ **PROJECTS**: The projects you work on.
   ```
  📦 Projects:
   Project A                6 hrs 47 mins       ███████████████████░░░░░░   77.43%
   Project B                1 hr 35 mins        ████░░░░░░░░░░░░░░░░░░░░░   18.07%
   Project C                23 mins             █░░░░░░░░░░░░░░░░░░░░░░░░   4.49%
  ```
+ **OPERATING_SYSTEMS**: The operating systems you use.
  ```
  💻 Operating Systems:
  Windows                  42 hrs 14 mins      ████████████████████░░░░░   70.00%
  Mac                      12 hrs 10 mins      ███████░░░░░░░░░░░░░░░░░░   20.00%
  Linux                    6 hrs  3 mins       ████░░░░░░░░░░░░░░░░░░░░░   10.00%
  ```

You can use the `WAKATIME_RANGE` environment variable to set the time range for WakaTime statistics. Each value will show a specific label as follows:
+ `last_7_days`: What I Focused On in the Last 7 Days
+ `last_30_days`: How I Spent My Time Over the Last 30 Days
+ `last_6_months`: Where My Time Went in the Last 6 Months
+ `last_year`: My Time Highlights from Last Year
+ `all_time`: How I’ve Used My Time Across All Time

**Note**: If you don't provide the `WAKATIME_API_KEY`, the `WAKATIME_SPENT_TIME` metric will not be shown.

### Progress Bar Versions
You can use the `PROGRESS_BAR_VERSION` environment variable to specify the version of the progress bar to use. The available versions are:
+ `1`: **Default Progress Bar**: Uses the default progress bar style.
```
████░░░░░░░░░░░░░░░░░░░░░
```
+ `2` **Square Symbol Progress Bar**: Uses the square symbol for the progress bar. This version also shows the half block (the remaining percentage is not enough to fill a whole block) for the progress bar.
```
🟩🟩🟩🟩🟩🟩🟩🟩🟨⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜
```

### Example Workflow

```yaml
name: Update README.md

on:
  schedule:
    - cron: '0 0 * * *' # Runs every day at midnight
  workflow_dispatch:
jobs:
  update-readme:
    name: Update README.md
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Run GitHub Stats Action
        uses: thanhhaudev/github-stats@master
        env:
          BRANCH_NAME: ${{ github.ref_name }}
          GITHUB_TOKEN: ${{ secrets.GH_TOKEN }} # GitHub token, required
          WAKATIME_API_KEY: ${{ secrets.WAKATIME_API_KEY }}
          WAKATIME_DATA: "EDITORS,LANGUAGES,PROJECTS,OPERATING_SYSTEMS" # show data, separated by comma
          SHOW_METRICS: "COMMIT_TIMES_OF_DAY,LANGUAGE_PER_REPO,COMMIT_DAYS_OF_WEEK,WAKATIME_SPENT_TIME" # show metrics, separated by comma
          SHOW_LAST_UPDATE: "true" # show last update time
          ONLY_MAIN_BRANCH: "true" # only fetch data from the main branch
```

---
> This project is inspired by [waka-readme-stats](https://github.com/anmol098/waka-readme-stats), which is a similar project that uses GitHub Actions to update your README with awesome metrics.
