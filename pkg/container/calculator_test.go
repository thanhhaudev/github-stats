package container

import (
	"log"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/config"
	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

func TestCalculateStreaks(t *testing.T) {
	tests := []struct {
		name            string
		commits         []github.Commit
		expectedCurrent int
		expectedLongest int
		description     string
	}{
		{
			name:            "empty commits",
			commits:         []github.Commit{},
			expectedCurrent: 0,
			expectedLongest: 0,
			description:     "No commits should result in 0 streaks",
		},
		{
			name: "single commit today",
			commits: []github.Commit{
				{CommittedDate: time.Now()},
			},
			expectedCurrent: 1,
			expectedLongest: 1,
			description:     "Single commit today should have streak of 1",
		},
		{
			name: "consecutive days - current streak",
			commits: []github.Commit{
				{CommittedDate: time.Now()},
				{CommittedDate: time.Now().AddDate(0, 0, -1)},
				{CommittedDate: time.Now().AddDate(0, 0, -2)},
				{CommittedDate: time.Now().AddDate(0, 0, -3)},
			},
			expectedCurrent: 4,
			expectedLongest: 4,
			description:     "4 consecutive days should have current and longest streak of 4",
		},
		{
			name: "broken streak",
			commits: []github.Commit{
				{CommittedDate: time.Now().AddDate(0, 0, -10)},
				{CommittedDate: time.Now().AddDate(0, 0, -11)},
				{CommittedDate: time.Now().AddDate(0, 0, -12)},
			},
			expectedCurrent: 0,
			expectedLongest: 3,
			description:     "Old commits should have current streak 0 but longest streak 3",
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
			expectedCurrent: 2,
			expectedLongest: 2,
			description:     "Multiple commits on same day should count as 1 day",
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
			expectedCurrent: 2,
			expectedLongest: 5,
			description:     "Current streak 2, but longest streak 5 in the past",
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

func newAIStatsContainer(stats *wakatime.Stats) *DataContainer {
	d := NewDataContainer(log.Default(), &ClientManager{}, &config.Config{})
	if stats != nil {
		d.Data.WakaTime = stats
	}
	return d
}

func aiStats(setup func(s *wakatime.Stats)) *wakatime.Stats {
	s := &wakatime.Stats{}
	setup(s)
	return s
}

func TestCalculateAIStats(t *testing.T) {
	tests := []struct {
		name  string
		stats *wakatime.Stats
		want  AIStats
	}{
		{
			name:  "no wakatime data",
			stats: nil,
			want:  AIStats{},
		},
		{
			name:  "wakatime present but no AI activity",
			stats: aiStats(func(s *wakatime.Stats) {}),
			want:  AIStats{},
		},
		{
			name: "reads top-level totals including doc-compliant avg prompt",
			stats: aiStats(func(s *wakatime.Stats) {
				s.Data.AIAdditions = 300
				s.Data.HumanAdditions = 200
				s.Data.AIInputTokens = 4000
				s.Data.AIOutputTokens = 6000
				s.Data.AIAvgPromptLength = 175
			}),
			want: AIStats{
				AIAdditions:     300,
				HumanAdditions:  200,
				AIInputTokens:   4000,
				AIOutputTokens:  6000,
				AvgPromptLength: 175,
				HasData:         true,
			},
		},
		{
			name: "exposes raw ai_prompt_length when avg field is missing",
			stats: aiStats(func(s *wakatime.Stats) {
				s.Data.AIAdditions = 10
				s.Data.AIInputTokens = 500
				s.Data.AIPromptLength = 484803
			}),
			want: AIStats{
				AIAdditions:   10,
				AIInputTokens: 500,
				PromptLength:  484803,
				HasData:       true,
			},
		},
		{
			name: "AI additions without tokens still triggers HasData",
			stats: aiStats(func(s *wakatime.Stats) {
				s.Data.AIAdditions = 5
			}),
			want: AIStats{
				AIAdditions: 5,
				HasData:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newAIStatsContainer(tt.stats)
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
			if got.PromptLength != tt.want.PromptLength {
				t.Errorf("PromptLength: got %d, want %d", got.PromptLength, tt.want.PromptLength)
			}
			if got.HasData != tt.want.HasData {
				t.Errorf("HasData: got %v, want %v", got.HasData, tt.want.HasData)
			}
		})
	}
}

func TestCacheRepoCountSuffix(t *testing.T) {
	tests := []struct {
		name   string
		hidden bool
		count  int
		want   string
	}{
		{
			name:   "hidden repo info omits count",
			hidden: true,
			count:  33,
			want:   "📦 Cache saved",
		},
		{
			name:   "singular count is formatted correctly",
			hidden: false,
			count:  1,
			want:   "📦 Cache saved (1 repo)",
		},
		{
			name:   "plural count is formatted correctly",
			hidden: false,
			count:  33,
			want:   "📦 Cache saved (33 repos)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cacheSavedLogMessage(tt.hidden, tt.count)
			if got != tt.want {
				t.Fatalf("unexpected message: want %q, got %q", tt.want, got)
			}
			if tt.hidden && strings.Contains(got, "repos") {
				t.Fatalf("hidden message leaked repo count: %q", got)
			}
		})
	}
}
