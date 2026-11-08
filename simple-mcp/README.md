# simple-mcp

A stdio Model Context Protocol (MCP) server using the official Go SDK. It exposes
five tools for arithmetic, greetings, mock weather, and a shared deck of cards,
plus a reusable poker-night prompt.

## Requirements

- Go 1.25 or newer (Go can automatically download the required toolchain).
- Internet access for the card tools, which call the Deck of Cards API.
- Node.js 22.19.0 or newer with npm/npx only if using MCP Inspector.

The weather tool always returns `76 degrees`; it does not fetch live weather.
No API keys are required by the current adapters.

## Layout

- `cmd/simple-mcp/main.go`: constructor wiring, stderr slog logger, and server startup.
- `internal/handlers`: thin MCP input/output adapters, tool registration, and prompt templates.
- `internal/services`: validation and business logic; no MCP imports. Defines the consumed `WeatherProvider` and `DeckProvider` interfaces.
- `internal/adapters/weather`: mock provider returning `76 degrees` for every city.
- `internal/adapters/cards`: HTTP client for the Deck of Cards API.
- `notes/first.md`: original request notes.
- [AGENTS.md](AGENTS.md): repository guidance for AI coding agents.
- [docs/behavior.md](docs/behavior.md): current behavior contract, including errors,
  deck state, concurrency, and the poker workflow.

The server sends a brief usage guide to MCP clients during initialization through
`ServerOptions.Instructions`. Tool descriptions also explain mock weather and
card side effects. Clients decide how to present this guidance to the model;
the Markdown documentation files are repository references, not MCP resources.

Dependencies are supplied through constructors, with no globals or `init()`.
To use a real weather API, implement `Forecast(context.Context, string) (string, error)`
on a new adapter and replace the provider construction in `main.go`.
Pass the context to any HTTP requests in that adapter.

## Build and run

From the workspace root, enter the Go module directory first:

```sh
cd simple-mcp
```

All remaining commands assume you are in `simple-mcp`. Build and check the server:

```sh
go build ./...
go vet ./...
go test ./...
```

Build and run the Windows executable:

```powershell
go build -o simple-mcp.exe ./cmd/simple-mcp
.\simple-mcp.exe
```

On macOS/Linux:

```sh
go build -o simple-mcp ./cmd/simple-mcp
./simple-mcp
```

The executable waits for MCP JSON-RPC traffic on stdin and sends protocol responses
on stdout. Logging uses `log/slog` on stderr and appends to `simple-mcp.log`
beside the executable. If the log file cannot be opened, logging continues on
stderr with a warning.
Stop with Ctrl+C or close stdin.

The startup logger is injected into each service. Every service invocation emits
an INFO `tool called` record with the tool name and relevant arguments (name,
city, number count, or card count and deck ID), including calls that fail service
validation. For example: `level=INFO msg="tool called" tool=draw_card count=3 deck_id=new`.
Requests rejected by MCP input-schema validation do not reach the service.

When an MCP client launches the server, its stderr is separate from your terminal.
After restarting the client connection to load the rebuilt server, follow the
log file in PowerShell (from this directory):

```powershell
Get-Content .\simple-mcp.log -Tail 20 -Wait
```

The file is created when the server starts. Logs append across restarts and are
not rotated automatically.

## Verify with MCP Inspector

[MCP Inspector](https://github.com/modelcontextprotocol/inspector) is the official
tool for interactively testing an MCP server. Requires Node.js 22.19.0 or newer
with npm/npx. From this directory, build the server and launch the Inspector:

```powershell
go build -o simple-mcp.exe ./cmd/simple-mcp
npx @modelcontextprotocol/inspector .\simple-mcp.exe
```

The Inspector starts the server and prints a browser URL. Open that URL, connect
using stdio if prompted, and list the tools. Call `add`, `greet`, `get_weather`, `draw_card`, and `reshuffle_deck`
with the example arguments below; also try an empty list or blank name/city to
check error responses. The Inspector owns the server process, so no separate
server launch is needed.

For a command-line check that lists the registered tools:

```powershell
npx @modelcontextprotocol/inspector --cli .\simple-mcp.exe --method tools/list
```

On macOS/Linux, build with `-o simple-mcp` and use `./simple-mcp` in the commands.

## Register with an MCP client

Build the executable, then add a stdio server to your client's MCP configuration.
Clients using the common `mcpServers` format can use:

```json
{
  "mcpServers": {
    "simple-mcp": {
      "command": "C:\\absolute\\path\\to\\simple-mcp\\simple-mcp.exe",
      "args": []
    }
  }
}
```

Replace `command` with the absolute path to your executable. If your client uses
a settings UI, choose stdio and enter the same executable path with no arguments.

### Codex

Add the following to `%USERPROFILE%\.codex\config.toml` on Windows
(`~/.codex/config.toml` on macOS/Linux), replacing the executable path:

```toml
[mcp_servers.my_server]
args = []
command = 'C:\absolute\path\to\simple-mcp\simple-mcp.exe'
startup_timeout_sec = 120
```

TOML single-quoted strings preserve Windows backslashes literally; JSON paths
require doubled backslashes as shown above. Restart the client connection after
changing the configuration or rebuilding the executable. See the
[official OpenAI MCP configuration documentation](https://learn.chatgpt.com/docs/extend/mcp).

## Prompts

MCP prompts are reusable instructions that a supporting client can discover with
`prompts/list` and retrieve with `prompts/get`. Select a prompt in your client's
prompt picker, then submit its returned message to the chat. Client support and
UI vary; prompts are separate from the tool list. See the
[MCP prompt specification](https://modelcontextprotocol.io/specification/2025-11-25/server/prompts).

### `poker_night`

No arguments. Hosts a five-card poker showdown for **Silas**, **Clementine**,
**Horace**, and **Beatrice**. It instructs the chat to:

1. Call `reshuffle_deck` once.
2. Call `draw_card` with `{"count":5}` four times sequentially, once per player.
3. Check the responses and rank the hands using standard five-card poker rules,
   including ace-low straights, kickers, and shared wins.
4. Show each player's cards, hand category, and placing, then explain the winner
   and finish with a brief reaction in the winner's own voice (each winner for ties).

The cast: **Silas**, a dry-witted prospector; **Clementine**, a mischievous saloon
pianist; **Horace**, a meticulous retired railway clerk; and **Beatrice**, a bold
frontier doctor with a sharp tongue. The prompt includes a short winning line
and signature gesture for each.

Retrieving the prompt does not call tools or change the deck. The chat executes
the workflow and judges the hands; the server does not contain a poker evaluator.
It needs access to this server's card tools to complete the deal. Rebuild the
executable and restart the client connection to discover the new prompt.

You can inspect the prompt over the protocol with MCP Inspector:

```powershell
npx @modelcontextprotocol/inspector --cli .\simple-mcp.exe --method prompts/list
npx @modelcontextprotocol/inspector --cli .\simple-mcp.exe --method prompts/get --prompt-name poker_night
```

## Tools

| Tool | Example arguments | Structured result |
| --- | --- | --- |
| `add` | `{"numbers":[1,2,3]}` | `{"sum":6}` |
| `greet` | `{"name":"Sam"}` | `{"greeting":"Hello, Sam!"}` |
| `get_weather` | `{"city":"Oslo"}` | `{"forecast":"76 degrees"}` |
| `draw_card` | `{}` or `{"count":2}` | Full API JSON: `success`, `deck_id`, `cards` (including images), and `remaining` |
| `reshuffle_deck` | `{}` | Full API JSON: `success`, `deck_id`, `shuffled`, and `remaining` |

The card tools share one `DeckService` instance constructed at startup. On the
first draw or reshuffle, the service calls the API with `new` and saves the
returned deck ID. Later calls reuse that ID. Reshuffling restores drawn cards
to the deck; drawing from an exhausted deck returns the API's failure response.
An omitted card count defaults to 1; explicit counts must be positive integers.
Both tools return the entire API response as structured content and JSON text,
including extra fields and API responses with `success: false`.

Deck state lives in memory until the server exits and is shared by all clients
of that server process. A mutex serializes each complete draw or reshuffle,
including the API request and saved deck ID update. Concurrent first-use calls
therefore reuse one deck, and a reshuffle cannot overlap a draw. Waiting calls
check cancellation after acquiring the mutex; API requests honor cancellation
and have a 15-second HTTP timeout.

The lock covers one tool call at a time. Another caller can still act between
the shuffle and draws in `poker_night`; the entire round is not atomic.
Transport, HTTP, and malformed-response failures become tool errors without
replacing the saved deck ID. See the [API documentation](https://deckofcardsapi.com/).

An empty list or blank name/city produces a clear tool error. Names and cities
are trimmed. Provider failures are returned as tool errors with city context.
Only the standard library and MCP SDK are imported directly; `go.mod` also lists
the SDK's transitive dependencies.

Service tests are table-driven; weather tests use a fake provider, including
failure and cancellation cases. In-memory MCP client/server tests exercise all
registered tools and error translation with an injected fake provider. Card tests
use a local mock HTTP server to verify first-use creation, shared deck reuse,
reshuffling, counts, full JSON preservation, and failure handling without calling
the live API.
