package docker

import (
	"context"
	"github.com/pkg/errors"
)

type NetworkContext struct {
	client *ispDockerClient
	id     string
}

func (ctx *NetworkContext) Close() error {
	if ctx.id != "" {
		if err := ctx.client.c.NetworkRemove(context.Background(), ctx.id); err != nil {
			return errors.Wrap(err, "network remove")
		}
	}
	return nil
}
