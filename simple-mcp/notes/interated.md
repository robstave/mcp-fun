Build a Go MCP server in `./simple-mcp` using stdio transport and the official MCP Go SDK. Use only the standard library and the SDK.

### Tools

| Tool | Input | Behavior |
|---|---|---|
| `add` | `numbers`: list of integers | Return their sum; reject an empty list. |
| `greet` | `name`: string | Return a greeting; reject blank or whitespace-only names. |
| `get_weather` | `city`: string | Return weather from the injected provider; reject blank or whitespace-only cities. |

Return clear validation errors.

### Architecture

- MCP handlers only translate inputs and outputs and call services.
- Business logic and validation live in services, with no MCP imports.
- Dependencies are injected through constructors. No globals or `init()` functions.
- `main.go` only constructs dependencies, registers tools, and starts the server.
- Define small interfaces in the package that consumes them.
- The weather service consumes a `WeatherProvider` interface defined in `services`.
- A mock weather adapter returns `"76 degrees"` for every city. A real HTTP adapter must be replaceable without changes to services or handlers.
- Pass `context.Context` through every call that could perform I/O.
- Reserve stdout for MCP protocol traffic. Send all logs to stderr using `log/slog`.

### Package layout

```text
simple-mcp/
├── README.md
├── go.mod
├── go.sum
├── cmd/
│   └── simple-mcp/
│       └── main.go
└── internal/
    ├── handlers/
    │   ├── add.go
    │   ├── greet.go
    │   ├── weather.go
    │   └── handlers_test.go
    ├── services/
    │   ├── add.go
    │   ├── add_test.go
    │   ├── greet.go
    │   ├── greet_test.go
    │   ├── weather.go
    │   └── weather_test.go
    └── adapters/
        └── weather/
            └── mock.go
```

### Tests and documentation

- Table-driven service tests cover successful calls and invalid inputs.
- Weather service tests use a fake provider and cover successful results and provider errors.
- At least one end-to-end tool test uses an in-memory MCP client/server with injected fakes, exercising input decoding, service invocation, and output translation.
- A short README explains the layout, how to build and run the server, and how to register it with an MCP client.

### Acceptance criteria

These commands must pass from `simple-mcp`:

```sh
go build ./...
go vet ./...
go test ./...
```

Implementation begins after your OK.