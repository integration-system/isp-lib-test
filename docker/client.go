package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/integration-system/isp-lib/config"
	"github.com/pkg/errors"
	"io"
)

type ispDockerClient struct {
	c *client.Client
}

func (c *ispDockerClient) Close() error {
	return c.c.Close()
}

func (c *ispDockerClient) RunPGContainer(image string, dbAndUserName string, password string, opts ...Option) (*ContainerContext, error) {
	vars := []string{
		fmt.Sprintf("POSTGRES_USER=%s", dbAndUserName),
		fmt.Sprintf("POSTGRES_PASSWORD=%s", password),
	}
	opts = append(opts, WithPortMapping(map[string]string{"5432": "5432"}))
	return c.runContainer(image, vars, opts...)
}

func (c *ispDockerClient) RunAppContainer(image string, localConfig interface{}, opts ...Option) (*ContainerContext, error) {
	vars := configToEnvVariables(localConfig, config.EnvPrefix+"_")
	return c.runContainer(image, vars, opts...)
}

func (c *ispDockerClient) runContainer(image string, envVars []string, opts ...Option) (*ContainerContext, error) {
	ops := &options{}
	for _, v := range opts {
		v(ops)
	}

	ctx := &ContainerContext{client: c}

	if ops.pullImage {
		pullOpts := types.ImagePullOptions{}
		if ops.registryCreds != "" {
			pullOpts.RegistryAuth = ops.registryCreds
		}
		reader, err := c.c.ImagePull(context.Background(), image, pullOpts)
		if err != nil {
			return ctx, errors.Wrap(err, "pull image")
		}
		ctx.imageId = image
		if ops.logger != nil {
			_, _ = io.Copy(ops.logger, reader)
		}
	}

	envVars = append(envVars, ops.env...)
	resp, err := c.c.ContainerCreate(context.Background(), &container.Config{
		Image: image,
		Env:   envVars,
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
