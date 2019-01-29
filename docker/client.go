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
	"os"
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
	return c.runContainer(image, vars, opts...)
}

func (c *ispDockerClient) RunRabbitContainer(image string, opts ...Option) (*ContainerContext, error) {
	return c.runContainer(image, nil, opts...)
}

// create and run container from specified image
// not pull image by default, use option PullImage to pull first
// localConfig and remoteConfig can be map or struct
func (c *ispDockerClient) RunAppContainer(image string, localConfig, remoteConfig interface{}, opts ...Option) (*ContainerContext, error) {
	vars := make([]string, 0)
	if localConfig != nil {
		vars = append(vars, configToEnvVariables(localConfig, config.LocalConfigEnvPrefix)...)
	}
	if remoteConfig != nil {
		vars = append(vars, configToEnvVariables(remoteConfig, config.RemoteConfigEnvPrefix)...)
	}
	return c.runContainer(image, vars, opts...)
}

func (c *ispDockerClient) CreateNetwork(name string) (*NetworkContext, error) {
	ctx := &NetworkContext{client: c}

	net, err := c.c.NetworkCreate(context.Background(), name, types.NetworkCreate{})
	if err != nil {
		return ctx, errors.Wrap(err, "network create")
	}
	ctx.id = net.ID

	return ctx, nil
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
		_, _ = io.Copy(os.Stdout, reader)
	}

	if envVars != nil {
		envVars = append(envVars, ops.env...)
	} else {
		envVars = ops.env
	}
	var hostCfg *container.HostConfig = nil
	if len(ops.portBinding) > 0 {
		hostCfg = &container.HostConfig{PortBindings: ops.portBinding}
	}
	resp, err := c.c.ContainerCreate(context.Background(), &container.Config{
		Image:        image,
		Env:          envVars,
		ExposedPorts: ops.portSet,
	}, hostCfg, nil, ops.name)
	if err != nil {
		return ctx, errors.Wrap(err, "create container")
	}
	ctx.containerId = resp.ID

	if ops.network != "" {
		if err := c.c.NetworkConnect(context.Background(), ops.network, resp.ID, nil); err != nil {
			return ctx, errors.Wrap(err, "network connect")
		}
	}

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
