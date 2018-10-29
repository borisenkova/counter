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

type source interface {
	io.ReadCloser
	origin() string
	load(ctx context.Context) error
}

const (
	errEmptyOriginStr   = "empty origin"
	errUnknownSourceStr = "unknown source"
)

func newSource(origin string, timeout time.Duration) (source, error) {
	if len(origin) == 0 {
		return nil, errors.New(errEmptyOriginStr)
	}
	if isRegularFile(origin) {
		source := &fileSource{}
		source.origin_ = origin
		return source, nil
	}
	if isHTTPURL(origin) {
		source := &urlSource{}
		source.origin_ = origin
		source.timeout = timeout
		return source, nil
	}
	return nil, errors.New(errUnknownSourceStr)
}

type content struct {
	io.ReadCloser
	contentIsLoaded bool
	origin_         string
}

func (c *content) Read(p []byte) (n int, err error) {
	if c.ReadCloser == nil {
		return 0, nil
	}

	return c.ReadCloser.Read(p)
}

func (c *content) Close() error {
	if c.contentIsLoaded {
		return c.ReadCloser.Close()
	}

	return nil
}

func (c *content) origin() string {
	return c.origin_
}

type fileSource struct {
	content
}

func (s *fileSource) load(ctx context.Context) error {
	if !s.contentIsLoaded {
		readCloser, err := loadFromFile(s.origin_)
		if err != nil {
			return err
		}

		s.contentIsLoaded = true
		s.ReadCloser = readCloser
	}
	return nil
}

func loadFromFile(name string) (io.ReadCloser, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("can't open file %s: %v", name, err)
	}

	return file, nil
}

// isRegularFile returns true if file from provided filepath is regular file
func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.Mode().IsRegular()
}

type urlSource struct {
	content
	timeout time.Duration
	cancel  context.CancelFunc
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

func (s *urlSource) Close() error {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	return s.content.Close()
}

func (s *urlSource) load(ctx context.Context) error {
	if !s.contentIsLoaded {
		return repeatUntilNotError(loadRepeatTimes, func() error {
			readCloser, cancel, err := loadFromURL(ctx, s.origin_, s.timeout)
			if err != nil {
				return fmt.Errorf("can't load source data from origin %s: %s", s.origin_, err)
			}

			s.cancel = cancel
			s.ReadCloser = readCloser
			s.contentIsLoaded = true
			return nil
		})
	}
	return nil
}

func loadFromURL(c context.Context, url string, timeout time.Duration) (io.ReadCloser, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(c, timeout)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	// Request identity encoding to avoid 'unexpected EOF' error if content
	// is gzipped
	// via https://stackoverflow.com/a/21160982
	req.Header.Add("Accept-Encoding", "identity")
	if err != nil {
		return nil, nil, fmt.Errorf("can't create HTTP request from URL %s: %s", url, err)
	}

	resp, err := httpDo(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("can't load URL %s content: %v", url, err)
	}

	return resp.Body, cancel, nil
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
