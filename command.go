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
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

type commandResource struct {
	url.URL
}

func (r *commandResource) Await(ctx context.Context) error {
	cmdString, err := url.QueryUnescape(r.URL.Path)
	if err != nil {
		return err
	}
	// TODO(uwe): Splitting by space is brittle
	cmdParts := strings.SplitN(cmdString, " ", 2)
	if len(cmdParts) == 0 {
		return fmt.Errorf("empty command")
	}
	cmd := cmdParts[0]
	args := cmdParts[1:]

	if err := exec.CommandContext(ctx, cmd, args...).Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &unavailabilityError{exitErr}
		}
		return err
	}

	return nil
}
