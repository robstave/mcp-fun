package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"example.com/simple-mcp/internal/handlers"
	"example.com/simple-mcp/internal/services"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type fakeWeatherProvider struct {
	city  string
	err   error
	calls int
}

func (p *fakeWeatherProvider) Forecast(ctx context.Context, city string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	p.city = city
	p.calls++
	return "rain from fake", p.err
}

func TestToolsOverMCP(t *testing.T) {
	for _, tt := range []struct {
		name, tool        string
		args              map[string]any
		wantJSON, wantErr string
		providerErr       error
		providerCalls     int
	}{
		{name: "add", tool: "add", args: map[string]any{"numbers": []int{2, -1, 5}}, wantJSON: `{"sum":6}`},
		{name: "greet", tool: "greet", args: map[string]any{"name": " Sam "}, wantJSON: `{"greeting":"Hello, Sam!"}`},
		{name: "weather fake", tool: "get_weather", args: map[string]any{"city": " Oslo "}, wantJSON: `{"forecast":"rain from fake"}`, providerCalls: 1},
		{name: "provider error", tool: "get_weather", args: map[string]any{"city": "Oslo"}, providerErr: errors.New("provider offline"), wantErr: "provider offline", providerCalls: 1},
		{name: "empty list", tool: "add", args: map[string]any{"numbers": []int{}}, wantErr: "numbers must not be empty"},
		{name: "empty name", tool: "greet", args: map[string]any{"name": " "}, wantErr: "name must not be empty"},
		{name: "empty city", tool: "get_weather", args: map[string]any{"city": " "}, wantErr: "city must not be empty"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			provider := &fakeWeatherProvider{err: tt.providerErr}
			server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "1.0.0"}, nil)
			handlers.NewHandlers(services.NewCalculator(slog.Default()), services.NewGreeter(slog.Default()), services.NewWeatherService(provider, slog.Default()), nil).Register(server)
			st, ct := mcp.NewInMemoryTransports()
			ss, err := server.Connect(ctx, st, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer ss.Close()
			client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
			cs, err := client.Connect(ctx, ct, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer cs.Close()
			listed, err := cs.ListTools(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(listed.Tools) != 5 {
				t.Fatalf("registered %d tools, want 5", len(listed.Tools))
			}
			result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: tt.tool, Arguments: tt.args})
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantErr != "" {
				if !result.IsError {
					t.Fatalf("expected tool error: %+v", result)
				}
				var content strings.Builder
				for _, item := range result.Content {
					if text, ok := item.(*mcp.TextContent); ok {
						content.WriteString(text.Text)
					}
				}
				if !strings.Contains(content.String(), tt.wantErr) {
					t.Fatalf("error = %q, want %q", content.String(), tt.wantErr)
				}
			} else {
				if result.IsError {
					t.Fatalf("unexpected tool error: %+v", result)
				}
				got, err := json.Marshal(result.StructuredContent)
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != tt.wantJSON {
					t.Fatalf("output = %s, want %s", got, tt.wantJSON)
				}
			}
			if provider.calls != tt.providerCalls {
				t.Fatalf("provider calls = %d, want %d", provider.calls, tt.providerCalls)
			}
			if provider.calls > 0 && provider.city != "Oslo" {
				t.Fatalf("provider city = %q", provider.city)
			}
		})
	}
}
