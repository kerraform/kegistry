package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kerraform/kegistry/internal/v1/module"
)

type ModuleService struct {
	client *Client
	url    *url.URL
}

func NewModuleClient(urlStr string, c *Client) (*ModuleService, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return &ModuleService{
		client: c,
		url:    url,
	}, nil
}

func (s *ModuleService) CreateVersion(ctx context.Context, namespace, name, provider, version string) (*url.URL, error) {
	b := &module.CreateModuleVersionRequest{
		Data: &module.CreateModuleVersionRequestData{
			Attributes: &module.CreateModuleVersionDataAttributes{
				Version: version,
			},
		},
	}

	req, err := s.client.NewPostRequest(fmt.Sprintf("%s/%s/%s/%s/versions", s.url, namespace, name, provider), b)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code, got: %d", resp.StatusCode)
	}

	r := &module.CreateModuleVersionResponse{}
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return nil, err
	}

	if r.Data.Links.Upload == "" {
		return nil, errors.New("invalid response")
	}

	url, err := url.Parse(r.Data.Links.Upload)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func (s *ModuleService) UploadModuleVersion(ctx context.Context, u *url.URL, pkg io.ReadWriter) error {
	opts := []RequestOpt{
		WithBinary(true),
	}

	if strings.HasPrefix(u.String(), "http") || strings.HasPrefix(u.String(), "https") {
		opts = append(opts, WithURL(u))
	}

	req, err := s.client.NewPutRequest(u.String(), nil, opts...)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code, got: %d", resp.StatusCode)
	}

	return nil
}
