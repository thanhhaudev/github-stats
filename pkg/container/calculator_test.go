package container

import (
	"log"
	"math"
	"testing"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/config"
	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
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

func newAIStatsContainer(projects []wakatime.StatsItem) *DataContainer {
	d := NewDataContainer(log.Default(), &ClientManager{}, &config.Config{})
	if projects != nil {
		d.Data.WakaTime = &wakatime.Stats{}
		d.Data.WakaTime.Data.Projects = projects
	}
	return d
}

func TestCalculateAIStats(t *testing.T) {
	tests := []struct {
		name     string
		projects []wakatime.StatsItem
		want     AIStats
	}{
		{
			name:     "no wakatime data",
			projects: nil,
			want:     AIStats{},
		},
		{
			name:     "wakatime present but no AI activity",
			projects: []wakatime.StatsItem{{Name: "proj"}},
			want:     AIStats{},
		},
		{
			name: "aggregates additions and tokens across projects",
			projects: []wakatime.StatsItem{
				{Name: "alpha", AIAdditions: 100, HumanAdditions: 50, AIInputTokens: 1000, AIOutputTokens: 2000, AIAveragePromptLength: 100},
				{Name: "beta", AIAdditions: 200, HumanAdditions: 150, AIInputTokens: 3000, AIOutputTokens: 4000, AIAveragePromptLength: 200},
			},
			want: AIStats{
				AIAdditions:     300,
				HumanAdditions:  200,
				AIInputTokens:   4000,
				AIOutputTokens:  6000,
				AvgPromptLength: (100*1000 + 200*3000) / 4000.0, // 175
				HasData:         true,
			},
		},
		{
			name: "skips projects with zero ai_input_tokens for prompt-length avg",
			projects: []wakatime.StatsItem{
				{Name: "with-ai", AIAdditions: 10, AIInputTokens: 500, AIAveragePromptLength: 80},
				{Name: "no-ai", HumanAdditions: 100, AIAveragePromptLength: 9999}, // bogus, should be ignored
			},
			want: AIStats{
				AIAdditions:     10,
				HumanAdditions:  100,
				AIInputTokens:   500,
				AvgPromptLength: 80,
				HasData:         true,
			},
		},
		{
			name: "AI additions without tokens still triggers HasData",
			projects: []wakatime.StatsItem{
				{Name: "p", AIAdditions: 5},
			},
			want: AIStats{
				AIAdditions: 5,
				HasData:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newAIStatsContainer(tt.projects)
			got := d.CalculateAIStats()

			if got.AIAdditions != tt.want.AIAdditions {
				t.Errorf("AIAdditions: got %d, want %d", got.AIAdditions, tt.want.AIAdditions)
			}
			if got.HumanAdditions != tt.want.HumanAdditions {
				t.Errorf("HumanAdditions: got %d, want %d", got.HumanAdditions, tt.want.HumanAdditions)
			}
			if got.AIInputTokens != tt.want.AIInputTokens {
				t.Errorf("AIInputTokens: got %d, want %d", got.AIInputTokens, tt.want.AIInputTokens)
			}
			if got.AIOutputTokens != tt.want.AIOutputTokens {
				t.Errorf("AIOutputTokens: got %d, want %d", got.AIOutputTokens, tt.want.AIOutputTokens)
			}
			if math.Abs(got.AvgPromptLength-tt.want.AvgPromptLength) > 0.001 {
				t.Errorf("AvgPromptLength: got %v, want %v", got.AvgPromptLength, tt.want.AvgPromptLength)
			}
			if got.HasData != tt.want.HasData {
				t.Errorf("HasData: got %v, want %v", got.HasData, tt.want.HasData)
			}
		})
	}
}

