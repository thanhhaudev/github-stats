package wakatime

type WakaTime struct {
	Stats *StatsService
}

// NewWakaTime creates a new WakaTime
func NewWakaTime(apiKey string) *WakaTime {
	if apiKey == "" {
		return nil
	}

	client := NewClient(apiKey)

	return &WakaTime{
		Stats: &StatsService{client},
	}
}
