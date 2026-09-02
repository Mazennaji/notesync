package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Mazennaji/notesync/core/internal/auth"
)

const (
	apiBase = "https://api.notion.com/v1"
	version = "2022-06-28"
	timeout = 20 * time.Second
)

type Client struct {
	token string
	http  *http.Client
}

func New() (*Client, error) {
	token, err := auth.Get()
	if err != nil {
		return nil, err
	}
	return &Client{token: token, http: &http.Client{Timeout: timeout}}, nil
}

func (c *Client) do(method, path string, body any) ([]byte, error) {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, apiBase+path, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Notion-Version", version)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("notion request failed: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("notion %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}
