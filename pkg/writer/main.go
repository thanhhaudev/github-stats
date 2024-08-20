package writer

import (
	"fmt"
	"github.com/thanhhaudev/github-stats/pkg/github"
	"math"
	"strings"
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

	// find top language
	for k, v := range c {
		if v > m {
			m = v
			t = k
		}
	}

	// Create a list of Data structs
	for k, v := range c {
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
	b.WriteString("\n")

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
