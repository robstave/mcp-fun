// Package cards implements the Deck of Cards HTTP API.
package cards

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient(client *http.Client, baseURL string) *Client {
	return &Client{http: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (c *Client) Draw(ctx context.Context, deckID string, count int) (json.RawMessage, error) {
	return c.request(ctx, deckID, "draw/?count="+strconv.Itoa(count))
}

func (c *Client) Shuffle(ctx context.Context, deckID string) (json.RawMessage, error) {
	return c.request(ctx, deckID, "shuffle/")
}

func (c *Client) request(ctx context.Context, deckID, action string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/"+url.PathEscape(deckID)+"/"+action, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deck API request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deck API HTTP status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read deck API response: %w", err)
	}
	return json.RawMessage(body), nil
}
