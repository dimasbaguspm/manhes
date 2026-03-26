package httputil

import (
	"net/http"
	"time"
)

const defaultUserAgent = "manga-engine/1.0"

// NewClient returns an HTTP client with sensible defaults.
func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &userAgentTransport{
			ua:   defaultUserAgent,
			base: http.DefaultTransport,
		},
	}
}

type userAgentTransport struct {
	ua   string
	base http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	r.Header.Set("User-Agent", t.ua)
	return t.base.RoundTrip(r)
}
