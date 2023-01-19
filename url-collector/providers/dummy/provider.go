package dummy

import (
	"context"
	"fmt"
	"time"
)

type DummyProvider struct {
}

func (dp *DummyProvider) GetPictures(ctx context.Context, date time.Time) ([]string, error) {
	return []string{
		fmt.Sprintf("example.com/dummy%s.jpg", date.Format("20060102")),
	}, nil
}
