package main

import (
	"net"
	"net/url"

	"context"
)

type tcpResource struct {
	url.URL
}

func (r *tcpResource) Await(ctx context.Context) error {
	dialer := &net.Dialer{}
	_, err := dialer.DialContext(ctx, r.URL.Scheme, r.URL.Host)
	if err != nil {
		return &unavailableError{err}
	}

	return nil
}
