// Package handlers translates MCP calls to application service calls.
package handlers

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Adder interface {
	Add(context.Context, []int) (int, error)
}

type Greeter interface {
	Greet(context.Context, string) (string, error)
}

type Forecaster interface {
	Forecast(context.Context, string) (string, error)
}

type Deck interface {
	Draw(context.Context, int) (json.RawMessage, error)
	Reshuffle(context.Context) (json.RawMessage, error)
}

type Handlers struct {
	adder   Adder
	greeter Greeter
	weather Forecaster
	deck    Deck
}

func NewHandlers(adder Adder, greeter Greeter, weather Forecaster, deck Deck) *Handlers {
	return &Handlers{adder: adder, greeter: greeter, weather: weather, deck: deck}
}

func (h *Handlers) Register(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{Name: "poker_night", Description: "Reshuffle, deal five cards each to four old-timey players, and rank their poker hands."}, h.pokerNight)
	mcp.AddTool(server, &mcp.Tool{Name: "add", Description: "Sum a non-empty list of integers."}, h.add)
	mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "Greet someone by name."}, h.greet)
	mcp.AddTool(server, &mcp.Tool{Name: "get_weather", Description: "Get mock weather for a nonblank city. Always returns 76 degrees; this is not a live forecast."}, h.getWeather)
	mcp.AddTool(server, &mcp.Tool{Name: "draw_card", Description: "Draw from the shared deck, removing the drawn cards. Count defaults to 1 and must be a positive integer. Creates a deck on first use. Returns full Deck of Cards API JSON; check success even when no tool error is reported. Draw sequentially to assign hands; do not automatically retry an uncertain draw."}, h.drawCard)
	mcp.AddTool(server, &mcp.Tool{Name: "reshuffle_deck", Description: "Reshuffle the shared deck and restore drawn cards; affects every caller using this server process. Creates a shuffled deck on first use. Returns full API JSON; check success. Do not reshuffle between players' hands in a round."}, h.reshuffleDeck)
}
