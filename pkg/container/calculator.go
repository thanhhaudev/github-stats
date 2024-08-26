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
