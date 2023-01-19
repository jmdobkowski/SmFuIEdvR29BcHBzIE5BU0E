package internal

import (
	"context"
	"fmt"
	"time"
)

type Provider interface {
	GetPictures(ctx context.Context, date time.Time) ([]string, error)
}

func Download(ctx context.Context, provider Provider, from, to time.Time) ([]string, error) {
	aproxPictures := from.Sub(from) / (24 * time.Hour)
	res := make([]string, 0, aproxPictures)
	for d := from; d.Before(to) || d.Equal(to); d = d.Add(24 * time.Hour) {
		elems, err := provider.GetPictures(ctx, d)
		if err != nil {
			return nil, fmt.Errorf("cannot fetch images for day (%v): %w", d, err)
		}
		res = append(res, elems...)
	}
	return res, nil
}
