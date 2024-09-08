package wakatime

import "log"

type WakaTime struct {
	Stats *StatsService
}

// NewWakaTime creates a new WakaTime
func NewWakaTime(logger *log.Logger, apiKey string, statsRange StatsRange) *WakaTime {
	if apiKey == "" {
		return nil
	}

	if !statsRange.IsValid() { // if the stats range is invalid, set it to the last 7 days
		statsRange = StatsRangeLast7Days
	}

	client := NewClient(apiKey)

	return &WakaTime{
		Stats: &StatsService{client, logger, statsRange},
	}
}
