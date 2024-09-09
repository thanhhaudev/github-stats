package wakatime

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const ApiUrl = "https://wakatime.com/api/v1/"

type Client struct {
	apiKey     string
	origin     string
	httpClient *http.Client
}

// newRequest creates a new http.Request
func (c *Client) newRequest(method, uri string, query url.Values) (*http.Request, error) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(c.apiKey))))

	req.URL.RawQuery = query.Encode()

	return req, nil
}

// do sends an HTTP request and decodes the response
func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// ref: https://wakatime.com/developers
	// 202 - Accepted: The request has been accepted for processing,
	// but the processing has not been completed.
	// The stats resource may return this code.
	if resp.StatusCode == http.StatusAccepted {
		return &WakaTimeError{StatusCode: resp.StatusCode, Message: "Accepted"}
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// get sends a GET request to the WakaTime API
func (c *Client) get(endpoint string, query url.Values, v interface{}) error {
	req, err := c.newRequest(http.MethodGet, c.origin+endpoint, query)
	if err != nil {
		return err
	}

	return c.do(req, v)
}

func (c *Client) GetWithContext(ctx context.Context, endpoint string, query url.Values, v interface{}) error {
	// Check if the context is already canceled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	req, err := c.newRequest(http.MethodGet, c.origin+endpoint, query)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	return c.do(req, v)
}

// NewClient creates a new service
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		origin:     ApiUrl,
		httpClient: http.DefaultClient,
	}
}
