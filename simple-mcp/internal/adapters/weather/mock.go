// Package weather holds WeatherProvider implementations.
package weather

import "context"

// MockProvider returns the same forecast for every city. Replace it with an
// HTTP-backed provider later; nothing else needs to change.
type MockProvider struct{}

func NewMockProvider() *MockProvider { return &MockProvider{} }

func (MockProvider) Forecast(ctx context.Context, _ string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return "76 degrees", nil
}
