# Working on simple-mcp

This directory is the Go module. Read `README.md` for setup and
`docs/behavior.md` for the current behavior contract before changing features.

## Architecture

- Keep `cmd/simple-mcp/main.go` focused on dependency wiring and server startup.
- Keep MCP schemas, tool/prompt registration, and protocol translation in
  `internal/handlers`. Client-facing usage guidance lives in
  `internal/handlers/instructions.go`.
- Keep validation and business logic in `internal/services`, with no MCP imports.
- Put external API implementations in `internal/adapters`. Define small provider
  interfaces in the consuming service package and inject dependencies through
  constructors. Avoid mutable package globals and `init()` functions.
- Pass `context.Context` through I/O and keep HTTP requests bounded by a timeout.

## Protocol and state

- Reserve stdout for MCP protocol traffic. Log with the injected `slog.Logger`
  to stderr and the existing executable-adjacent log file.
- Weather is mock data. Do not describe it as a live forecast.
- Draw and reshuffle share one in-memory deck per server process. Hold the deck
  mutex across reading the ID, the provider call, and response acceptance.
  Never copy `DeckService` after use or bypass that lock in new operations.
- Preserve complete card API JSON, including unknown fields. API `success:false`
  is returned as content; transport/HTTP/response-validation failures are tool
  errors. Failed operations must not replace the saved deck ID.
- Individual card calls are serialized; multi-call workflows are not atomic.
  Do not promise isolated games or uninterrupted poker rounds.
- Retrieving `poker_night` returns instructions without calling tools. Poker
  grading currently belongs to the chat; there is no server-side evaluator.

## Verification and documentation

Run from this directory after code changes:

```sh
gofmt -w <changed-go-files>
go build ./...
go vet ./...
go test ./...
```

Use `go test -race ./...` for concurrency changes when CGO and a C compiler are
available. Card tests use local HTTP mock servers and need loopback access;
tests must not depend on the live Deck of Cards API. Use in-memory MCP sessions
to verify protocol behavior, and service tests for validation and state changes.

Keep the README, behavior contract, server instructions, tool descriptions, and
prompt aligned when behavior changes. Rebuild the executable and restart the
client connection to load runtime changes. Do not terminate a user's running
server solely to replace a locked executable; report the required restart.
