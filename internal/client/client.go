package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type service struct {
	client *Client
}

type Client struct {
	baseURL   *url.URL
	client    *http.Client
	common    service
	userAgent string

	Module   *ModuleService
	Provider *ProviderService
}

type ClientOpts struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	UserAgent  string
}
type ClientOpt func(o *ClientOpts)

func WithHTTPClient(client *http.Client) ClientOpt {
	return func(o *ClientOpts) {
		o.HTTPClient = client
	}
}

func WithUserAgent(ua string) ClientOpt {
	return func(o *ClientOpts) {
		o.UserAgent = ua
	}
}

func New(baseURL *url.URL, opts ...ClientOpt) *Client {
	var o ClientOpts
	for _, opt := range opts {
		opt(&o)
	}

	if o.HTTPClient == nil {
		o.HTTPClient = http.DefaultClient
	}

	c := &Client{
		baseURL: baseURL,
		client:  o.HTTPClient,
	}

	c.Module = (*ModuleService)(&c.common)
	c.Provider = (*ProviderService)(&c.common)
	return c
}

func (c *Client) ServiceDiscovery(ctx context.Context) {

}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.client.Do(req.WithContext(ctx))
}

// NewGetRequest creates an API GET request.
func (c *Client) NewGetRequest(urlStr string) (*http.Request, error) {
	return c.NewRequest("GET", urlStr, nil)
}

// NewPOSTRequest creates an API POST request.
func (c *Client) NewPostRequest(urlStr string, body interface{}) (*http.Request, error) {
	return c.NewRequest("POST", urlStr, body)
}

// NewPatchRequest creates an API Patch request.
func (c *Client) NewPatchRequest(urlStr string, body interface{}) (*http.Request, error) {
	return c.NewRequest("PATCH", urlStr, body)
}

// NewDeleteRequest creates an API Delete request.
func (c *Client) NewDeleteRequest(urlStr string) (*http.Request, error) {
	return c.NewRequest("DELETE", urlStr, nil)
}

// NewRequest creates an API request.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := c.baseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return req, nil
}
