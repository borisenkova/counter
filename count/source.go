package count

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Source struct {
	io.ReadCloser
	origin          string
	contentIsLoaded bool
	load            loadContentFunc
	cancel          context.CancelFunc
}

type loadContentFunc func(context.Context, string) (io.ReadCloser, context.CancelFunc, error)

const (
	errEmptyOriginStr   = "empty origin"
	errUnknownSourceStr = "unknown source"
)

func NewSource(origin string, timeout time.Duration) (*Source, error) {
	if len(origin) == 0 {
		return nil, errors.New(errEmptyOriginStr)
	}
	if isRegularFile(origin) {
		return &Source{origin: origin, load: fromFile}, nil
	}
	if isHTTPURL(origin) {
		return &Source{origin: origin, load: fromURL(timeout)}, nil
	}
	return nil, errors.New(errUnknownSourceStr)
}

const loadRepeatTimes = 10

func repeatUntilNotError(times int, f func() error) (err error) {
	for i := 0; i < times; i++ {
		err = f()
		if err == nil {
			return
		}
	}

	return
}

func (s *Source) get(ctx context.Context) error {
	if !s.contentIsLoaded {
		return repeatUntilNotError(loadRepeatTimes, func() error {
			readCloser, cancel, err := s.load(ctx, s.origin)
			if err != nil {
				return fmt.Errorf("can't load source data from origin %s: %s", s.origin, err)
			}

			s.cancel = cancel
			s.ReadCloser = readCloser
			s.contentIsLoaded = true
			return nil
		})
	}
	return nil
}

func (s *Source) Read(p []byte) (n int, err error) {
	if s.ReadCloser == nil {
		return 0, nil
	}

	return s.ReadCloser.Read(p)
}

func (s *Source) Close() error {
	if s.contentIsLoaded {
		if s.cancel != nil {
			s.cancel()
		}

		return s.ReadCloser.Close()
	}

	return nil
}

// isRegularFile returns true if file from provided filepath is regular file
func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.Mode().IsRegular()
}

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

func isHTTPURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	scheme := strings.ToLower(parsed.Scheme)
	return scheme == schemeHTTP || scheme == schemeHTTPS
}

func fromFile(ctx context.Context, name string) (io.ReadCloser, context.CancelFunc, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, nil, fmt.Errorf("can't open file %s: %v", name, err)
	}

	return file, nil, nil
}

func fromURL(timeout time.Duration) loadContentFunc {
	return func(c context.Context, url string) (io.ReadCloser, context.CancelFunc, error) {
		ctx, cancel := context.WithTimeout(c, timeout)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("can't create HTTP request from URL %s: %s", url, err)
		}

		resp, err := httpDo(ctx, req)
		if err != nil {
			return nil, nil, fmt.Errorf("can't load URL %s content: %v", url, err)
		}

		return resp.Body, cancel, nil
	}
}

// httpDo issues the HTTP request in goroutine and returns response and error if
// it occurred. If ctx.Done is closed while the request is running, httpDo
// cancels the request and returns ctx.Err.
// Slightly changed version of httpDo from
// https://blog.golang.org/context/google/google.go
func httpDo(ctx context.Context, req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	req = req.WithContext(ctx)
	resultChan := make(chan struct {
		r *http.Response
		e error
	}, 1)

	go func() {
		r, e := client.Do(req)
		resultChan <- struct {
			r *http.Response
			e error
		}{r: r, e: e}
	}()

	select {
	case <-ctx.Done():
		<-resultChan
		return nil, ctx.Err()
	case res := <-resultChan:
		return res.r, res.e
	}
}
