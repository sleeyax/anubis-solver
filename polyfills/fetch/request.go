package fetch

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	RequestRedirectError  = "error"
	RequestRedirectFollow = "follow"
	RequestRedirectManual = "manual"
)

type RequestInit struct {
	Body     string            `json:"body"`
	Headers  map[string]string `json:"headers"`
	Method   string            `json:"method"`
	Redirect string            `json:"redirect"`
}

type Request struct {
	Body       string
	Method     string
	Redirect   string
	Header     http.Header
	URL        *url.URL
	RemoteAddr string
}

func ParseRequestURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("url '%s' is not valid, %w", rawURL, err)
	}

	switch u.Scheme {
	case "http", "https":
	case "": // then scheme is empty, it's a local request
		if !strings.HasPrefix(u.Path, "/") {
			return nil, fmt.Errorf("unsupported relatve path %s", u.Path)
		}
	default:
		return nil, fmt.Errorf("unsupported scheme %s", u.Scheme)
	}

	return u, nil
}
