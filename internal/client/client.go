package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	serverURL string
	token     string
	interval  time.Duration
	httpC     *http.Client
}

func New(serverURL, token string, intervalSec int) *Client {
	return &Client{
		serverURL: serverURL,
		token:     token,
		interval:  time.Duration(intervalSec) * time.Second,
		httpC:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Run(ctx context.Context) {
	c.report(ctx)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("client stopped")
			return
		case <-ticker.C:
			c.report(ctx)
		}
	}
}

func (c *Client) report(ctx context.Context) {
	body, _ := json.Marshal(map[string]string{
		"token": c.token,
	})

	url := fmt.Sprintf("%s/report", c.serverURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		slog.Error("failed to create request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpC.Do(req)
	if err != nil {
		slog.Error("failed to report ip", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		slog.Info("ip reported successfully")
	} else {
		slog.Warn("ip report failed", "status", resp.StatusCode)
	}
}
