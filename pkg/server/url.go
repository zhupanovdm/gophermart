package server

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

type UrlMatcher struct {
	urlPatterns []*regexp.Regexp
}

func (r *UrlMatcher) SetPattern(patterns ...string) error {
	for i, p := range patterns {
		expr, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("failed to add pattern to filter (%d): %w", i, err)
		}
		r.urlPatterns = append(r.urlPatterns, expr)
	}
	return nil
}

func (r *UrlMatcher) Match(req *http.Request) bool {
	for _, pattern := range r.urlPatterns {
		if pattern.MatchString(req.URL.Path) {
			return true
		}
	}
	return false
}

func NewURLMatcher() *UrlMatcher {
	return &UrlMatcher{
		urlPatterns: make([]*regexp.Regexp, 0),
	}
}

func ParseURL(addr string) (*url.URL, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid destination address: %s: %w", addr, err)
	}
	if u.Host == "" {
		return ParseURL("http://" + addr)
	}
	return u, nil
}
