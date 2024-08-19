package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const ApiEndpoint = "https://api.github.com/graphql"

type Client struct {
	token      string
	origin     string
	httpClient *http.Client
}

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

func (c *Client) Post(req *Request, v interface{}) error {
	payload := new(bytes.Buffer)
	if err := json.NewEncoder(payload).Encode(req); err != nil {
		return err
	}

	httpReq, err := c.newRequest(http.MethodPost, c.origin, payload)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// NewClient creates a new GitHub client
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		origin:     ApiEndpoint,
		httpClient: http.DefaultClient,
	}
}
