package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
}

type Option func(*Client)

type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.headers[key] = value
	}
}

func WithBasicAuth(username, password string) Option {
	return func(c *Client) {
		c.headers["Authorization"] = "Basic " + basicAuth(username, password)
	}
}

func WithBearerToken(token string) Option {
	return func(c *Client) {
		c.headers["Authorization"] = "Bearer " + token
	}
}

func WithAPIKey(key string, headerName ...string) Option {
	name := "X-API-Key"
	if len(headerName) > 0 {
		name = headerName[0]
	}
	return func(c *Client) {
		c.headers[name] = key
	}
}

func WithIdempotencyKey(key string) Option {
	return func(c *Client) {
		c.headers["Idempotency-Key"] = key
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64Encode([]byte(auth))
}

func base64Encode(data []byte) string {
	const base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	result := make([]byte, 0, (len(data)+2)/3*4)

	for i := 0; i < len(data); i += 3 {
		var n uint32
		remaining := len(data) - i
		switch remaining {
		case 1:
			n = uint32(data[i]) << 16
			result = append(result, base64Table[n>>18], base64Table[(n>>12)&0x3f], '=', '=')
		case 2:
			n = uint32(data[i])<<16 | uint32(data[i+1])<<8
			result = append(result, base64Table[n>>18], base64Table[(n>>12)&0x3f], base64Table[(n>>6)&0x3f], '=')
		default:
			n = uint32(data[i])<<16 | uint32(data[i+1])<<8 | uint32(data[i+2])
			result = append(result, base64Table[n>>18], base64Table[(n>>12)&0x3f], base64Table[(n>>6)&0x3f], base64Table[n&0x3f])
		}
	}
	return string(result)
}

func (c *Client) do(method, path string, body interface{}) (*Response, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

func (c *Client) Get(path string) (*Response, error) {
	return c.do("GET", path, nil)
}

func (c *Client) Post(path string, body interface{}) (*Response, error) {
	return c.do("POST", path, body)
}

func (c *Client) Put(path string, body interface{}) (*Response, error) {
	return c.do("PUT", path, body)
}

func (c *Client) Patch(path string, body interface{}) (*Response, error) {
	return c.do("PATCH", path, body)
}

func (c *Client) Delete(path string) (*Response, error) {
	return c.do("DELETE", path, nil)
}

func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

func (r *Response) String() string {
	return string(r.Body)
}

func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

func (r *Response) IsError() bool {
	return r.StatusCode >= 400
}
