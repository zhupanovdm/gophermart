package handlers

import (
	"fmt"
	"net/http"
	"regexp"
)

type RequestMatcher struct {
	urlPatterns []*regexp.Regexp
}

func (r *RequestMatcher) URLPattern(patterns ...string) error {
	for i, p := range patterns {
		expr, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("failed to add pattern to filter (%d): %w", i, err)
		}
		r.urlPatterns = append(r.urlPatterns, expr)
	}
	return nil
}

func (r *RequestMatcher) MatchURL(req *http.Request) bool {
	for _, pattern := range r.urlPatterns {
		if pattern.MatchString(req.URL.Path) {
			return true
		}
	}
	return false
}

func NewRequestMatcher() *RequestMatcher {
	return &RequestMatcher{
		urlPatterns: make([]*regexp.Regexp, 0),
	}
}
