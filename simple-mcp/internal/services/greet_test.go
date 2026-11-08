package services

import (
	"context"
	"log/slog"
	"testing"
)

func TestGreeter_Greet(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr string
	}{
		{"simple", "Sam", "Hello, Sam!", ""},
		{"trims space", "  Sam  ", "Hello, Sam!", ""},
		{"empty", "", "", "name must not be empty"},
		{"blank", " \t\n ", "", "name must not be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGreeter(slog.Default()).Greet(context.Background(), tt.in)
			if (err == nil && tt.wantErr != "") || (err != nil && err.Error() != tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
