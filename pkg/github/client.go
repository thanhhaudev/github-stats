package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const ApiEndpoint = "https://api.github.com"
const defaultHTTPTimeout = 30 * time.Second

type Client struct {
	token        string
	origin       string
	debug        bool
	hideRepoInfo bool
	httpClient   *http.Client
}

type GraphQLError struct {
	Message string `json:"message"`
	Path    []any  `json:"path"`
}

type Response struct {
	Errors []GraphQLError `json:"errors"`
}

// ErrRateLimited signals that GitHub returned a primary/secondary rate-limit
// response. Callers should stop issuing further REST calls for this run.
var ErrRateLimited = errors.New("github rest rate limit reached")

type Request struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// Var sets a variable in the request
func (r *Request) Var(key string, value interface{}) {
	if r.Variables == nil {
		r.Variables = make(map[string]interface{})
	}

	r.Variables[key] = value
}

// Vars returns the variables in the request
func (r *Request) Vars() map[string]interface{} {
	return r.Variables
}

// NewRequest creates a new request
func NewRequest(query string) *Request {
	return &Request{
		Query: query,
	}
}

func (c *Client) Post(req *Request, path string, v interface{}) error {
	payload := new(bytes.Buffer)
	if err := json.NewEncoder(payload).Encode(req); err != nil {
		return err
	}

	httpReq, err := c.newRequest(http.MethodPost, c.origin+path, payload)
	if err != nil {
		return err
	}

	return c.do(httpReq, v)
}

// PostWithContext makes a POST request with a context
func (c *Client) PostWithContext(ctx context.Context, req *Request, path string, v interface{}) error {
	// Check if the context is already canceled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload := new(bytes.Buffer)
	if err := json.NewEncoder(payload).Encode(req); err != nil {
		return err
	}

	httpReq, err := c.newRequest(http.MethodPost, c.origin+path, payload)
	if err != nil {
		return err
	}

	httpReq = httpReq.WithContext(ctx)

	return c.do(httpReq, v)
}

func (c *Client) newRequest(method, uri string, payload *bytes.Buffer) (*http.Request, error) {
	req, err := http.NewRequest(method, uri, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) do(httpReq *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var body bytes.Buffer
	if _, err := body.ReadFrom(resp.Body); err != nil {
		return err
	}

	var gqlResp Response
	if err := json.Unmarshal(body.Bytes(), &gqlResp); err == nil && len(gqlResp.Errors) > 0 {
		var msgs []string
		for _, e := range gqlResp.Errors {
			msg := e.Message
			if c.token != "" {
				msg = strings.ReplaceAll(msg, c.token, "[output hidden]")
			}
			msgs = append(msgs, msg)
		}

		if !c.debug || c.hideRepoInfo {
			return fmt.Errorf("❌ could not fetch data from GitHub, please check your GitHub token")
		}

		return fmt.Errorf("github graphql error: %s", strings.Join(msgs, "; "))
	}

	return json.Unmarshal(body.Bytes(), v)
}

// GetWithContext performs a REST GET against c.origin+path and decodes the JSON
// body into v. A 429, or a 403 with X-RateLimit-Remaining: 0, is reported as
// ErrRateLimited so callers can degrade gracefully instead of failing the run.
func (c *Client) GetWithContext(ctx context.Context, path string, v interface{}) error {
	_, err := c.getURL(ctx, c.origin+path, v)
	return err
}

// getURL performs a REST GET against an absolute URL, decodes the JSON body into
// v, and returns the response headers (so callers can follow Link pagination).
// The rate-limit detection matches GetWithContext.
func (c *Client) getURL(ctx context.Context, url string, v interface{}) (http.Header, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests ||
		(resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0") {
		return nil, ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return nil, err
	}

	return resp.Header, nil
}

// nextLink returns the rel="next" URL from an RFC 5988 Link header, or "" when
// there is no next page.
func nextLink(header string) string {
	for _, part := range strings.Split(header, ",") {
		segments := strings.Split(part, ";")
		if len(segments) < 2 {
			continue
		}

		url := strings.TrimSpace(segments[0])
		if !strings.HasPrefix(url, "<") || !strings.HasSuffix(url, ">") {
			continue
		}

		for _, attr := range segments[1:] {
			if strings.TrimSpace(attr) == `rel="next"` {
				return url[1 : len(url)-1]
			}
		}
	}

	return ""
}

// NewClient creates a new GitHub client
func NewClient(token string, debug bool, hideRepoInfo bool) *Client {
	return &Client{
		token:        token,
		origin:       ApiEndpoint,
		debug:        debug,
		hideRepoInfo: hideRepoInfo,
		httpClient:   &http.Client{Timeout: defaultHTTPTimeout},
	}
}
