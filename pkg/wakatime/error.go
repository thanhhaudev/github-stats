package wakatime

import (
	"fmt"
	"net/http"
)

type WakaTimeError struct {
	StatusCode int
	Message    string
}

func (e *WakaTimeError) Error() string {
	return fmt.Sprintf("WakaTime API error: %s (status code: %d)", e.Message, e.StatusCode)
}

func (e *WakaTimeError) IsNotCompleted() bool {
	return e.StatusCode == http.StatusAccepted
}
