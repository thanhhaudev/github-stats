package wakatime

import "testing"

func TestNewClient_SetsTimeout(t *testing.T) {
	c := NewClient("api-key")
	if c.httpClient.Timeout == 0 {
		t.Fatal("expected default timeout to be set")
	}
}
