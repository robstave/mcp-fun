package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeDeckProvider struct {
	ids  []string
	body json.RawMessage
	err  error
}

// concurrentDeckProvider detects overlapping operations independently of the
// service's lock. The short delay gives competing callers time to enter.
type concurrentDeckProvider struct {
	active atomic.Int32
	mu     sync.Mutex
	ids    []string
}

func (p *concurrentDeckProvider) operation(id string) (json.RawMessage, error) {
	active := p.active.Add(1)
	defer p.active.Add(-1)
	if active != 1 {
		return nil, errors.New("provider operations overlapped")
	}
	p.mu.Lock()
	p.ids = append(p.ids, id)
	p.mu.Unlock()
	time.Sleep(time.Millisecond)
	return json.RawMessage(`{"success":true,"deck_id":"shared"}`), nil
}

func (p *concurrentDeckProvider) Draw(_ context.Context, id string, _ int) (json.RawMessage, error) {
	return p.operation(id)
}

func (p *concurrentDeckProvider) Shuffle(_ context.Context, id string) (json.RawMessage, error) {
	return p.operation(id)
}

func TestConcurrentDeckOperations(t *testing.T) {
	p := &concurrentDeckProvider{}
	s := NewDeckService(p, slog.Default())
	const calls = 24
	start := make(chan struct{})
	results := make(chan error, calls)
	for i := range calls {
		go func() {
			<-start
			var err error
			if i%2 == 0 {
				_, err = s.Draw(context.Background(), 5)
			} else {
				_, err = s.Reshuffle(context.Background())
			}
			results <- err
		}()
	}
	close(start)
	for range calls {
		if err := <-results; err != nil {
			t.Errorf("concurrent operation: %v", err)
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.ids) != calls {
		t.Fatalf("provider calls = %d, want %d", len(p.ids), calls)
	}
	for i, id := range p.ids {
		want := "shared"
		if i == 0 {
			want = "new"
		}
		if id != want {
			t.Errorf("call %d used deck %q, want %q", i, id, want)
		}
	}
}

func (p *fakeDeckProvider) Draw(ctx context.Context, id string, count int) (json.RawMessage, error) {
	p.ids = append(p.ids, id)
	return p.body, p.err
}
func (p *fakeDeckProvider) Shuffle(ctx context.Context, id string) (json.RawMessage, error) {
	p.ids = append(p.ids, id)
	return p.body, p.err
}

func TestDeckStateAfterFailures(t *testing.T) {
	p := &fakeDeckProvider{err: errors.New("offline")}
	s := NewDeckService(p, slog.Default())
	if _, err := s.Draw(context.Background(), 1); !errors.Is(err, p.err) {
		t.Fatalf("error = %v", err)
	}
	p.err = nil
	p.body = json.RawMessage(`{"success":false,"deck_id":"ignored","error":"failed"}`)
	if _, err := s.Reshuffle(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.body = json.RawMessage(`{"success":true,"deck_id":"saved"}`)
	if _, err := s.Draw(context.Background(), 2); err != nil {
		t.Fatal(err)
	}
	p.body = json.RawMessage(`{"success":true}`)
	if _, err := s.Reshuffle(context.Background()); err == nil {
		t.Fatal("expected missing ID error")
	}
	p.body = json.RawMessage(`{"success":false,"error":"exhausted"}`)
	if _, err := s.Draw(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	want := []string{"new", "new", "new", "saved", "saved"}
	for i, id := range p.ids {
		if id != want[i] {
			t.Fatalf("ids = %v, want %v", p.ids, want)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.Draw(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
	if _, err := s.Reshuffle(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
	if _, err := s.Draw(context.Background(), 0); err == nil {
		t.Fatal("expected invalid count error")
	}
	if len(p.ids) != len(want) {
		t.Fatal("invalid or canceled call reached provider")
	}
}
