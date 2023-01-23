package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"testing"
	"time"
)

func TestServeHTTP(t *testing.T) {
	testCases := []struct {
		name   string
		method string
		url    string

		providerCall func(t *testing.T, from, to time.Time) ([]string, error)

		expectedResult        int
		expectedBodyRegExp    string
		expectedBodyNotRegExp string
	}{
		{name: "for invalid path return 404", url: "/testinvalidpath", expectedResult: 404},
		{name: "for invalid method return 404", url: "/pictures", method: "POST", expectedResult: 404},
		{name: "for missing arguments return 400", url: "/pictures", expectedResult: 400, expectedBodyRegExp: "missing"},
		{name: "for missing to return 400", url: "/pictures?from=2022-01-01", expectedResult: 400, expectedBodyRegExp: "missing.*to"},
		{name: "for missing from return 400", url: "/pictures?to=2022-01-01", expectedResult: 400, expectedBodyRegExp: "missing.*from"},
		{name: "for empty to return 400", url: "/pictures?from=2022-01-01&to=", expectedResult: 400, expectedBodyRegExp: "missing.*to"},
		{name: "for empty from return 400", url: "/pictures?from=&to=2022-01-01", expectedResult: 400, expectedBodyRegExp: "missing.*from"},
		{name: "for malformed to return 400", url: "/pictures?from=2022-01-01&to=2022-01-", expectedResult: 400, expectedBodyRegExp: "parse.*to"},
		{name: "for malformed from return 400", url: "/pictures?from=2022-01-&to=2022-01-05", expectedResult: 400, expectedBodyRegExp: "parse.*from"},
		{name: "for invalid to return 400", url: "/pictures?from=2022-01-01&to=2022-01-32", expectedResult: 400, expectedBodyRegExp: "parse.*to"},
		{name: "for invalid from return 400", url: "/pictures?from=2022-01-32&to=2022-02-05", expectedResult: 400, expectedBodyRegExp: "parse.*from"},
		{name: "for invalid from return 400", url: "/pictures?from=2022-01-32&to=2022-02-05", expectedResult: 400, expectedBodyRegExp: "parse.*from"},
		// these should be in the APODProvider test
		//{name: "for future dates return 400", url: "/pictures?from=2042-01-01&to=2042-01-05", expectedResult: 400, expectedBodyRegExp: "future"},
		//{name: "for dates before APOD started return 400", url: "/pictures?from=1995-06-15&to=1995-06-15", expectedResult: 400, expectedBodyRegExp: "before 1995-06-16"},
		{
			name: "pass range to provider",
			url:  "/pictures?from=2022-01-01&to=2022-02-05",

			providerCall: func(t *testing.T, from, to time.Time) ([]string, error) {
				expFrom, _ := time.Parse("2006-01-02", "2022-01-01")
				if !from.Equal(expFrom) {
					t.Errorf("expected from equal %v, got %v", expFrom, from)
				}
				expTo, _ := time.Parse("2006-01-02", "2022-02-05")
				if !to.Equal(expTo) {
					t.Errorf("expected to equal %v, got %v", expTo, to)
				}
				return []string{}, nil
			},

			expectedResult: 200,
		},
		{
			name: "if provider returns request error return 400 and include message",
			url:  "/pictures?from=2022-01-01&to=2022-02-05",

			providerCall: func(t *testing.T, from, to time.Time) ([]string, error) {
				return nil, BadRequestErrorf("TestError")
			},

			expectedResult:     400,
			expectedBodyRegExp: "TestError",
		},
		{
			name: "if provider returns custom error return 500 and do not include message",
			url:  "/pictures?from=2022-01-01&to=2022-02-05",

			providerCall: func(t *testing.T, from, to time.Time) ([]string, error) {
				return nil, fmt.Errorf("test")
			},

			expectedResult:        500,
			expectedBodyNotRegExp: "test",
		},
		{
			name: "if provider returns result return 200 and the result",
			url:  "/pictures?from=2022-01-01&to=2022-02-05",

			providerCall: func(t *testing.T, from, to time.Time) ([]string, error) {
				return []string{
					"testresult1",
					"testresult2",
					"testresult3",
				}, nil
			},

			expectedResult:     200,
			expectedBodyRegExp: "(?s)testresult1.*testresult2.*testresult3",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			server := Server{
				Provider: &mockProvider{
					f: func(ctx context.Context, from, to time.Time) ([]string, error) {
						if tt.providerCall != nil {
							return tt.providerCall(t, from, to)
						}
						return []string{}, nil
					},
				},
			}

			u, err := url.ParseRequestURI(tt.url)
			if err != nil {
				t.Fatalf("cannot parse expected URI: %v", err)
			}
			m := http.MethodGet
			if tt.method != "" {
				m = tt.method
			}
			req := &http.Request{
				URL:    u,
				Method: m,
			}

			buf := &bytes.Buffer{}
			w := &responseWriter{Writer: buf}

			server.ServeHTTP(w, req)

			if w.Code != tt.expectedResult {
				t.Errorf("expected result %d, got %d", tt.expectedResult, w.Code)
			}

			bytes := buf.Bytes()
			if !json.Valid(bytes) {
				t.Errorf("response does not contain valid JSON")
			}

			if matched, _ := regexp.Match(tt.expectedBodyRegExp, bytes); !matched {
				t.Errorf("expected response matching %s, got %s", tt.expectedBodyRegExp, string(bytes))
			}
			if tt.expectedBodyNotRegExp != "" {
				if matched, _ := regexp.Match(tt.expectedBodyNotRegExp, bytes); matched {
					t.Errorf("expected response not matching %s, got %s", tt.expectedBodyNotRegExp, string(bytes))
				}
			}
		})
	}
}

type responseWriter struct {
	Code   int
	Writer io.Writer
}

func (responseWriter) Header() http.Header {
	return http.Header{}
}

func (r *responseWriter) Write(p []byte) (int, error) {
	if r.Code == 0 {
		r.Code = 200
	}
	return r.Writer.Write(p)
}

func (r *responseWriter) WriteHeader(code int) {
	r.Code = code
}

type mockProvider struct {
	f func(ctx context.Context, from, to time.Time) ([]string, error)
}

func (p mockProvider) GetPictures(ctx context.Context, from, to time.Time) ([]string, error) {
	return p.f(ctx, from, to)
}
