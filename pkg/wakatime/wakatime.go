package wakatime

type WakaTime struct {
	Stats *StatsService
}

// NewWakaTime creates a new WakaTime
func NewWakaTime(apiKey string) *WakaTime {
	client := NewClient(apiKey)

	return &WakaTime{
		Stats: &StatsService{client},
	}
}
