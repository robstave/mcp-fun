package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"example.com/simple-mcp/internal/adapters/cards"
	"example.com/simple-mcp/internal/handlers"
	"example.com/simple-mcp/internal/services"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCardToolsOverMCP(t *testing.T) {
	for _, first := range []string{"draw_card", "reshuffle_deck"} {
		t.Run(first+" first", func(t *testing.T) {
			var paths []string
			payload := `{"success":true,"deck_id":"saved123","cards":[{"code":"6H","image":"https://example.com/6H.png","images":{"svg":"https://example.com/6H.svg","png":"https://example.com/6H.png"},"value":"6","suit":"HEARTS"}],"remaining":51,"extra":{"future":true}}`
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				paths = append(paths, r.URL.RequestURI())
				fmt.Fprint(w, payload)
			}))
			defer api.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1"}, nil)
			deck := services.NewDeckService(cards.NewClient(api.Client(), api.URL), slog.Default())
			handlers.NewHandlers(services.NewCalculator(slog.Default()), services.NewGreeter(slog.Default()), services.NewWeatherService(&fakeWeatherProvider{}, slog.Default()), deck).Register(server)
			st, ct := mcp.NewInMemoryTransports()
			ss, err := server.Connect(ctx, st, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer ss.Close()
			cs, err := mcp.NewClient(&mcp.Implementation{Name: "client", Version: "1"}, nil).Connect(ctx, ct, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer cs.Close()
			steps := []struct {
				tool    string
				args    map[string]any
				wantErr bool
			}{
				{first, map[string]any{}, false},
				{"draw_card", map[string]any{"count": 2}, false},
				{"reshuffle_deck", map[string]any{}, false},
				{"draw_card", map[string]any{}, false},
				{"draw_card", map[string]any{"count": 0}, true},
				{"draw_card", map[string]any{"count": -1}, true},
				{"draw_card", map[string]any{"count": 1.5}, true},
				{"draw_card", map[string]any{"count": "two"}, true},
			}
			for _, step := range steps {
				result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: step.tool, Arguments: step.args})
				if err != nil {
					t.Fatal(err)
				}
				if result.IsError != step.wantErr {
					t.Fatalf("%s(%v): %+v", step.tool, step.args, result)
				}
				if step.wantErr {
					continue
				}
				var want any
				if err := json.Unmarshal([]byte(payload), &want); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(result.StructuredContent, want) {
					t.Fatalf("full JSON lost: %+v", result.StructuredContent)
				}
				if len(result.Content) != 1 || result.Content[0].(*mcp.TextContent).Text != payload {
					t.Fatalf("text JSON lost: %+v", result.Content)
				}
			}
			initial := "/new/draw/?count=1"
			if first == "reshuffle_deck" {
				initial = "/new/shuffle/"
			}
			want := []string{initial, "/saved123/draw/?count=2", "/saved123/shuffle/", "/saved123/draw/?count=1"}
			if !reflect.DeepEqual(paths, want) {
				t.Fatalf("requests = %v, want %v", paths, want)
			}
		})
	}
}

func TestCardToolHTTPError(t *testing.T) {
	// Transport and response failures are covered below the MCP boundary; the
	// shared SDK error translation is exercised in TestToolsOverMCP.
	for _, tt := range []struct {
		name, body, wantErr string
		status              int
	}{
		{"http error", `unavailable`, "503", 503},
		{"invalid json", `<html>error</html>`, "decode deck API response", 200},
		{"trailing json", `{"success":true,"deck_id":"x"} {}`, "decode deck API response", 200},
		{"missing id", `{"success":true}`, "missing deck_id", 200},
		{"missing success", `{}`, "missing success", 200},
		{"null", `null`, "missing success", 200},
		{"api failure", `{"success":false,"error":"Not enough cards remaining","remaining":0,"cards":[]}`, "", 200},
	} {
		t.Run(tt.name, func(t *testing.T) {
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tt.status); fmt.Fprint(w, tt.body) }))
			defer api.Close()
			service := services.NewDeckService(cards.NewClient(api.Client(), api.URL), slog.Default())
			body, err := service.Draw(context.Background(), 1)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %s", err, tt.wantErr)
				}
			} else if err != nil || string(body) != tt.body {
				t.Fatalf("response = %s, %v", body, err)
			}
		})
	}
}
