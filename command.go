package main

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"context"
)

type commandResource struct {
	url.URL
}

func (r *commandResource) Await(ctx context.Context) error {
	// FIXME(uwe): Splitting by space is brittle
	cmdString, err := url.QueryUnescape(r.URL.Path)
	if err != nil {
		return err
	}
	cmdParts := strings.SplitN(cmdString, " ", 2)
	if len(cmdParts) == 0 {
		return fmt.Errorf("empty command")
	}
	cmd := cmdParts[0]
	args := cmdParts[1:]

	if err := exec.CommandContext(ctx, cmd, args...).Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return ErrUnavailable
		}
		return err
	}

	return nil
}
