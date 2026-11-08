package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type DeckProvider interface {
	Draw(context.Context, string, int) (json.RawMessage, error)
	Shuffle(context.Context, string) (json.RawMessage, error)
}

// DeckService is shared by all card tools for the server lifetime.
// Its mutex serializes complete operations, including provider calls.
// A DeckService must not be copied after first use.
type DeckService struct {
	mu       sync.Mutex
	provider DeckProvider
	deckID   string
	logger   *slog.Logger
}

func NewDeckService(provider DeckProvider, logger *slog.Logger) *DeckService {
	return &DeckService{provider: provider, logger: logger}
}

func (s *DeckService) currentDeck() string {
	if s.deckID == "" {
		return "new"
	}
	return s.deckID
}

func (s *DeckService) Draw(ctx context.Context, count int) (json.RawMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.InfoContext(ctx, "tool called", "tool", "draw_card", "count", count, "deck_id", s.currentDeck())
	if count < 1 {
		return nil, errors.New("count must be a positive integer")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	body, err := s.provider.Draw(ctx, s.currentDeck(), count)
	return s.accept(body, err)
}

func (s *DeckService) Reshuffle(ctx context.Context) (json.RawMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.InfoContext(ctx, "tool called", "tool", "reshuffle_deck", "deck_id", s.currentDeck())
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	body, err := s.provider.Shuffle(ctx, s.currentDeck())
	return s.accept(body, err)
}

func (s *DeckService) accept(body json.RawMessage, err error) (json.RawMessage, error) {
	if err != nil {
		return nil, fmt.Errorf("deck operation: %w", err)
	}
	var metadata struct {
		Success *bool  `json:"success"`
		DeckID  string `json:"deck_id"`
	}
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("decode deck API response: %w", err)
	}
	if metadata.Success == nil {
		return nil, errors.New("deck API response is missing success")
	}
	if *metadata.Success {
		if metadata.DeckID == "" {
			return nil, errors.New("deck API response is missing deck_id")
		}
		s.deckID = metadata.DeckID
	}
	// Preserve the entire response, including API failures and unknown fields.
	return body, nil
}
