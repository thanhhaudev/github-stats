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
| Name                | Description                                                                                                                                                                                      | Required                          | Default                   |
|---------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------|---------------------------|
| `GITHUB_TOKEN`      | The GitHub token to authenticate API requests.                                                                                                                                                   | Yes                               | -                         |
| `SHOW_METRICS`      | The metrics to show in the `README.md` file.                                                                                                                                                     | Yes                               | -                         |
| `WAKATIME_API_KEY`  | The WakaTime API key to fetch coding activity statistics.                                                                                                                                        | No                                | -                         |
| `WAKATIME_RANGE`    | The range for WakaTime statistics (e.g., `last_7_days`, `last_30_days`, `last_6_months`, `last_year`, `all_time`).                                                                               | No                                | last_7_days               |
| `WAKATIME_DATA`     | The data to show from WakaTime statistics.                                                                                                                                                       | If `WAKATIME_API_KEY` is provided | -                         |
| `TIME_ZONE`         | The timezone to use for statistics.                                                                                                                                                              | No                                | UTC                       |
| `TIME_LAYOUT`       | The layout of the time to show in the last update time.                                                                                                                                          | No                                | 2006-01-02 15:04:05 -0700 |
| `SHOW_LAST_UPDATE`  | Whether to show the last update time in the `README.md` file.                                                                                                                                    | No                                | -                         |
| `ONLY_MAIN_BRANCH`  | Whether to fetch data only from the main branch. If you don’t set this, it will search for commits in all branches of the repository to count the number of commits, which might take more time. | No                                | -                         |
| `COMMIT_MESSAGE`    | The commit message to use when updating the `README.md`.                                                                                                                                         | No                                | 📝 Update README.md       |
| `COMMIT_USER_NAME`  | The name to use for the commit.                                                                                                                                                                  | No                                | GitHub Action             |
| `COMMIT_USER_EMAIL` | The email to use for the commit.                                                                                                                                                                 | No                                | action@github.com         |
| `SECTION_NAME`      | The section name in the `README.md` to update.                                                                                                                                                   | No                                | readme-stats              |
| `HIDE_REPO_INFO`    | Whether to hide the repository information in action logs.                                                                                                                                       | No                                | -                         |

### Metrics
The `SHOW_METRICS` environment variable is used to specify the metrics to show in the `README.md` file. You can choose from the following metrics:

**COMMIT_TIME_OF_DAY**: The time of day you make commits.

   **🕒 I'm An Afternoon Warrior 🥷🏻**
   ```
   🌅 Morning             492 commits         ████████░░░░░░░░░░░░░░░░░   33.61%
   🌞 Daytime             916 commits         ████████████████░░░░░░░░░   62.57%
   🌆 Evening             52 commits          █░░░░░░░░░░░░░░░░░░░░░░░░   03.55%
   🌙 Night               4 commits           ░░░░░░░░░░░░░░░░░░░░░░░░░   00.27%
   ```

**COMMIT_DAYS_OF_WEEK**: The days of the week you make commits.

   **📅 I'm Most Productive on Sundays**
   ```
   Monday       50 commits     ███░░░░░░░░░░░░░░░░░░░░░░   13.19% 
   Tuesday      85 commits     █████░░░░░░░░░░░░░░░░░░░░   22.43% 
   Wednesday    56 commits     ███░░░░░░░░░░░░░░░░░░░░░░   14.78% 
   Thursday     44 commits     ███░░░░░░░░░░░░░░░░░░░░░░   11.61% 
   Friday       28 commits     █░░░░░░░░░░░░░░░░░░░░░░░░   7.39% 
   Saturday     30 commits     ██░░░░░░░░░░░░░░░░░░░░░░░   7.92% 
   Sunday       86 commits     █████░░░░░░░░░░░░░░░░░░░░   22.69%
   ```

**LANGUAGE_PER_REPO**: The languages you use in each repository.

   **🔥 I Mostly Code in Go**
   ```
   Go                       6 repos             █████████████████████░░░░   85.71%
   TypeScript               1 repo              ████░░░░░░░░░░░░░░░░░░░░░   14.29%
   ```

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
  Windows                  42 hrs 14 mins      █████████████████████████   70.00%
  Mac                      12 hrs 10 mins      ███████░░░░░░░░░░░░░░░░░░   20.00%
  Linux                    6 hrs  3 mins       ████░░░░░░░░░░░░░░░░░░░░░   10.00%
  ```

You can use the `WAKATIME_RANGE` environment variable to set the time range for WakaTime statistics. Each value will show a specific label as follows:
+ `last_7_days`: 📊 This Week I Spent My Time On
+ `last_30_days`: 📊 This Month I Spent My Time On
+ `last_6_months`: 📊 In the Last 6 Months I Spent My Time On
+ `last_year`: 📊 This Year I Spent My Time On
+ `all_time`: 📊 All The Time I Spent On


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
          SHOW_METRICS: "COMMIT_TIME_OF_DAY,LANGUAGE_PER_REPO,COMMIT_DAYS_OF_WEEK,WAKATIME_SPENT_TIME" # show metrics, separated by comma
          SHOW_LAST_UPDATE: "true" # show last update time
          ONLY_MAIN_BRANCH: "true" # only fetch data from the main branch
```

---
#### Inspired From
- [waka-readme-stats](https://github.com/anmol098/waka-readme-stats)
