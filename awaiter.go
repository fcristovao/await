// Copyright (C) 2016-2018 Betalo AB
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"context"
	"errors"
	"time"
)

const retryDelay = 500 * time.Millisecond

type timeoutError struct {
	Reason error
}

// Error implements the error interface.
func (e *timeoutError) Error() string {
	return e.Reason.Error()
}

type awaiter struct {
	logger  *LevelLogger
	timeout time.Duration
}

func (a *awaiter) run(resources []resource) error {
	if a.logger == nil {
		a.logger = NewLogger(errorLevel)
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	var latestErr error

	go func() {
		for _, res := range resources {
			a.logger.Infof("Awaiting resource: %s", res)

			for {
				select {
				case <-ctx.Done():
					// Exceeded timeout
					return
				default:
					// Still time left, let's continue
				}

				if latestErr = res.Await(ctx); latestErr != nil {
					if e, ok := latestErr.(*unavailabilityError); ok {
						// transient error
						a.logger.Debugf("Resource unavailable: %v", e)
					} else {
						// Maybe transient error
						a.logger.Errorf("Error: failed to await resource: %v", latestErr)
					}
					time.Sleep(retryDelay)
				} else {
					// Resource found, move on to next one
					break
				}
			}
		}

		cancel() // All resources are available
	}()

	<-ctx.Done()
	switch ctx.Err() {
	case context.Canceled:
		if latestErr != nil {
			return latestErr
		}
		// All resources are available
		return nil
	case context.DeadlineExceeded:
		if latestErr == nil {
			// Time out even before the first try
			latestErr = errors.New("initial await did not finish")
		}
		return &unavailabilityError{latestErr}
	default:
		return errors.New("unknown error")
	}
}
