package config

import (
	"strings"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with minimal settings",
			config: &Config{
				GitHubToken: "ghp_test123",
				ShowMetrics: []string{"COMMIT_TIMES_OF_DAY"},
			},
			wantErr: false,
		},
		{
			name: "valid config with all settings",
			config: &Config{
				GitHubToken:    "ghp_test123",
				WakaTimeAPIKey: "waka_test123",
				WakaTimeRange:  "last_7_days",
				WakaTimeData:   []string{"EDITORS", "LANGUAGES"},
				ShowMetrics:    []string{"COMMIT_TIMES_OF_DAY", "LANGUAGE_PER_REPO"},
				ProgressBarVersion: "2",
			},
			wantErr: false,
		},
		{
			name: "missing GITHUB_TOKEN",
			config: &Config{
				ShowMetrics: []string{"COMMIT_TIMES_OF_DAY"},
			},
			wantErr: true,
			errMsg:  "GITHUB_TOKEN is required",
		},
		{
			name: "missing SHOW_METRICS",
			config: &Config{
				GitHubToken: "ghp_test123",
			},
			wantErr: true,
			errMsg:  "SHOW_METRICS is required",
		},
		{
			name: "invalid WAKATIME_RANGE",
			config: &Config{
				GitHubToken:    "ghp_test123",
				WakaTimeAPIKey: "waka_test123",
				WakaTimeRange:  "invalid_range",
				ShowMetrics:    []string{"COMMIT_TIMES_OF_DAY"},
			},
			wantErr: true,
			errMsg:  "WAKATIME_RANGE must be one of",
		},
		{
			name: "invalid WAKATIME_DATA",
			config: &Config{
				GitHubToken:    "ghp_test123",
				WakaTimeAPIKey: "waka_test123",
				WakaTimeData:   []string{"INVALID_DATA"},
				ShowMetrics:    []string{"COMMIT_TIMES_OF_DAY"},
			},
			wantErr: true,
			errMsg:  "WAKATIME_DATA contains invalid value",
		},
		{
			name: "invalid SHOW_METRICS value",
			config: &Config{
				GitHubToken: "ghp_test123",
				ShowMetrics: []string{"INVALID_METRIC"},
			},
			wantErr: true,
			errMsg:  "SHOW_METRICS contains invalid value",
		},
		{
			name: "invalid PROGRESS_BAR_VERSION",
			config: &Config{
				GitHubToken:        "ghp_test123",
				ShowMetrics:        []string{"COMMIT_TIMES_OF_DAY"},
				ProgressBarVersion: "3",
			},
			wantErr: true,
			errMsg:  "PROGRESS_BAR_VERSION must be '1' or '2'",
		},
		{
			name: "valid WAKATIME_RANGE - last_30_days",
			config: &Config{
				GitHubToken:    "ghp_test123",
				WakaTimeAPIKey: "waka_test123",
				WakaTimeRange:  "last_30_days",
				ShowMetrics:    []string{"COMMIT_TIMES_OF_DAY"},
			},
			wantErr: false,
		},
		{
			name: "valid WAKATIME_RANGE - all_time",
			config: &Config{
				GitHubToken:    "ghp_test123",
				WakaTimeAPIKey: "waka_test123",
				WakaTimeRange:  "all_time",
				ShowMetrics:    []string{"COMMIT_TIMES_OF_DAY"},
			},
			wantErr: false,
		},
		{
			name: "WakaTime range without API key should not error",
			config: &Config{
				GitHubToken:   "ghp_test123",
				WakaTimeRange: "invalid_range",
				ShowMetrics:   []string{"COMMIT_TIMES_OF_DAY"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Config.Validate() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

