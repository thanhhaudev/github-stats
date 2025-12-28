package container

import (
	"testing"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/github"
)

func TestCalculateStreaks(t *testing.T) {
	tests := []struct {
		name                 string
		commits              []github.Commit
		expectedCurrent      int
		expectedLongest      int
		description          string
	}{
		{
			name:                 "empty commits",
			commits:              []github.Commit{},
			expectedCurrent:      0,
			expectedLongest:      0,
			description:          "No commits should result in 0 streaks",
		},
		{
			name: "single commit today",
			commits: []github.Commit{
				{CommittedDate: time.Now()},
			},
			expectedCurrent:      1,
			expectedLongest:      1,
			description:          "Single commit today should have streak of 1",
		},
		{
			name: "consecutive days - current streak",
			commits: []github.Commit{
				{CommittedDate: time.Now()},
				{CommittedDate: time.Now().AddDate(0, 0, -1)},
				{CommittedDate: time.Now().AddDate(0, 0, -2)},
				{CommittedDate: time.Now().AddDate(0, 0, -3)},
			},
			expectedCurrent:      4,
			expectedLongest:      4,
			description:          "4 consecutive days should have current and longest streak of 4",
		},
		{
			name: "broken streak",
			commits: []github.Commit{
				{CommittedDate: time.Now().AddDate(0, 0, -10)},
				{CommittedDate: time.Now().AddDate(0, 0, -11)},
				{CommittedDate: time.Now().AddDate(0, 0, -12)},
			},
			expectedCurrent:      0,
			expectedLongest:      3,
			description:          "Old commits should have current streak 0 but longest streak 3",
		},
		{
			name: "multiple commits same day",
			commits: []github.Commit{
				// Use noon to avoid crossing midnight when adding hours
				{CommittedDate: time.Now().Truncate(24 * time.Hour).Add(12 * time.Hour)},
				{CommittedDate: time.Now().Truncate(24 * time.Hour).Add(13 * time.Hour)},
				{CommittedDate: time.Now().Truncate(24 * time.Hour).Add(14 * time.Hour)},
				{CommittedDate: time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour).Add(12 * time.Hour)},
				{CommittedDate: time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour).Add(15 * time.Hour)},
			},
			expectedCurrent:      2,
			expectedLongest:      2,
			description:          "Multiple commits on same day should count as 1 day",
		},
		{
			name: "longest streak in the past",
			commits: []github.Commit{
				{CommittedDate: time.Now()},
				{CommittedDate: time.Now().AddDate(0, 0, -1)},
				{CommittedDate: time.Now().AddDate(0, 0, -5)},
				{CommittedDate: time.Now().AddDate(0, 0, -6)},
				{CommittedDate: time.Now().AddDate(0, 0, -7)},
				{CommittedDate: time.Now().AddDate(0, 0, -8)},
				{CommittedDate: time.Now().AddDate(0, 0, -9)},
			},
			expectedCurrent:      2,
			expectedLongest:      5,
			description:          "Current streak 2, but longest streak 5 in the past",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, longest := calculateStreaks(tt.commits)

			if current != tt.expectedCurrent {
				t.Errorf("%s: Expected current streak %d, got %d", tt.description, tt.expectedCurrent, current)
			}

			if longest != tt.expectedLongest {
				t.Errorf("%s: Expected longest streak %d, got %d", tt.description, tt.expectedLongest, longest)
			}
		})
	}
}

