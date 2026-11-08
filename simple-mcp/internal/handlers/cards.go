package handlers

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DrawInput struct {
	Count *int `json:"count,omitempty" jsonschema:"Number of cards to draw; defaults to 1"`
}

func cardResult(body json.RawMessage, err error) (*mcp.CallToolResult, any, error) {
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		StructuredContent: body,
		Content:           []mcp.Content{&mcp.TextContent{Text: string(body)}},
	}, nil, nil
}

func (h *Handlers) drawCard(ctx context.Context, _ *mcp.CallToolRequest, in DrawInput) (*mcp.CallToolResult, any, error) {
	count := 1
	if in.Count != nil {
		count = *in.Count
	}
	return cardResult(h.deck.Draw(ctx, count))
}

func (h *Handlers) reshuffleDeck(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	return cardResult(h.deck.Reshuffle(ctx))
}
