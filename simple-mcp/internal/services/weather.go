package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

// WeatherProvider supplies forecasts. Implementations must honor ctx.
type WeatherProvider interface {
	Forecast(ctx context.Context, city string) (string, error)
}

type WeatherService struct {
	provider WeatherProvider
	logger   *slog.Logger
}

func NewWeatherService(provider WeatherProvider, logger *slog.Logger) *WeatherService {
	return &WeatherService{provider: provider, logger: logger}
}

func (s *WeatherService) Forecast(ctx context.Context, city string) (string, error) {
	s.logger.InfoContext(ctx, "tool called", "tool", "get_weather", "city", city)
	city = strings.TrimSpace(city)
	if city == "" {
		return "", errors.New("city must not be empty")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	forecast, err := s.provider.Forecast(ctx, city)
	if err != nil {
		return "", fmt.Errorf("get weather for %q: %w", city, err)
	}
	return forecast, nil
}
