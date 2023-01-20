package internal

import (
	"context"
	"fmt"
	"time"
)

const concurrentRequests = 5

var reqSem chan int = make(chan int, concurrentRequests)

type Provider interface {
	GetPicture(ctx context.Context, date time.Time) (string, error)
}

func Download(ctx context.Context, provider Provider, from, to time.Time) ([]string, error) {
	n := (int(to.Sub(from).Hours()) / 24) + 1

	results := make(chan string, n)
	errors := make(chan error)
	for d := from; d.Before(to) || d.Equal(to); d = d.Add(24 * time.Hour) {
		d := d
		reqSem <- 1
		go func() {
			elem, err := provider.GetPicture(ctx, d)
			if err != nil {
				errors <- fmt.Errorf("cannot fetch image for day (%v): %w", d, err)
			}
			results <- elem
			<-reqSem
		}()
	}

	var res []string
	for i := 0; i < n; i++ {
		select {
		case r := <-results:
			res = append(res, r)
		case err := <-errors:
			return nil, err
		}
	}

	return res, nil
}
