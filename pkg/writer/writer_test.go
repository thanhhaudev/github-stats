package writer

import (
	"strings"
	"testing"

	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

func TestMakeCodingStreakList(t *testing.T) {
	tests := []struct {
		name          string
		stats         *wakatime.AllTimeSinceTodayStats
		currentStreak int
		longestStreak int
		expected      []string // strings that should be present in output
		shouldBeEmpty bool
	}{
		{
			name:          "nil stats and no streaks returns empty string",
			stats:         nil,
			currentStreak: 0,
			longestStreak: 0,
			expected:      []string{},
			shouldBeEmpty: true,
		},
		{
			name:          "works without WakaTime but with streaks",
			stats:         nil,
			currentStreak: 7,
			longestStreak: 30,
			expected: []string{
				"📈 Coding Streak",
				"🔥 Current Streak:",
				"🏆 Longest Streak:",
				"7 days",
				"30 days",
			},
			shouldBeEmpty: false,
		},
		{
			name:          "valid stats with real streak data",
			currentStreak: 14,
			longestStreak: 45,
			stats: &wakatime.AllTimeSinceTodayStats{
				Data: struct {
					TotalSeconds      float64 `json:"total_seconds"`
					Text              string  `json:"text"`
					Decimal           string  `json:"decimal"`
					Digital           string  `json:"digital"`
					DailyAverage      int     `json:"daily_average"`
					IsUpToDate        bool    `json:"is_up_to_date"`
					PercentCalculated int     `json:"percent_calculated"`
					Range             struct {
						Start     string `json:"start"`
						StartDate string `json:"start_date"`
						StartText string `json:"start_text"`
						End       string `json:"end"`
						EndDate   string `json:"end_date"`
						EndText   string `json:"end_text"`
						Timezone  string `json:"timezone"`
					} `json:"range"`
					Timeout int `json:"timeout"`
				}{
					TotalSeconds: 4979819.370848,
					Text:         "1,383 hrs 16 mins",
					DailyAverage: 13437, // ~3.7 hours
					Range: struct {
						Start     string `json:"start"`
						StartDate string `json:"start_date"`
						StartText string `json:"start_text"`
						End       string `json:"end"`
						EndDate   string `json:"end_date"`
						EndText   string `json:"end_text"`
						Timezone  string `json:"timezone"`
					}{
						StartDate: "2024-08-01",
						EndDate:   "2025-12-21",
					},
				},
			},
			expected: []string{
				"📈 Coding Streak",
				"🔥 Current Streak:",
				"🏆 Longest Streak:",
				"📊 Daily Average:",
				"💪 Total Coding Time:",
				"🎯 Coding Consistency:",
				"📅 Active Days:",
				"1,383 hrs 16 mins",
				"14 days",
				"45 days",
			},
			shouldBeEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MakeCodingStreakList(tt.stats, tt.currentStreak, tt.longestStreak)

			if tt.shouldBeEmpty {
				if result != "" {
					t.Errorf("Expected empty string, got: %s", result)
				}
				return
			}

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't.\nGot: %s", expected, result)
				}
			}

			// Verify it contains the markdown code block
			if !strings.Contains(result, "```text") {
				t.Error("Expected output to contain markdown code block")
			}
		})
	}
}

