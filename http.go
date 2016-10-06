package main

import (
	"net/http"
	"net/url"

	"context"
)

type httpResource struct {
	url.URL
}

func (r *httpResource) Await(ctx context.Context) error {
	client := &http.Client{}

	// TODO(uwe): Use fragment to set method

	req, err := http.NewRequest("GET", r.URL.String(), nil)
	if err != nil {
		return err
	}

	// TODO(uwe): Use k/v pairs in fragment to set headers

	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return ErrUnavailable
	}
	defer resp.Body.Close()

	// TODO(uwe): Use fragment to set tolerated status code

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return ErrUnavailable
}
