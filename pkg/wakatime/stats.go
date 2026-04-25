package wakatime

import (
	"context"
	"errors"
	"fmt"
	"log"
)

type StatsService struct {
	*Client
	Logger *log.Logger
	Range  StatsRange
}

type StatsItem struct {
	Name    string  `json:"name"`
	Digital string  `json:"digital"`
	Percent float64 `json:"percent"`
	Text    string  `json:"text"`
	Hours   int     `json:"hours"`
	Minutes int     `json:"minutes"`
	Seconds int     `json:"seconds"`

	// AI attribution fields (populated on projects when the user has GenAI tooling integrated)
	AIAdditions           int64   `json:"ai_additions"`
	AIDeletions           int64   `json:"ai_deletions"`
	HumanAdditions        int64   `json:"human_additions"`
	HumanDeletions        int64   `json:"human_deletions"`
	AIInputTokens         int64   `json:"ai_input_tokens"`
	AIOutputTokens        int64   `json:"ai_output_tokens"`
	AIAveragePromptLength float64 `json:"ai_average_prompt_length"`
}

type Stats struct {
	Data struct {
		Status           string      `json:"status"`
		Range            string      `json:"range"`
		Languages        []StatsItem `json:"languages"`
		Editors          []StatsItem `json:"editors"`
		Projects         []StatsItem `json:"projects"`
		OperatingSystems []StatsItem `json:"operating_systems"`
	} `json:"data"`
}

type AllTimeSinceTodayStats struct {
	Data struct {
		TotalSeconds      float64 `json:"total_seconds"`
		Text              string  `json:"text"`
		Decimal           string  `json:"decimal"`
		Digital           string  `json:"digital"`
		DailyAverage      float64 `json:"daily_average"`
		IsUpToDate        bool    `json:"is_up_to_date"`
		PercentCalculated float64 `json:"percent_calculated"`
		Range             struct {
			Start     string `json:"start"`
			StartDate string `json:"start_date"`
			StartText string `json:"start_text"`
			End       string `json:"end"`
			EndDate   string `json:"end_date"`
			EndText   string `json:"end_text"`
			Timezone  string `json:"timezone"`
		} `json:"range"`
		Timeout float64 `json:"timeout"`
	} `json:"data"`
}

type StatsRange string

func (s StatsRange) IsValid() bool {
	switch s {
	case StatsRangeLast7Days, StatsRangeLast30Days, StatsRangeLast6Months, StatsLastYear, StatsRangeAllTime:
		return true
	}

	return false
}

const (
	StatsRangeLast7Days   StatsRange = "last_7_days"
	StatsRangeLast30Days  StatsRange = "last_30_days"
	StatsRangeLast6Months StatsRange = "last_6_months"
	StatsLastYear         StatsRange = "last_year"
	StatsRangeAllTime     StatsRange = "all_time"
)

// Get retrieves the user's coding activity statistics
func (s *StatsService) Get(ctx context.Context) (*Stats, error) {
	var stats Stats

	err := s.GetWithContext(ctx, fmt.Sprintf("users/current/stats/%s", s.Range), nil, &stats)
	if err != nil {
		var wakaTimeErr *WakaTimeError
		if errors.As(err, &wakaTimeErr) && wakaTimeErr.IsNotCompleted() {
			s.Logger.Println("WakaTime processing has not completed yet, please retry after a few minutes")

			return &stats, nil
		}

		return nil, err
	}

	return &stats, nil
}

// GetAllTimeSinceToday retrieves the user's all-time coding statistics since today
func (s *StatsService) GetAllTimeSinceToday(ctx context.Context) (*AllTimeSinceTodayStats, error) {
	var stats AllTimeSinceTodayStats

	err := s.GetWithContext(ctx, "users/current/all_time_since_today", nil, &stats)
	if err != nil {
		var wakaTimeErr *WakaTimeError
		if errors.As(err, &wakaTimeErr) && wakaTimeErr.IsNotCompleted() {
			s.Logger.Println("WakaTime all-time stats processing has not completed yet, please retry after a few minutes")

			return &stats, nil
		}

		return nil, err
	}

	return &stats, nil
}
