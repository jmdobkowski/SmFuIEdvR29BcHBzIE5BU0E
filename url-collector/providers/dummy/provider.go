package dummy

import (
	"context"
	"fmt"
	"time"
)

type DummyProvider struct {
}

func (dp *DummyProvider) GetPicture(ctx context.Context, date time.Time) (string, error) {
	return fmt.Sprintf("example.com/dummy%s.jpg", date.Format("20060102")), nil
}
