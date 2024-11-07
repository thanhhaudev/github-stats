package writer

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

const (
	nameLength        = 25
	descriptionLength = 20
	graphLength       = 25
)

type Data struct {
	Name        string
	Description string
	Percent     float64
	Hours       int
	Minutes     int
	Seconds     int
}

const (
	Morning WeekTime = iota
	Daytime
	Evening
	Night
)

type WeekTime int

func (w WeekTime) String() string {
	return longWeekTimeNames[w]
}

// MakeLanguageAndToolList returns a list of languages and tools used in the repositories
func MakeLanguageAndToolList(l map[string][2]interface{}, totalSize int) string {
	if len(l) == 0 {
		return ""
	}

	// Create a map to store the sizes for sorting
	sizeMap := make(map[string]int)
	for key, value := range l {
		sizeMap[key] = value[1].(int)
	}

	res := strings.Builder{}
	for _, k := range sortMapByValue(sizeMap) {
		c := l[k][0].(string)
		s := l[k][1].(int)
		res.WriteString(fmt.Sprintf("![%s](https://img.shields.io/badge/%s-%05.2f%%25-%s?&logo=%s&labelColor=151b23)\n", k, k, float64(s)/float64(totalSize)*100, c[1:], k))
	}

	return "**💬 Languages & Tools**\n\n" + res.String() + "\n\n"
}

// MakeWakaActivityList returns a list of activities
func MakeWakaActivityList(s *wakatime.Stats, i []string) string {
	if s == nil || len(i) == 0 {
		return ""
	}

	var res string
	for _, v := range i {
		switch v {
		case "LANGUAGES":
			res = res + fmt.Sprintf("💬 Languages:") + makeList(buildWakaData(s.Data.Languages)...) + "\n"
		case "EDITORS":
			res = res + fmt.Sprintf("📝 Editors:") + makeList(buildWakaData(s.Data.Editors)...) + "\n"
		case "OPERATING_SYSTEMS":
			res = res + fmt.Sprintf("💻 Operating Systems:") + makeList(buildWakaData(s.Data.OperatingSystems)...) + "\n"
		case "PROJECTS":
			res = res + fmt.Sprintf("📦 Projects:") + makeList(buildWakaData(s.Data.Projects)...) + "\n"
		}
	}

	res = strings.TrimSuffix(res, "\n") // trim last newline

	return fmt.Sprintf("**📊 %s**\n\n", wakaRangeNames[s.Data.Range]) + "```text\n" + res + "```\n\n"
}

func buildWakaData(i []wakatime.StatsItem) []Data {
	var (
		data      []Data
		otherData Data
	)
	for _, d := range i {
		if d.Minutes < 10 && d.Hours == 0 || d.Name == "Other" {
			otherData.Percent += d.Percent
			otherData.Hours += d.Hours
			otherData.Minutes += d.Minutes
			otherData.Seconds += d.Seconds

			continue
		}

		data = append(data, Data{
			Name:        d.Name,
			Description: d.Text,
			Percent:     d.Percent,
			Hours:       d.Hours,
			Minutes:     d.Minutes,
			Seconds:     d.Seconds,
		})
	}

	if otherData.Percent > 0 {
		data = append(data, Data{
			Name:        "Others",
			Description: formatTime(otherData.Hours, otherData.Minutes),
			Percent:     otherData.Percent,
		})
	}

	return data
}

// MakeLastUpdatedOn returns a string with the last updated time
func MakeLastUpdatedOn(t string) string {
	return fmt.Sprintf("\n\n⏳ *Last updated on %s*", t)
}

// MakeCommitTimeOfDayList returns a list of commits made during different times of the day
func MakeCommitTimeOfDayList(commits []github.Commit) string {
	if len(commits) == 0 {
		return ""
	}

	timeRanges := map[WeekTime][2]int{
		Morning: {6, 12},
		Daytime: {12, 18},
		Evening: {18, 23},
		Night:   {23, 6},
	}

	total := len(commits)
	counts := make(map[WeekTime]int)

	for _, commit := range commits {
		hour := commit.CommittedDate.Hour()
		for period, rangeHours := range timeRanges {
			if rangeHours[0] < rangeHours[1] { // if the range is not across midnight
				if hour >= rangeHours[0] && hour < rangeHours[1] {
					counts[period]++
					break
				}
			} else { // if the range is across midnight
				if hour >= rangeHours[0] || hour < rangeHours[1] {
					counts[period]++
					break
				}
			}
		}
	}

	var data []Data
	var topWeek WeekTime
	var topVal int

	for i, n := range longWeekTimeNames {
		weekTime := WeekTime(i)
		weekCommit := counts[WeekTime(i)]
		if weekCommit > topVal {
			topVal = weekCommit
			topWeek = weekTime
		}

		data = append(data, Data{
			Name: fmt.Sprintf("%s %s", weekTimeEmoji[weekTime], n),
			Description: fmt.Sprintf("%s %s", addCommas(weekCommit), func() string {
				if weekCommit > 1 {
					return "commits"
				}

				return "commit"
			}()),
			Percent: float64(weekCommit) / float64(total) * 100,
		})
	}

	return fmt.Sprintf("**🕒 I'm %s**\n\n", weekTimeStatuses[topWeek]) + "```text" + makeList(data...) + "```\n\n"
}

// MakeCommitDaysOfWeekList returns a list of commits made on each day of the week
func MakeCommitDaysOfWeekList(wd map[time.Weekday]int, total int) string {
	if total == 0 {
		return ""
	}

	var (
		topName string
		topVal  int
		data    []Data
	)

	weekdays := []time.Weekday{
		time.Sunday,
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
		time.Saturday,
	}

	for _, weekday := range weekdays {
		if wd[weekday] > topVal {
			topVal = wd[weekday]
			topName = weekday.String()
		}

		data = append(data, Data{
			Name: weekday.String(),
			Description: fmt.Sprintf("%s %s", addCommas(wd[weekday]), func() string {
				if wd[weekday] > 1 {
					return "commits"
				}

				return "commit"
			}()),
			Percent: float64(wd[weekday]) / float64(total) * 100,
		})
	}

	return fmt.Sprintf("**📅 I'm Most Productive on %s**\n\n", topName) + "```text" + makeList(data...) + "```\n\n"
}

// MakeLanguagePerRepoList returns a list of languages and the percentage of repositories that use them
func MakeLanguagePerRepoList(r []github.Repository) string {
	if len(r) == 0 {
		return ""
	}

	var (
		count   float64
		topName string
		topNum  int
		data    []Data
		repos   = make(map[string]int)
	)

	for _, v := range r {
		if v.PrimaryLanguage == nil {
			continue // Skip repositories without a primary language
		}

		repos[v.PrimaryLanguage.Name]++
		count++
	}

	if len(repos) == 0 {
		return ""
	}

	// Create a list of Data structs
	for _, name := range sortMapByValue(repos) {
		num := repos[name]
		if num > topNum { // find top language
			topNum = num
			topName = name
		}

		data = append(data, Data{
			Name: name,
			Description: fmt.Sprintf("%s %s", addCommas(num), func() string {
				if num > 1 {
					return "repos"
				}
				return "repo"
			}()),
			Percent: float64(num) / count * 100,
		})
	}

	return fmt.Sprintf("**🔥 I Mostly Code in %s**\n\n", topName) + "```text" + makeList(data...) + "```\n\n"
}

func makeList(d ...Data) string {
	if len(d) == 0 {
		return "\nNo data available\n"
	}

	var (
		b   strings.Builder
		ver = os.Getenv("PROGRESS_BAR_VERSION")
	)

	for _, val := range d {
		b.WriteString(formatData(val, ver))
	}

	b.WriteString("\n")

	return b.String()
}

func makeProgressBar(p float64) string {
	filledChar := "█"
	emptyChar := "░"

	filledLength := int(math.Round(p / (100 / graphLength)))
	emptyLength := graphLength - filledLength

	return strings.Repeat(filledChar, filledLength) + strings.Repeat(emptyChar, emptyLength)
}

func makeProgressBarV2(p float64) string {
	filledChar := "🟩"
	halfFilledChar := "🟨"
	emptyChar := "⬜"

	percentagePerBlock := float64(100 / graphLength)
	filledLength := int(math.Floor(p / percentagePerBlock))
	remainingPercentage := p - (float64(filledLength) * percentagePerBlock)
	halfFilledLength := 0
	if remainingPercentage > 0 {
		halfFilledLength = 1
	}
	emptyLength := graphLength - filledLength - halfFilledLength

	return strings.Repeat(filledChar, filledLength) + strings.Repeat(halfFilledChar, halfFilledLength) + strings.Repeat(emptyChar, emptyLength)
}

func formatData(data Data, version string) string {
	var b strings.Builder

	n := truncateString(data.Name, nameLength)
	d := truncateString(data.Description, descriptionLength)

	b.WriteString("\n")
	b.WriteString(n)
	b.WriteString(strings.Repeat(" ", nameLength-utf8.RuneCountInString(n)))
	b.WriteString(d)
	b.WriteString(strings.Repeat(" ", descriptionLength-utf8.RuneCountInString(d)))

	if version == "2" {
		b.WriteString(makeProgressBarV2(data.Percent))
	} else {
		b.WriteString(makeProgressBar(data.Percent))
	}

	b.WriteString("   ")
	b.WriteString(formatPercent(data.Percent))

	return b.String()
}

func formatPercent(p float64) string {
	return fmt.Sprintf("%05.2f%%", p)
}

func truncateString(s string, l int) string {
	if utf8.RuneCountInString(s) > l {
		return s[:l]
	}

	return s
}

func addCommas(n int) string {
	str := fmt.Sprintf("%d", n)
	var result []string
	for i, c := range str {
		if (len(str)-i)%3 == 0 && i != 0 {
			result = append(result, ",")
		}
		result = append(result, string(c))
	}

	return strings.Join(result, "")
}

func formatTime(hours, minutes int) string {
	var result string
	if hours > 0 {
		result += fmt.Sprintf("%d %s", hours, func() string {
			if hours > 1 {
				return "hrs"
			}

			return "hr"
		}())
	}

	if minutes > 0 {
		result += fmt.Sprintf(" %d %s", minutes, func() string {
			if minutes > 1 {
				return "mins"
			}

			return "min"
		}())
	}

	return strings.TrimSpace(result)
}

// sortMapByValue sorts a map by its values in descending order
func sortMapByValue(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})

	return keys
}
