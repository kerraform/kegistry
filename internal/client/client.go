package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/kerraform/kegistry/internal/model"
)

type Client struct {
	baseURL   *url.URL
	client    *http.Client
	userAgent string
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

	return c
}

func (c *Client) ServiceDiscovery(ctx context.Context) (*model.Service, error) {
	req, err := c.NewGetRequest(".well-known/terraform.json")
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	svc := &model.Service{}
	if err := json.NewDecoder(resp.Body).Decode(svc); err != nil {
		return nil, err
	}

	return svc, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.client.Do(req.WithContext(ctx))
}

type RequestOpts struct {
	isBinary bool
	url      *url.URL
}

type RequestOpt func(*RequestOpts)

func WithBinary(isBinary bool) RequestOpt {
	return func(ro *RequestOpts) {
		ro.isBinary = isBinary
	}
}

func WithURL(url *url.URL) RequestOpt {
	return func(ro *RequestOpts) {
		ro.url = url
	}
}

// NewGetRequest creates an API GET request.
func (c *Client) NewGetRequest(urlStr string, opts ...RequestOpt) (*http.Request, error) {
	return c.NewRequest("GET", urlStr, nil, opts...)
}

// NewPOSTRequest creates an API POST request.
func (c *Client) NewPostRequest(urlStr string, body interface{}, opts ...RequestOpt) (*http.Request, error) {
	return c.NewRequest("POST", urlStr, body, opts...)
}

// NewPatchRequest creates an API Patch request.
func (c *Client) NewPatchRequest(urlStr string, body interface{}, opts ...RequestOpt) (*http.Request, error) {
	return c.NewRequest("PATCH", urlStr, body, opts...)
}

// NewPutRequest creates an API Patch request.
func (c *Client) NewPutRequest(urlStr string, body interface{}, opts ...RequestOpt) (*http.Request, error) {
	return c.NewRequest("PUT", urlStr, body, opts...)
}

// NewDeleteRequest creates an API Delete request.
func (c *Client) NewDeleteRequest(urlStr string, opts ...RequestOpt) (*http.Request, error) {
	return c.NewRequest("DELETE", urlStr, nil, opts...)
}

// NewRequest creates an API request.
func (c *Client) NewRequest(method, urlStr string, body interface{}, opts ...RequestOpt) (*http.Request, error) {
	var o RequestOpts
	for _, fn := range opts {
		fn(&o)
	}

	u := o.url
	if u == nil {
		var err error
		u, err = c.baseURL.Parse(urlStr)
		if err != nil {
			return nil, err
		}
	}

	var buf io.ReadWriter
	if body != nil {
		if o.isBinary {
			var ok bool
			buf, ok = body.(io.ReadWriter)
			if !ok {
				return nil, errors.New("cannot cast body to buffer")
			}
		} else {
			buf = &bytes.Buffer{}
			enc := json.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			err := enc.Encode(body)
			if err != nil {
				return nil, err
			}
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
