package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jmdobkowski/SmFuIEdvR29BcHBzIE5BU0E/url-collector/internal"
)

// APODProvider uses NASA APOD API to return URLs of one space photo per day
type APODProvider struct {
	reqSem     chan int
	httpClient http.Client
	apiKey     string
}

// NewAPODProvider initializes and returns a new instance of APODProvider
func NewAPODProvider(apiKey string, concurrentRequests int) *APODProvider {
	return &APODProvider{
		reqSem:     make(chan int, concurrentRequests),
		apiKey:     apiKey,
		httpClient: http.Client{Timeout: 10 * time.Second},
	}
}

// GetPictures queries APOD for each day between from and to and returns a slice of the returned URLs of the image files
func (ap *APODProvider) GetPictures(ctx context.Context, from, to time.Time) ([]string, error) {
	if from.Before(time.Date(1995, 06, 16, 0, 0, 0, 0, time.UTC)) {
		return nil, internal.BadRequestErrorf("cannot query before 1995-06-16")
	}
	if to.After(time.Now()) {
		return nil, internal.BadRequestErrorf("cannot query for future date %v", to)
	}

	return ResolveDateRange(ctx, from, to, ap.getPicture)
}

type apodResponse struct {
	Url string `json:"url"`
}

func (ap *APODProvider) getPicture(ctx context.Context, date time.Time) (string, error) {

	url := fmt.Sprintf(
		"https://api.nasa.gov/planetary/apod?api_key=%s&date=%s",
		ap.apiKey,
		date.Format("2006-01-02"),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}

	ap.reqSem <- 1
	log.Printf("sending request to: %v", url)
	resp, err := ap.httpClient.Do(req)
	<-ap.reqSem
	if err != nil {
		log.Printf("request to %v failed: %v", url, err)
		return "", fmt.Errorf("request to %v failed: %w", url, err)
	}
	log.Printf("completed request to: %v", url)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2*1024))
		return "", fmt.Errorf("request to %s resulted with %d (%s)", url, resp.StatusCode, string(body))
	}

	decoder := json.NewDecoder(io.LimitReader(resp.Body, 50*1024))
	var r apodResponse
	err = decoder.Decode(&r)
	if err != nil {
		return "", fmt.Errorf("failed to parse response from %v: %w", url, err)
	}
	return r.Url, nil
}
