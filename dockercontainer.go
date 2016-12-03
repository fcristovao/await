// Copyright (c) 2016 Betalo AB
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
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type dockercontainerResource struct {
	url.URL
}

func (r *dockercontainerResource) Await(ctx context.Context) error {
	dockerHost := r.Host
	containerID := r.Path
	if dockerHost == "" {
		// Assume socket file path to docker daemon was provided
		dir, file := filepath.Split(containerID)
		if dir != "" && dir != "/" {
			dockerHost = "unix://" + dir
			containerID = file
		}
	} else {
		dockerHost = "tcp://" + dockerHost
	}
	containerID = strings.TrimPrefix(containerID, "/")

	opts := parseFragment(r.Fragment)

	var containerName string
	if containerNames, ok := opts["name"]; ok && len(containerNames) > 0 {
		containerName = containerNames[0]
	}

	var containerImage string
	if containerImages, ok := opts["image"]; ok && len(containerImages) > 0 {
		containerImage = containerImages[0]
	}

	// IDEA(uwe): Allow to specify more container health statuses
	health := "healthy"

	// Might override. In the "worst" case the value is empty, then the docker
	// client picks a default socket path.
	if dockerHost != "" {
		os.Setenv("DOCKER_HOST", dockerHost)
	}

	// Be careful to not end up in the future. Docker version 1.12, which maps
	// to remote API version 1.24, introduced HEALTHCHECK.
	if _, ok := os.LookupEnv("DOCKER_API_VERSION"); !ok {
		os.Setenv("DOCKER_API_VERSION", "1.24")
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	if containerID == "" {
		if containerName == "" && containerImage == "" {
			return fmt.Errorf("container id empty")
		}

		containers, listErr := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
		if listErr != nil {
			return listErr
		}

		if containerName != "" {
			if containerID = findContainerIDByName(containers, containerName); containerID == "" {
				return &unavailabilityError{fmt.Errorf("container %s not found", containerName)}
			}
		} else if containerImage != "" {
			if containerID = findContainerIDByImage(containers, containerImage); containerID == "" {
				return &unavailabilityError{fmt.Errorf("container %s not found", containerImage)}
			}
		}
	}

	inspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		if client.IsErrContainerNotFound(err) {
			return &unavailabilityError{fmt.Errorf("container %s not found", containerID)}
		}
		return err
	}

	status, err := healthStatus(inspect)
	if err != nil {
		return &unavailabilityError{err}
	}

	if status != health {
		return &unavailabilityError{fmt.Errorf("container health status: %v", status)}
	}

	return nil
}

func findContainerIDByName(containers []types.Container, name string) string {
	name = "/" + name
	for _, container := range containers {
		idx := indexOf(container.Names, name)
		if idx != -1 {
			return container.ID
		}
	}
	return ""
}

func findContainerIDByImage(containers []types.Container, image string) string {
	for _, container := range containers {
		if container.Image == image {
			return container.ID
		}
	}
	return ""
}

func healthStatus(inspect types.ContainerJSON) (string, error) {
	// See https://docs.docker.com/engine/reference/api/images/event_state.png

	if inspect.State == nil {
		return "", fmt.Errorf("unknown container state")
	}

	if !inspect.State.Running {
		return "", fmt.Errorf("container not running")
	}

	if inspect.State.Paused {
		return "", fmt.Errorf("container paused")
	}

	if inspect.State.Restarting {
		return "", fmt.Errorf("container restarting")
	}

	if inspect.State.Health == nil {
		return "", fmt.Errorf("container without health check")
	}

	// See https://github.com/docker/docker/blob/master/api/types/types.go#L298

	return inspect.State.Health.Status, nil
}
