package container

import (
	"fmt"
	"time"
)

// CommitStats stores the calculated commit data
type CommitStats struct {
	TotalCommits     int
	YearlyCommits    map[int]int
	DailyCommits     map[time.Weekday]int
	QuarterlyCommits map[string]int
}

// LanguageStats stores the calculated language data
type LanguageStats struct {
	TotalLanguages int
	TotalSize      int
	Languages      map[string][2]interface{}
}

// CalculateCommits calculates the number of commits per year and per day of the week
// return commits per year, commits per day of the week
func (d *DataContainer) CalculateCommits() *CommitStats {
	yearlyCommits := make(map[int]int)
	quarterlyCommits := make(map[string]int, 4)
	dailyCommits := make(map[time.Weekday]int, 7)

	var totalCommits int
	for _, commit := range d.Data.Commits {
		commitDate := commit.CommittedDate
		year := commitDate.Year()
		day := commitDate.Weekday()
		month := commitDate.Month()
		quarter := (int(month)-1)/3 + 1

		yearlyCommits[year]++
		dailyCommits[day]++
		key := fmt.Sprintf("%d-Q%d", year, quarter)
		quarterlyCommits[key]++
		totalCommits++
	}

	return &CommitStats{
		TotalCommits:     totalCommits,
		YearlyCommits:    yearlyCommits,
		DailyCommits:     dailyCommits,
		QuarterlyCommits: quarterlyCommits,
	}
}

// CalculateLanguages calculates the number of languages used in repositories on GitHub
func (d *DataContainer) CalculateLanguages() *LanguageStats {
	totalLanguages := make(map[string][2]interface{}) // [name][2]string{color, size}
	totalSize := 0

	for _, repo := range d.Data.Repositories {
		for _, lang := range repo.Languages.Edges {
			name := lang.Node.Name
			color := lang.Node.Color
			size := lang.Size

			if _, ok := totalLanguages[name]; ok {
				langData := totalLanguages[name]
				langData[1] = langData[1].(int) + size
				totalLanguages[name] = langData
			} else {
				totalLanguages[name] = [2]interface{}{color, size}
			}

			totalSize += size
		}
	}

	return &LanguageStats{
		TotalLanguages: len(totalLanguages),
		TotalSize:      totalSize,
		Languages:      totalLanguages,
	}
}
