Build a Go MCP server (stdio, official go-sdk) in this empty directory.
Plan first: show me the package layout and wait for my OK.

Tools: add (list of ints), greet (name), get_weather (city).

Architecture: thin MCP handlers, business logic in services with no MCP
imports, constructor-based dependency injection, wiring only in main.
Weather goes through a WeatherProvider interface with a mock adapter that
returns "76 degrees" for every city, so a real API can be swapped in later.

Table-driven tests for the services, with a fake provider for weather.
Log to stderr only. Done when build, vet, and tests pass.