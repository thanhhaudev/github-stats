package container

import (
	"fmt"
	"sort"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/github"
)

// CommitStats stores the calculated commit data
type CommitStats struct {
	TotalCommits     int
	YearlyCommits    map[int]int
	DailyCommits     map[time.Weekday]int
	QuarterlyCommits map[string]int
	CurrentStreak    int
	LongestStreak    int
}

// LanguageStats stores the calculated language data
type LanguageStats struct {
	TotalLanguages int
	TotalSize      int
	Languages      map[string][2]interface{}
}

// AIStats stores AI vs human attribution from WakaTime's top-level totals.
// AvgPromptLength comes from doc-defined ai_average_prompt_length when WakaTime
// returns it; PromptLength is the raw total chars typed (ai_prompt_length) used
// as fallback since the average field is currently missing in API responses.
// HasData is false when nothing meaningful was reported, signalling writers to
// hide the block entirely (avoids rendering a section full of zeros).
type AIStats struct {
	AIAdditions     int64
	AIDeletions     int64
	HumanAdditions  int64
	HumanDeletions  int64
	AIInputTokens   int64
	AIOutputTokens  int64
	AvgPromptLength float64
	PromptLength    int64
	HasData         bool
}

// CalculateAIStats reads AI attribution from the top-level WakaTime stats.
// We trust the API's own aggregation rather than re-summing per-project items.
func (d *DataContainer) CalculateAIStats() *AIStats {
	if d.Data.WakaTime == nil {
		return &AIStats{}
	}

	src := d.Data.WakaTime.Data
	s := AIStats{
		AIAdditions:     src.AIAdditions,
		AIDeletions:     src.AIDeletions,
		HumanAdditions:  src.HumanAdditions,
		HumanDeletions:  src.HumanDeletions,
		AIInputTokens:   src.AIInputTokens,
		AIOutputTokens:  src.AIOutputTokens,
		AvgPromptLength: src.AIAvgPromptLength,
		PromptLength:    src.AIPromptLength,
	}
	s.HasData = s.AIAdditions > 0 || s.AIInputTokens > 0
	return &s
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

	// Calculate streaks
	currentStreak, longestStreak := calculateStreaks(d.Data.Commits)

	return &CommitStats{
		TotalCommits:     totalCommits,
		YearlyCommits:    yearlyCommits,
		DailyCommits:     dailyCommits,
		QuarterlyCommits: quarterlyCommits,
		CurrentStreak:    currentStreak,
		LongestStreak:    longestStreak,
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

// truncateToMidnight truncates a time to midnight in its local timezone
func truncateToMidnight(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// calculateStreaks calculates the current and longest commit streaks
// A streak is defined as consecutive days with at least one commit
func calculateStreaks(commits []github.Commit) (currentStreak, longestStreak int) {
	if len(commits) == 0 {
		return 0, 0
	}

	// get timezone from the first commit
	loc := commits[0].CommittedDate.Location() // use the timezone from TIME_ZONE env variable

	// create a map of unique commit dates
	uniqueDates := make(map[string]time.Time)
	for _, commit := range commits {
		midnight := truncateToMidnight(commit.CommittedDate)
		dateKey := midnight.Format("2006-01-02")
		if _, exists := uniqueDates[dateKey]; !exists {
			uniqueDates[dateKey] = midnight
		}
	}

	// convert map to sorted slice of dates
	var dates []time.Time
	for _, date := range uniqueDates {
		dates = append(dates, date)
	}

	// sort dates in descending order (most recent first)
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].After(dates[j])
	})

	if len(dates) == 0 {
		return 0, 0
	}

	// calculate current streak from today backwards
	now := time.Now().In(loc)
	today := truncateToMidnight(now)
	yesterday := today.AddDate(0, 0, -1)

	// Check if there's a commit today or yesterday to start the streak
	currentStreak = 0

	// Start from the most recent commit date (already at midnight)
	mostRecentDate := dates[0]
	if mostRecentDate.Equal(today) || mostRecentDate.Equal(yesterday) {
		expectedDate := mostRecentDate
		currentStreak = 1 // Count the first day

		for i := 1; i < len(dates); i++ {
			currentDate := dates[i] // Already at midnight

			// move to next expected date
			nextExpectedDate := expectedDate.AddDate(0, 0, -1)

			// check if this date is the next consecutive day
			if currentDate.Equal(nextExpectedDate) {
				currentStreak++
				expectedDate = nextExpectedDate
			} else {
				break // gap in streak
			}
		}
	}

	// calculate longest streak
	longestStreak = 0
	tempStreak := 1

	for i := 0; i < len(dates)-1; i++ {
		currentDate := dates[i] // Already at midnight
		nextDate := dates[i+1]  // Already at midnight

		// Check if dates are consecutive (1 day apart)
		daysDiff := int(currentDate.Sub(nextDate).Hours() / 24)

		if daysDiff == 1 {
			tempStreak++
		} else {
			if tempStreak > longestStreak {
				longestStreak = tempStreak
			}
			tempStreak = 1
		}
	}

	// check the last streak
	if tempStreak > longestStreak {
		longestStreak = tempStreak
	}

	// if current streak is longer than longest, update longest
	if currentStreak > longestStreak {
		longestStreak = currentStreak
	}

	return currentStreak, longestStreak
}
