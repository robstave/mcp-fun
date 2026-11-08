package handlers

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GreetInput struct {
	Name string `json:"name" jsonschema:"Name of the person to greet"`
}

type GreetOutput struct {
	Greeting string `json:"greeting"`
}

func (h *Handlers) greet(ctx context.Context, _ *mcp.CallToolRequest, in GreetInput) (*mcp.CallToolResult, GreetOutput, error) {
	greeting, err := h.greeter.Greet(ctx, in.Name)
	return nil, GreetOutput{Greeting: greeting}, err
}
