package handlers_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"example.com/simple-mcp/internal/handlers"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestPokerNightPromptOverMCP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1"}, &mcp.ServerOptions{Instructions: handlers.Instructions})
	// No services: listing and retrieving prompts must not invoke any tools.
	handlers.NewHandlers(nil, nil, nil, nil).Register(server)
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
	if got := cs.InitializeResult().Instructions; got != handlers.Instructions || got == "" {
		t.Fatalf("initialization instructions = %q, want server usage guide", got)
	}
	listed, err := cs.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Prompts) != 1 || listed.Prompts[0].Name != "poker_night" {
		t.Fatalf("unexpected prompts: %+v", listed.Prompts)
	}
	if len(listed.Prompts[0].Arguments) != 0 {
		t.Fatal("poker_night should require no arguments")
	}
	result, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: "poker_night"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Messages) != 1 || result.Messages[0].Role != "user" {
		t.Fatalf("unexpected prompt messages: %+v", result.Messages)
	}
	content, ok := result.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("prompt content is %T, want text", result.Messages[0].Content)
	}
	for _, instruction := range []string{"Silas", "Clementine", "Horace", "Beatrice", "reshuffle_deck", "draw_card", `{"count":5}`, "sequentially", "success:false", "deck_id", "all 20 cards are distinct", "A-2-3-4-5", "Suits never break ties", "kickers", "tied winners"} {
		if !strings.Contains(content.Text, instruction) {
			t.Errorf("prompt is missing %q", instruction)
		}
	}
	if _, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: "unknown"}); err == nil {
		t.Fatal("unknown prompt should return a protocol error")
	}
}
