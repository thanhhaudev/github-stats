package wakatime

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"testing"
)

func newTestStatsService(status int, body string) *StatsService {
	client := NewClient("api-key")
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	return &StatsService{
		Client: client,
		Logger: log.New(io.Discard, "", 0),
		Range:  StatsRangeLast7Days,
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestStatsServiceGetReturnsNotReadyWhenAccepted(t *testing.T) {
	service := newTestStatsService(http.StatusAccepted, "")

	_, err := service.Get(context.Background())

	if !errors.Is(err, ErrStatsNotReady) {
		t.Fatalf("expected ErrStatsNotReady, got %v", err)
	}
}

func TestStatsServiceGetReturnsNotReadyWhenStatsAreStale(t *testing.T) {
	service := newTestStatsService(http.StatusOK, `{"data":{"status":"ok","range":"last_7_days","is_up_to_date":false}}`)

	_, err := service.Get(context.Background())

	if !errors.Is(err, ErrStatsNotReady) {
		t.Fatalf("expected ErrStatsNotReady, got %v", err)
	}
}
