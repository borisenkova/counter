package count

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Source struct {
	io.ReadCloser
	origin          string
	contentIsLoaded bool
	load            loadContentFunc
}

type loadContentFunc func(string) (io.ReadCloser, error)

func NewSource(origin string, httpClient *http.Client) (*Source, error) {
	if len(origin) == 0 {
		return nil, errors.New("empty.origin")
	}
	if isFile(origin) {
		return &Source{origin: origin, load: fromFile}, nil
	}
	if isHTTPURL(origin) {
		return &Source{origin: origin, load: fromURL(httpClient)}, nil
	}
	return nil, errors.New("unknown.source")
}

func (s *Source) Read(p []byte) (n int, err error) {
	if !s.contentIsLoaded {
		readCloser, err := s.load(s.origin)
		if err != nil {
			return 0, err
		}

		s.ReadCloser = readCloser
		s.contentIsLoaded = true
	}

	return s.ReadCloser.Read(p)
}

func (s *Source) Close() error {
	if s.contentIsLoaded {
		return s.ReadCloser.Close()
	}

	return nil
}

func isFile(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
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

func fromFile(name string) (io.ReadCloser, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("can't open file %s: %v", name, err)
	}

	return file, nil
}

func fromURL(httpClient *http.Client) loadContentFunc {
	return func(url string) (io.ReadCloser, error) {
		resp, err := httpClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("can't load URL %s content: %v", url, err)
		}

		return resp.Body, nil
	}
}
