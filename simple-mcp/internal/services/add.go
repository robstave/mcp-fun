package services

import (
	"context"
	"errors"
	"log/slog"
)

// Calculator does local arithmetic.
type Calculator struct{ logger *slog.Logger }

func NewCalculator(logger *slog.Logger) *Calculator { return &Calculator{logger: logger} }

// Add returns the sum of nums.
func (c *Calculator) Add(ctx context.Context, nums []int) (int, error) {
	c.logger.InfoContext(ctx, "tool called", "tool", "add", "number_count", len(nums))
	if len(nums) == 0 {
		return 0, errors.New("numbers must not be empty")
	}
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return sum, nil
}
