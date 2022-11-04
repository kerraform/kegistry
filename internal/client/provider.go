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

	"github.com/kerraform/kegistry/internal/v1/provider"
	"github.com/kerraform/kegistry/internal/v1/request"
)

type ProviderService struct {
	client *Client
	url    *url.URL
}

func NewProviderClient(urlStr string, c *Client) (*ProviderService, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return &ProviderService{
		client: c,
		url:    url,
	}, nil
}

func (s *ProviderService) CreateVersionPlatform(ctx context.Context, namespace, name, version, pos, arch string) (*url.URL, error) {
	b := &provider.CreateProviderPlatformRequest{
		Data: &request.Data[provider.CreateProviderPlatformRequestDataAttributes, provider.DataType]{
			Attributes: &provider.CreateProviderPlatformRequestDataAttributes{
				OS:   pos,
				Arch: arch,
			},
		},
	}

	req, err := s.client.NewPostRequest(fmt.Sprintf("%s%s/%s/versions/%s/platforms", s.url, namespace, name, version), b)
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

	r := &provider.CreateProviderPlatformResponse{}
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return nil, err
	}

	if r.Data.Links.ProviderBinaryUploads == "" {
		return nil, errors.New("invalid response")
	}

	url, err := url.Parse(r.Data.Links.ProviderBinaryUploads)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func (s *ProviderService) SaveGPGKey(ctx context.Context, key io.ReadWriter) error {
	opts := []RequestOpt{
		WithBinary(true),
	}

	req, err := s.client.NewPutRequest("gpg-key", key, opts...)
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

func (s *ProviderService) UploadProviderBinary(ctx context.Context, u *url.URL, b io.ReadWriter) error {
	opts := []RequestOpt{
		WithBinary(true),
	}

	if strings.HasPrefix(u.String(), "http") || strings.HasPrefix(u.String(), "https") {
		opts = append(opts, WithURL(u))
	}

	req, err := s.client.NewPutRequest(u.String(), b, opts...)
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
