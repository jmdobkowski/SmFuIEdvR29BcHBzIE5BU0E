package providers

import (
	"context"
	"fmt"
	"log"
	"time"
)

// DummyProvider is an implementation of the Provider interface to be whenever possible during development not to over use API rate limits
type DummyProvider struct {
	reqSem chan int
}

func NewDummyProvider(concurrentRequests int) *DummyProvider {
	return &DummyProvider{
		reqSem: make(chan int, concurrentRequests),
	}
}

func (ap *DummyProvider) GetPictures(ctx context.Context, from, to time.Time) ([]string, error) {
	return ResolveDateRange(ctx, from, to, ap.getPicture)
}

func (dp *DummyProvider) getPicture(ctx context.Context, date time.Time) (string, error) {
	url := fmt.Sprintf("dummy://example.com/?date=%s", date.Format("20060102"))

	dp.reqSem <- 1
	log.Printf("sending request to: %v", url)

	select {
	case <-time.After(1 * time.Second):
		break
	case <-ctx.Done():
		return "", fmt.Errorf("context cancelled")
	}

	log.Printf("completed request to: %v", url)
	<-dp.reqSem

	return url, nil
}
