package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type ContainerContext struct {
	imageId     string
	containerId string
	client      *ispDockerClient
	ipAddr      string
	started     bool
}

// force delete container and image
func (ctx *ContainerContext) Close() error {
	err := ctx.ForceRemoveContainer()

	removeImageErr := ctx.ForceRemoveImage()
	if removeImageErr != nil {
		if err == nil {
			err = removeImageErr
		} else {
			err = errors.New(strings.Join([]string{err.Error(), removeImageErr.Error()}, "; "))
		}
	}

	return err
}

func (ctx *ContainerContext) ForceRemoveContainer() error {
	if ctx.containerId != "" {
		err := ctx.client.c.ContainerRemove(
			context.Background(),
			ctx.containerId,
			types.ContainerRemoveOptions{Force: true},
		)
		if err != nil {
			return errors.Wrap(err, "container remove")
		}
		ctx.containerId = ""
	}

	return nil
}

func (ctx *ContainerContext) ForceRemoveImage() error {
	if ctx.imageId != "" {
		_, err := ctx.client.c.ImageRemove(
			context.Background(),
			ctx.imageId,
			types.ImageRemoveOptions{Force: true},
		)
		if err != nil {
			return errors.Wrap(err, "image remove")
		}
		ctx.imageId = ""
	}

	return nil
}

// StopContainer stops a container without terminating the process.
// The process is blocked until the container stops or the timeout expires.
func (ctx *ContainerContext) StopContainer(timeout time.Duration) error {
	if ctx.containerId != "" && ctx.started {
		err := ctx.client.c.ContainerStop(
			context.Background(),
			ctx.containerId,
			&timeout,
		)
		if err != nil {
			return errors.Wrap(err, "container stop")
		}
	}
	ctx.started = false
	return nil
}

// StartContainer sends a request to the docker daemon to start a container.
func (ctx *ContainerContext) StartContainer() error {
	if ctx.containerId != "" && !ctx.started {
		err := ctx.client.c.ContainerStart(
			context.Background(),
			ctx.containerId,
			types.ContainerStartOptions{},
		)
		if err != nil {
			return errors.Wrap(err, "container start")
		}
	}
	ctx.started = true
	return nil
}

func (ctx *ContainerContext) GetIPAddress() string {
	return ctx.ipAddr
}
