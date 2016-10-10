package main

import (
	"fmt"
	"net/url"

	"context"
)

type websocketResource struct {
	url.URL
}

func (*websocketResource) Await(context.Context) error {
	// TODO(uwe): Implement
	return fmt.Errorf("Not yet implemented. Pull-requests are welcome")
}
