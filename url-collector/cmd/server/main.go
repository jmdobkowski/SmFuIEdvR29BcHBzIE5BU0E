package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jmdobkowski/SmFuIEdvR29BcHBzIE5BU0E/url-collector/internal"
	"github.com/jmdobkowski/SmFuIEdvR29BcHBzIE5BU0E/url-collector/providers/nasa"
)

func main() {
	http.HandleFunc("/pictures", pictures)
	http.ListenAndServe(":4090", nil)
}

func pictures(w http.ResponseWriter, req *http.Request) {
	fromParam := req.URL.Query().Get("from")
	toParam := req.URL.Query().Get("to")

	from, err := time.Parse("2006-01-02", fromParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cannot parse parameter\"from\"")
		return
	}

	to, err := time.Parse("2006-01-02", toParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cannot parse parameter\"to\"")
		return
	}

	urls, err := internal.Download(req.Context(), &nasa.APODProvider{}, from, to)
	if err != nil {
		log.Printf("could not fetch for (%v,%v): %v", from, to, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeResponse(w, urls)
}

type response struct {
	Urls []string `json:"urls"`
}

func writeResponse(w http.ResponseWriter, urls []string) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(response{urls})
}

type errorResponse struct {
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(errorResponse{message})
}
