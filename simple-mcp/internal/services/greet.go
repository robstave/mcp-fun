package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
)

// Greeter produces greetings.
type Greeter struct{ logger *slog.Logger }

func NewGreeter(logger *slog.Logger) *Greeter { return &Greeter{logger: logger} }

// Greet returns "Hello, <name>!".
func (g *Greeter) Greet(ctx context.Context, name string) (string, error) {
	g.logger.InfoContext(ctx, "tool called", "tool", "greet", "name", name)
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("name must not be empty")
	}
	return "Hello, " + name + "!", nil
}
