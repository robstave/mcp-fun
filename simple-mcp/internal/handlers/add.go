package handlers

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AddInput struct {
	Numbers []int `json:"numbers" jsonschema:"Integers to sum"`
}

type AddOutput struct {
	Sum int `json:"sum"`
}

func (h *Handlers) add(ctx context.Context, _ *mcp.CallToolRequest, in AddInput) (*mcp.CallToolResult, AddOutput, error) {
	sum, err := h.adder.Add(ctx, in.Numbers)
	return nil, AddOutput{Sum: sum}, err
}
