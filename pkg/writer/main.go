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
		return fmt.Errorf("section tags not found in README.md")
	}

	u = string(b)[:si+len(s)] + "\n" + u + "\n" + string(b)[ei:]

	return os.WriteFile(f, []byte(u), 0644)
}

func MakeCommitDaysOfWeekList(wd map[time.Weekday]int, total int) string {
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
			Description: fmt.Sprintf("%d %s", wd[weekday], func() string {
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
	var (
		l float64
		t string
		m int
		d []Data
		c = make(map[string]int)
	)

	for _, v := range r {
		if v.PrimaryLanguage == nil {
			continue // Skip repositories without a primary language
		}

		c[v.PrimaryLanguage.Name]++
		l++
	}

	if len(c) == 0 {
		return ""
	}

	// Create a list of Data structs
	for k, v := range c {
		if v > m { // find top language
			m = v
			t = k
		}

		d = append(d, Data{
			Name: k,
			Description: fmt.Sprintf("%d %s", v, func() string {
				if v > 1 {
					return "repos"
				}

				return "repo"
			}()),
			Percent: math.Round(float64(c[k]) / l * 100),
		})
	}

	return fmt.Sprintf("**I Mostly Code in %s**\n\n", t) + "```text" + makeList(d...) + "```\n\n"
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
