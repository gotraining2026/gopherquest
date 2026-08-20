package goproxy

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/campaign"
)

const defaultBaseURL = "https://proxy.golang.org"

// Client retrieves latest module metadata from the Go Module Proxy.
type Client struct {
	HTTP    Doer
	BaseURL string
}

type latestResponse struct {
	Version string `json:"Version"`
	Time    string `json:"Time"`
}

// Latest retrieves the latest version and publication time for a module.
func (c *Client) Latest(ctx context.Context, modulePath string) (campaign.ModuleInfo, error) {
	endpoint, err := latestURL(c.BaseURL, modulePath)
	if err != nil {
		return campaign.ModuleInfo{}, fmt.Errorf("build module provider URL: %w", err)
	}

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return campaign.ModuleInfo{}, fmt.Errorf("build module provider request: %w: %w", campaign.ErrProvider, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return campaign.ModuleInfo{}, fmt.Errorf("send module provider request: %w: %w", campaign.ErrProvider, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return campaign.ModuleInfo{}, fmt.Errorf("module provider returned status %d: %w", resp.StatusCode, campaign.ErrProvider)
	}

	var wire latestResponse
	if err := json.UnmarshalRead(resp.Body, &wire); err != nil {
		return campaign.ModuleInfo{}, fmt.Errorf("decode module provider response: %w: %w", campaign.ErrProvider, err)
	}
	if wire.Version == "" || wire.Time == "" {
		return campaign.ModuleInfo{}, fmt.Errorf("module provider response is incomplete: %w", campaign.ErrProvider)
	}

	publishedAt, err := time.Parse(time.RFC3339Nano, wire.Time)
	if err != nil {
		return campaign.ModuleInfo{}, fmt.Errorf("parse module provider time: %w: %w", campaign.ErrProvider, err)
	}

	return campaign.ModuleInfo{
		LatestVersion: wire.Version,
		PublishedAt:   publishedAt,
	}, nil
}

func latestURL(baseURL, modulePath string) (string, error) {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	base, err := url.Parse(baseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return "", fmt.Errorf("invalid provider base URL: %w", campaign.ErrInvalid)
	}

	segments := strings.Split(modulePath, "/")
	escaped := make([]string, len(segments))
	for i, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return "", fmt.Errorf("invalid module path segment: %w", campaign.ErrInvalid)
		}
		escaped[i] = url.PathEscape(segment)
	}

	base.Path = strings.TrimRight(base.Path, "/") + "/" + strings.Join(escaped, "/") + "/@latest"
	base.RawPath = ""
	return base.String(), nil
}
