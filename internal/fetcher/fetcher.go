package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Fetcher struct {
	Timeout time.Duration
}

func NewFetcher(timeout time.Duration) *Fetcher {
	return &Fetcher{Timeout: timeout}
}

func (f *Fetcher) FetchAdsTxt(ctx context.Context, domain string) (string, error) {
	url := fmt.Sprintf("https://%s/ads.txt", domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	cl := &http.Client{Timeout: f.Timeout}
	resp, err := cl.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch failed: %s", resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
