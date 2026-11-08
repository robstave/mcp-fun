package handlers

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// pokerNight returns instructions only; retrieving it never changes the deck.
func (h *Handlers) pokerNight(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "A five-card poker showdown for Silas, Clementine, Horace, and Beatrice.",
		Messages: []*mcp.PromptMessage{{
			Role: "user",
			Content: &mcp.TextContent{Text: `Host a friendly five-card poker showdown for these four players, in this order:
Silas, Clementine, Horace, and Beatrice.

Keep these personalities brief and consistent:
- Silas: a dusty, soft-spoken prospector with a dry wit. On winning, tips his battered hat: "Reckon that vein paid out."
- Clementine: a sharp-eyed saloon pianist with a mischievous grin. On winning, taps a jaunty rhythm: "That's my kind of grand finale."
- Horace: a fussy retired railway clerk in an immaculate waistcoat. On winning, straightens his cuffs: "Precisely as the figures suggested."
- Beatrice: a fearless frontier doctor with a warm smile and a razor-sharp tongue. On winning, pockets her spectacles: "A healthy result. For me, anyhow."

Use the tools from this MCP server and follow these steps:
1. Call reshuffle_deck with {} once. Wait for a successful response before dealing.
2. Call draw_card with {"count":5} for Silas, then Clementine, then Horace, then Beatrice. Make these four calls sequentially, waiting for each response. Keep the same shared deck throughout; do not reshuffle between players or make unrelated card calls.
3. Use only the cards actually returned by the tools. Check success is true, each draw returns exactly five valid cards, and all 20 cards are distinct. If a tool fails (including success:false), returns invalid or incomplete cards, or changes deck_id, stop and explain the problem. Do not invent missing cards or silently redeal.
4. Grade the four hands using standard five-card high poker, with no jokers, wild cards, discards, or betting. Rank categories from strongest to weakest: straight flush (a royal flush is ace-high), four of a kind, full house, flush, straight, three of a kind, two pair, one pair, high card. Aces are high except in A-2-3-4-5, which is a five-high straight; sequences cannot wrap around.
5. Break ties within a category by comparing: straight/straight flush highest card; four of a kind rank then kicker; full house triple then pair; flush/high card all five ranks descending; three of a kind triple then kickers descending; two pair higher pair, lower pair, then kicker; one pair pair rank then kickers descending. Suits never break ties. Equal rank comparisons are a shared win.
6. Show a table with each player's name, all five cards, hand category, and placing from best to worst. Announce the winner (or tied winners) and briefly explain the decisive ranks or kickers.
7. End with one short in-character action and line from the winner, using their personality above. You may use the example winning line or vary it in the same voice. For a tie, give each tied winner a brief reaction. Keep narration short and friendly; personality never changes the cards, rankings, or tie rules.`},
		}},
	}, nil
}
