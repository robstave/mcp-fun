package services

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestServicesLogToolCalls(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	ctx := context.Background()
	deck := NewDeckService(&fakeDeckProvider{body: json.RawMessage(`{"success":true,"deck_id":"saved"}`)}, logger)
	// Invalid inputs are still logged before validation.
	NewCalculator(logger).Add(ctx, nil)
	NewGreeter(logger).Greet(ctx, "ducky")
	NewWeatherService(nil, logger).Forecast(ctx, "")
	deck.Draw(ctx, 3)
	deck.Reshuffle(ctx)

	decoder := json.NewDecoder(&output)
	for _, tool := range []string{"add", "greet", "get_weather", "draw_card", "reshuffle_deck"} {
		var record map[string]any
		if err := decoder.Decode(&record); err != nil {
			t.Fatal(err)
		}
		if record["tool"] != tool || record["msg"] != "tool called" || record["level"] != "INFO" {
			t.Fatalf("unexpected log record: %v", record)
		}
		if tool == "draw_card" && (record["count"] != float64(3) || record["deck_id"] != "new") {
			t.Fatalf("draw attributes: %v", record)
		}
		if tool == "reshuffle_deck" && record["deck_id"] != "saved" {
			t.Fatalf("reshuffle attributes: %v", record)
		}
	}
	if decoder.More() {
		t.Fatal("unexpected extra log records")
	}
}
