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

	err := s.Client.GetWithContext(ctx, fmt.Sprintf("users/current/stats/%s", s.Range), nil, &stats)
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
