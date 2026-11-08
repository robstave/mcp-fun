# Server behavior contract

This document describes the current implementation. It is a reference for
maintainers and AI coding agents; MCP clients receive the shorter server
instructions during initialization. This file is not automatically loaded into
a chat and is not currently exposed as an MCP resource.

## Transport and lifecycle

The server uses stdio MCP with the official Go SDK. Stdout contains protocol
messages only. Logs go to stderr and append to `simple-mcp.log` beside the
executable; a file-open failure falls back to stderr. Logs are not rotated.
The executable runs until stdin closes, interruption, or a server error.

Dependencies are injected at startup. There are five tools and one prompt.
No current adapter requires an API key. Only card operations use the network.

## Tool inputs and outputs

| Tool | Input | Behavior |
| --- | --- | --- |
| `add` | Required `numbers`: nonempty list of integers | Returns `sum`; rejects an empty list. Arithmetic uses Go `int` with no overflow detection. |
| `greet` | Required `name`: string | Trims whitespace; returns `greeting` as `Hello, <name>!`; rejects a blank name. |
| `get_weather` | Required `city`: string | Trims whitespace; rejects a blank city; returns `forecast`. The injected mock returns `76 degrees` for every city. |
| `draw_card` | Optional `count`: integer, default 1 | Rejects nonpositive counts; draws from the shared deck and returns complete API JSON. |
| `reshuffle_deck` | No arguments (`{}`) | Shuffles the shared deck, restoring drawn cards, and returns complete API JSON. |

The MCP SDK checks input schemas before invoking services. Services perform
business validation and emit an INFO `tool called` record even for calls failing
service validation. Weather provider errors include the city in the tool error.

## Deck state and failures

- The first valid draw or reshuffle uses API deck ID `new`.
- A response with `success:true` and a nonempty `deck_id` saves that ID for later
  operations. The deck is shared by card tools for the server process lifetime;
  it is not partitioned by chat or player and is lost on server exit.
- API responses are returned as both structured content and JSON text, including
  extra fields and `success:false` responses. Clients must check `success` as
  well as MCP tool errors. An unsuccessful API response leaves the saved ID intact.
- Request failures, HTTP statuses other than 200, invalid JSON, missing/invalid
  `success`, or a successful response without `deck_id` become tool errors and
  leave the saved ID intact.
- A failed or canceled HTTP request may already have changed the remote deck.
  There is no rollback or automatic retry; clients should not assume retrying a
  draw is safe or invent cards after an uncertain result.
- Card HTTP requests carry the caller's context and use a 15-second client timeout.

## Concurrency

A mutex covers each complete draw or reshuffle: reading the current deck ID,
calling the provider, validating the result, and updating the saved ID. This
prevents concurrent first-use calls from creating separate decks and prevents
draws and reshuffles from overlapping within this service.

Calls wait for the mutex before checking cancellation. Waiting is not itself
interruptible by context cancellation; a canceled request reaching the lock will
return without contacting the provider after it acquires the mutex. There is no
guarantee that calls acquire the lock in arrival order.

The mutex protects one operation, not an entire poker round. Other requests can
interleave between a shuffle and subsequent draws. It does not synchronize
separate server processes or external consumers of the remote deck.

## `poker_night` prompt

The prompt takes no arguments and returns a user message for a showdown among
Silas, Clementine, Horace, and Beatrice. Listing or retrieving it has no deck
side effects. A supporting client submits the message to the chat, which:

1. Calls `reshuffle_deck` once and waits for success.
2. Draws five cards for each player, in that order, with four sequential calls.
3. Checks success, a consistent deck ID, exactly five valid cards per response,
   and 20 distinct cards. It stops on errors rather than silently redealing.
4. Ranks hands using standard five-card high poker, no wild cards, ace-low
   straights, rank/kicker tie breaks, and no suit tie breaks.
5. Displays the cards, categories, placings, and explanation of the winner or tie.
6. Ends with a short action and line in the winner's voice, or one for each tied
   winner. Silas is a dry-witted prospector, Clementine a mischievous saloon
   pianist, Horace a meticulous retired railway clerk, and Beatrice a bold,
   sharp-tongued frontier doctor. These personalities affect narration only.

The server supplies the instructions and card tools. The chat performs the
workflow and evaluates hands; these checks and rankings are not enforced by a
server-side poker evaluator. Prompt availability in the UI depends on the client.
