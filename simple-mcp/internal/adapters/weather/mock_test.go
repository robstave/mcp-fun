package weather

import (
	"context"
	"testing"
)

func TestMockProvider_AllCities76(t *testing.T) {
	p := NewMockProvider()
	for _, city := range []string{"Anaheim", "Oslo", "Tokyo"} {
		got, err := p.Forecast(context.Background(), city)
		if err != nil || got != "76 degrees" {
			t.Errorf("%s: got (%q, %v)", city, got, err)
		}
	}
}

func TestMockProvider_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewMockProvider().Forecast(ctx, "x"); err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
