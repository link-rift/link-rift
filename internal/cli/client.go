package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIResponse is the standard envelope from the Linkrift API.
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *APIError       `json:"error,omitempty"`
	Meta    *APIMeta        `json:"meta,omitempty"`
}

// APIError represents an API error body.
type APIError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// APIMeta represents pagination metadata.
type APIMeta struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// Client is the HTTP client for the Linkrift API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientFromConfig creates a client from the saved CLI config.
func NewClientFromConfig() (*Client, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("not authenticated — run `linkrift login` first")
	}

	return NewClient(cfg.APIBaseURL, cfg.AccessToken), nil
}

// Do performs a raw HTTP request and returns the parsed API response.
func (c *Client) Do(method, path string, body any) (*APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respData, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		msg := "unknown error"
		if apiResp.Error != nil {
			msg = apiResp.Error.Message
		}
		return &apiResp, fmt.Errorf("API error: %s", msg)
	}

	return &apiResp, nil
}

// Get performs a GET request.
func (c *Client) Get(path string) (*APIResponse, error) {
	return c.Do(http.MethodGet, path, nil)
}

// Post performs a POST request.
func (c *Client) Post(path string, body any) (*APIResponse, error) {
	return c.Do(http.MethodPost, path, body)
}

// Put performs a PUT request.
func (c *Client) Put(path string, body any) (*APIResponse, error) {
	return c.Do(http.MethodPut, path, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) (*APIResponse, error) {
	return c.Do(http.MethodDelete, path, nil)
}
