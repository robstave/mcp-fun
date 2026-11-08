package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

type fakeProvider struct {
	result string
	err    error
	calls  int
	city   string
	ctx    context.Context
}

func (p *fakeProvider) Forecast(ctx context.Context, city string) (string, error) {
	p.calls++
	p.city, p.ctx = city, ctx
	return p.result, p.err
}

func TestWeatherServiceForecast(t *testing.T) {
	providerErr := errors.New("provider unavailable")
	for _, tt := range []struct {
		name, city, want, wantErr string
		providerErr               error
		cancel                    bool
		calls                     int
	}{
		{name: "success", city: "Oslo", want: "snow", calls: 1},
		{name: "trimmed", city: "  Oslo \t", want: "snow", calls: 1},
		{name: "empty", wantErr: "city must not be empty"},
		{name: "blank", city: " \t\n", wantErr: "city must not be empty"},
		{name: "provider error", city: "Oslo", providerErr: providerErr, wantErr: "provider unavailable", calls: 1},
		{name: "canceled", city: "Oslo", cancel: true, wantErr: "context canceled"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancel {
				cancel()
			}
			provider := &fakeProvider{result: "snow", err: tt.providerErr}
			got, err := NewWeatherService(provider, slog.Default()).Forecast(ctx, tt.city)
			if got != tt.want {
				t.Fatalf("forecast = %q, want %q", got, tt.want)
			}
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			if tt.providerErr != nil && !errors.Is(err, tt.providerErr) {
				t.Fatalf("provider error was not preserved: %v", err)
			}
			if provider.calls != tt.calls {
				t.Fatalf("provider calls = %d, want %d", provider.calls, tt.calls)
			}
			if tt.calls > 0 && (provider.city != "Oslo" || provider.ctx != ctx) {
				t.Fatalf("provider received city %q, context %v", provider.city, provider.ctx)
			}
		})
	}
}
