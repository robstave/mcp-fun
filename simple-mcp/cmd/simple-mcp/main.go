// Command simple-mcp runs an MCP server over stdio.
// main only wires dependencies and starts the server.
package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"example.com/simple-mcp/internal/adapters/cards"
	"example.com/simple-mcp/internal/adapters/weather"
	"example.com/simple-mcp/internal/handlers"
	"example.com/simple-mcp/internal/services"
)

func main() {
	// stdout is the protocol channel. Mirror stderr logs to a file so they can
	// be followed from a terminal even when an MCP client owns this process.
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	executable, err := os.Executable()
	if err != nil {
		logger.Error("locate executable for log file", "err", err)
		os.Exit(1)
	}
	logPath := filepath.Join(filepath.Dir(executable), "simple-mcp.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		logger.Warn("could not open log file; using stderr", "path", logPath, "err", err)
	} else {
		defer logFile.Close()
		logger = slog.New(slog.NewTextHandler(io.MultiWriter(os.Stderr, logFile), nil))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Wiring: swap weather.NewMockProvider() for a real client later.
	tools := handlers.NewHandlers(
		services.NewCalculator(logger),
		services.NewGreeter(logger),
		services.NewWeatherService(weather.NewMockProvider(), logger),
		services.NewDeckService(cards.NewClient(&http.Client{Timeout: 15 * time.Second}, "https://deckofcardsapi.com/api/deck"), logger),
	)

	server := mcp.NewServer(&mcp.Implementation{Name: "simple-mcp", Version: "v0.1.0"}, &mcp.ServerOptions{
		Logger:       logger,
		Instructions: handlers.Instructions,
	})
	tools.Register(server)

	logger.Info("server starting!")
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
