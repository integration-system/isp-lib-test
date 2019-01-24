package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"io"
)

type ispDockerClient struct {
	c *client.Client
}

func (c *ispDockerClient) Close() error {
	return c.c.Close()
}

func (c *ispDockerClient) RunPGContainer(image string, dbAndUserName string, password string, opts ...RunPGOption) (*ContainerContext, error) {
	ops := &runPgOptions{}
	for _, v := range opts {
		v(ops)
	}

	ctx := &ContainerContext{imageId: image, client: c}

	reader, err := c.c.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		return ctx, errors.Wrap(err, "pull image")
	}
	if ops.logger != nil {
		_, _ = io.Copy(ops.logger, reader)
	}

	vars := []string{
		fmt.Sprintf("POSTGRES_USER=%s", dbAndUserName),
		fmt.Sprintf("POSTGRES_PASSWORD=%s", password),
	}
	resp, err := c.c.ContainerCreate(context.Background(), &container.Config{
		Image: image,
		Env:   vars,
	}, nil, nil, "")
	if err != nil {
		return ctx, errors.Wrap(err, "create container")
	}
	ctx.containerId = resp.ID

	err = c.c.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return ctx, errors.Wrap(err, "start container")
	}

	if ops.logger != nil {
		reader, err := c.c.ContainerLogs(
			context.Background(),
			resp.ID,
			types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true},
		)
		if err != nil {
			return ctx, errors.Wrap(err, "attach logger")
		}
		go func() {
			_, _ = io.Copy(ops.logger, reader)
		}()
	}

	return ctx, nil
}

func NewClient() (*ispDockerClient, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &ispDockerClient{
		c: cli,
	}, nil
}
