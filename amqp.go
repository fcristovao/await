package main

import (
	"context"
	"net/url"

	"github.com/streadway/amqp"
)

type amqpResource struct {
	url.URL
}

func (a amqpResource) Await(ctx context.Context) error {
	conn, err := amqp.Dial(a.String())
	if err != nil {
		return &unavailabilityError{Reason: err}
	}
	_ = conn.Close()
	return nil
}
