package wakatime

import "fmt"

type StatsService struct {
	*Client
}

type StatsItem struct {
	Name    string  `json:"name"`
	Digital string  `json:"digital"`
	Percent float64 `json:"percent"`
	Hours   float64 `json:"hours"`
	Minutes int     `json:"minutes"`
	Seconds int     `json:"seconds"`
}

type Stats struct {
	Data struct {
		Languages        []StatsItem `json:"languages"`
		Editors          []StatsItem `json:"editors"`
		OperatingSystems []StatsItem `json:"operating_systems"`
	} `json:"data"`
}

type StatsRange string

const (
	StatsRangeLast7Days   StatsRange = "last_7_days"
	StatsRangeLast30Days  StatsRange = "last_30_days"
	StatsRangeLast6Months StatsRange = "last_6_months"
)

// Get retrieves the user's coding activity statistics
func (s *StatsService) Get(r StatsRange) (*Stats, error) {
	var stats Stats

	err := s.Client.get(fmt.Sprintf("users/current/stats/%s", r), nil, &stats)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
