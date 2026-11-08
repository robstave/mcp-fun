package handlers

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type WeatherInput struct {
	City string `json:"city" jsonschema:"City to get weather for"`
}

type WeatherOutput struct {
	Forecast string `json:"forecast"`
}

func (h *Handlers) getWeather(ctx context.Context, _ *mcp.CallToolRequest, in WeatherInput) (*mcp.CallToolResult, WeatherOutput, error) {
	forecast, err := h.weather.Forecast(ctx, in.City)
	return nil, WeatherOutput{Forecast: forecast}, err
}
