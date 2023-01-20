package nasa

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type APODProvider struct {
}

type apodResponse struct {
	Copyright string `json:"copyright"`
	Url       string `json:"url"`
}

func (ap *APODProvider) GetPicture(ctx context.Context, date time.Time) (string, error) {
	url := "https://api.nasa.gov/planetary/apod?api_key=DEMO_KEY&date=" + date.Format("2006-01-02")

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}
	log.Printf("sending request to: %v", url)
	resp, err := http.DefaultClient.Do(req)
	log.Printf("completed request to: %v", url)

	if err != nil {
		return "", fmt.Errorf("request to %v failed: %w", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("request to %v resulted with %v: %v", url, resp.StatusCode, string(body))
	}

	decoder := json.NewDecoder(resp.Body)

	var r apodResponse
	err = decoder.Decode(&r)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to parse response from %v: %w", url, err)
	}
	return r.Url, nil
}
