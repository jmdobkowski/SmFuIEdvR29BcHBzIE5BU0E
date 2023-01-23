package providers

import (
	"context"
	"fmt"
	"time"
)

type PerDateFunc func(ctx context.Context, date time.Time) (string, error)

type chunk struct {
	index  int
	result string
}

// ResolveDateRange calls perDateFunc concurrently for each date between from and to and returns collected results in chronological order
func ResolveDateRange(ctx context.Context, from, to time.Time, perDateFunc PerDateFunc) ([]string, error) {
	const day = 24 * time.Hour

	n := int(to.Sub(from)/day) + 1
	results := make(chan chunk, n)
	errors := make(chan error, n)
	for i := 0; i < n; i++ {
		d := from.AddDate(0, 0, i)
		i := i
		go func() {
			r, err := perDateFunc(ctx, d)
			if err != nil {
				errors <- fmt.Errorf("cannot get result for day (%v): %w", d, err)
			}
			results <- chunk{index: i, result: r}
		}()
	}

	res := make([]string, n)
	for i := 0; i < n; i++ {
		select {
		case r := <-results:
			res[r.index] = r.result
		case err := <-errors:
			return nil, err
		}
	}
	return res, nil
}
