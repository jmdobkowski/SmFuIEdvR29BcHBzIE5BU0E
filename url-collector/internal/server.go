package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"
)

const UrlDateFormat = "2006-01-02"

type PictureProvider interface {
	GetPictures(ctx context.Context, from, to time.Time) ([]string, error)
}

type Server struct {
	Provider PictureProvider
}

func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" && req.URL.Path == "/pictures" {
		s.handlePictures(w, req)
		return
	}

	writeError(w, ErrNotFound)
}

func (s Server) handlePictures(w http.ResponseWriter, req *http.Request) {
	from, err := parseDateFromQuery(req.URL.Query(), "from")
	if err != nil {
		writeError(w, err)
		return
	}
	to, err := parseDateFromQuery(req.URL.Query(), "to")
	if err != nil {
		writeError(w, err)
		return
	}
	if to.Before(from) {
		writeError(w, BadRequestErrorf("invalid date range (%v,%v)", from, to))
		return
	}

	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()
	urls, err := s.Provider.GetPictures(ctx, from, to)
	if err != nil {
		writeError(w, err)
		cancel()
		return
	}
	writeResponse(w, urls)
}

func parseDateFromQuery(query url.Values, param string) (time.Time, error) {
	s := query.Get(param)
	if s == "" {
		return time.Time{}, BadRequestErrorf("missing query parameter '%s'", param)
	}
	d, err := time.Parse(UrlDateFormat, s)
	if err != nil {
		return time.Time{}, BadRequestErrorf("cannot parse parameter '%s'", param)
	}
	return d, nil
}

type response struct {
	Urls []string `json:"urls"`
}

func writeResponse(w http.ResponseWriter, urls []string) {
	w.Header().Add("content-type", "application/json")

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(response{urls})
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Add("content-type", "application/json")

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	var requestError RequestError
	switch {
	case errors.As(err, &requestError):
		w.WriteHeader(requestError.Status)
		encoder.Encode(errorResponse{requestError.Error()})
	default:
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("internal server error: %v", err)
		encoder.Encode(errorResponse{"internal server error"})
	}
}
