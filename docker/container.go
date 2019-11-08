package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"strings"
)

type ContainerContext struct {
	imageId     string
	containerId string
	client      *ispDockerClient
	ipAddr      string
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

func (ctx *ContainerContext) GetIPAddress() string {
	return ctx.ipAddr
}
