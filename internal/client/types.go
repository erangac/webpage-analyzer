package client

import "context"

// HTTPClient defines the interface for HTTP operations.
type HTTPClient interface {
	FetchWebpage(ctx context.Context, url string) ([]byte, int, error)
	ParseHTML(content []byte) (interface{}, error)
}
