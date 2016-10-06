package main

import (
	"net/url"
	"os"
	"path/filepath"

	"context"
)

type fileResource struct {
	url.URL
}

func (r *fileResource) Await(context.Context) error {
	// Check for relative paths
	filePath := r.URL.Path
	if r.URL.Host != "" {
		filePath = filepath.Join(r.URL.Host, filePath)
	}

	tags := parseTags(r.URL.Fragment)

	_, err := os.Stat(filePath)
	if _, ok := tags["absent"]; ok {
		if err == nil {
			return ErrUnavailable
		} else if os.IsNotExist(err) {
			return nil
		}
	} else {
		if err == nil {
			return nil
		} else if os.IsNotExist(err) {
			return ErrUnavailable
		}
	}

	return err
}
