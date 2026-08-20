package goproxy

import "net/http"

// Doer is the HTTP transport seam used by the module provider client.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

var _ Doer = (*http.Client)(nil)
