package github

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewClient_SetsTimeout(t *testing.T) {
	c := NewClient("token", false, false)
	if c.httpClient.Timeout == 0 {
		t.Fatal("expected default timeout to be set")
	}
}

func TestClient_do_HidesGraphQLErrorsWhenRequested(t *testing.T) {
	c := NewClient("ghp_secret", true, true)
	c.httpClient = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"errors":[{"message":"failed for ghp_secret and alice/repo"}]}`)),
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = c.do(req, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if strings.Contains(msg, "ghp_secret") || strings.Contains(msg, "alice/repo") {
		t.Fatalf("expected sensitive data to be hidden, got %q", msg)
	}
	if !strings.Contains(msg, "could not fetch data from GitHub") {
		t.Fatalf("expected generic error, got %q", msg)
	}
}

func TestClient_do_ReturnsDetailedErrorsOnlyWhenVisible(t *testing.T) {
	c := NewClient("ghp_secret", true, false)
	c.httpClient = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"errors":[{"message":"failed for ghp_secret and alice/repo"}]}`)),
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = c.do(req, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if strings.Contains(msg, "ghp_secret") {
		t.Fatalf("expected token to be redacted, got %q", msg)
	}
	if !strings.Contains(msg, "github graphql error:") {
		t.Fatalf("expected detailed graphql error, got %q", msg)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
