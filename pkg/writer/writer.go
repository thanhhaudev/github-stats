package writer

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/github"
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

// UpdateReadme updates the README.md file with the provided stats
func UpdateReadme(u, n string) error {
	f := "README.md"
	b, err := os.ReadFile("README.md")
	if err != nil {
		return err
	}

	s := fmt.Sprintf("<!--START_SECTION:%s-->", n)
	e := fmt.Sprintf("<!--END_SECTION:%s-->", n)

	si := strings.Index(string(b), s)
	ei := strings.Index(string(b), e)

	if si == -1 || ei == -1 {
		return fmt.Errorf("section tags %s or %s not found in %s", s, e, f)
	}

	u = string(b)[:si+len(s)] + "\n" + u + "\n" + string(b)[ei:]

	return os.WriteFile(f, []byte(u), 0644)
}

func MakeLastUpdatedOn(t string) string {
	return fmt.Sprintf("\n\n>⏳ Last updated on %s", t)
}

// MakeCommitTimeOfDayList returns a list of commits made during different times of the day
func MakeCommitTimeOfDayList(commits []github.Commit) string {
	if len(commits) == 0 {
		return ""
	}

	timeRanges := map[WeekTime][2]int{
		Morning: {6, 12},
		Daytime: {12, 18},
		Evening: {18, 24},
		Night:   {0, 6},
	}

	total := len(commits)
	counts := make(map[WeekTime]int)

	for _, commit := range commits {
		hour := commit.CommittedDate.Hour()
		for period, rangeHours := range timeRanges {
			if hour >= rangeHours[0] && hour < rangeHours[1] {
				counts[period]++
				break
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
	for name, num := range repos {
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
		return ""
	}

	var b strings.Builder
	for _, v := range d {
		b.WriteString(formatData(v))
	}

	b.WriteString("\n")

	return b.String()
}

func makeGraph(p float64) string {
	d, e, q := "█", "░", math.Round(p/(100/graphLength))

	return strings.Repeat(d, int(q)) + strings.Repeat(e, graphLength-int(q))
}

func formatData(v Data) string {
	var b strings.Builder

	n := truncateString(v.Name, nameLength)
	d := truncateString(v.Description, descriptionLength)

	b.WriteString("\n")
	b.WriteString(n)
	b.WriteString(strings.Repeat(" ", nameLength-len(n)))
	b.WriteString(d)
	b.WriteString(strings.Repeat(" ", descriptionLength-len(d)))
	b.WriteString(makeGraph(v.Percent))
	b.WriteString("   ")
	b.WriteString(formatPercent(v.Percent))

	return b.String()
}

func formatPercent(p float64) string {
	return fmt.Sprintf("%05.2f%%", p)
}

func truncateString(s string, l int) string {
	if len(s) > l {
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
