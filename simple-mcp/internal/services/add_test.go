package services

import (
	"context"
	"log/slog"
	"testing"
)

func TestCalculator_Add(t *testing.T) {
	tests := []struct {
		name    string
		in      []int
		want    int
		wantErr string
	}{
		{"single", []int{5}, 5, ""},
		{"multiple", []int{1, 2, 3, 4}, 10, ""},
		{"negatives", []int{-5, 2, 3}, 0, ""},
		{"empty", []int{}, 0, "numbers must not be empty"},
		{"nil", nil, 0, "numbers must not be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCalculator(slog.Default()).Add(context.Background(), tt.in)
			if (err == nil && tt.wantErr != "") || (err != nil && err.Error() != tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
