package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
)

type ContainerContext struct {
	imageId     string
	containerId string
	client      *ispDockerClient
}

// force delete container and image
func (ctx *ContainerContext) Close() error {
	if err := ctx.ForceRemoveContainer(); err != nil {
		return err
	}

	if err := ctx.ForceDeleteImage(); err != nil {
		return err
	}

	return nil
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
	}

	return nil
}

func (ctx *ContainerContext) ForceDeleteImage() error {
	if ctx.imageId != "" {
		_, err := ctx.client.c.ImageRemove(
			context.Background(),
			ctx.imageId,
			types.ImageRemoveOptions{Force: true},
		)
		if err != nil {
			return errors.Wrap(err, "image remove")
		}
	}

	return nil
}
