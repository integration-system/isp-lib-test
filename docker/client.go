package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/integration-system/isp-lib/v2/config"
	"github.com/pkg/errors"
)

type ispDockerClient struct {
	c *client.Client
}

func (c *ispDockerClient) Close() error {
	return c.c.Close()
}

// returns the address available from both docker containers and host machine
// can be used to bind from the host machine and later access from docker containers
func (c *ispDockerClient) GetBridgeAddress() (string, error) {
	args := filters.NewArgs()
	args.Add("name", "bridge")
	opts := types.NetworkListOptions{Filters: args}
	networkList, err := c.c.NetworkList(context.Background(), opts)
	if err != nil {
		return "", errors.Wrap(err, "get bridge network")
	} else if len(networkList) == 0 {
		return "", errors.New("bridge network does not exist")
	} else if len(networkList[0].IPAM.Config) == 0 {
		return "", errors.New("bridge network has 0 configurations")
	}
	ip := networkList[0].IPAM.Config[0].Gateway
	return ip, nil
}

// create and run postgreSQL container
// expect image from https://hub.docker.com/_/postgres
func (c *ispDockerClient) RunPGContainer(image string, dbAndUserName string, password string, opts ...Option) (*ContainerContext, error) {
	vars := []string{
		fmt.Sprintf("POSTGRES_USER=%s", dbAndUserName),
		fmt.Sprintf("POSTGRES_PASSWORD=%s", password),
	}
	return c.runContainer(image, vars, opts...)
}

//create and run container from specified image
func (c *ispDockerClient) RunContainer(image string, opts ...Option) (*ContainerContext, error) {
	return c.runContainer(image, nil, opts...)
}

// create and run isp application container, override local and remote configuration through environment variables
// localConfig and remoteConfig can be map or struct
func (c *ispDockerClient) RunAppContainer(image string, localConfig, remoteConfig interface{}, opts ...Option) (*ContainerContext, error) {
	vars := make([]string, 0)
	if localConfig != nil {
		vars = append(vars, configToEnvVariables(localConfig, config.LocalConfigEnvPrefix, false)...)
	}
	if remoteConfig != nil {
		vars = append(vars, configToEnvVariables(remoteConfig, config.RemoteConfigEnvPrefix, true)...)
	}
	return c.runContainer(image, vars, opts...)
}

// create docker network with specified name
// NetworkContext.Close remove network
func (c *ispDockerClient) CreateNetwork(name string) (*NetworkContext, error) {
	ctx := &NetworkContext{client: c}

	net, err := c.c.NetworkCreate(context.Background(), name, types.NetworkCreate{
		CheckDuplicate: true,
	})
	if err != nil {
		return ctx, errors.Wrap(err, "network create")
	}
	ctx.id = net.ID
	ctx.name = name

	return ctx, nil
}

// create and run container from specified image
// dont pull image by default, use option PullImage to pull first
// never return nil ContainerContext
func (c *ispDockerClient) runContainer(image string, envVars []string, opts ...Option) (*ContainerContext, error) {
	ops := &options{}
	for _, v := range opts {
		v(ops)
	}
	if ops.imageName != "" {
		image = ops.imageName
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
		writer := ioutil.Discard
		if ops.logger != nil {
			writer = ops.logger
		}
		_, _ = io.Copy(writer, reader)
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
	if len(ops.volume) > 0 {
		if hostCfg == nil {
			hostCfg = &container.HostConfig{}
		}
		hostCfg.Binds = ops.volume
	}
	resp, err := c.c.ContainerCreate(context.Background(), &container.Config{
		Image:        image,
		Env:          envVars,
		ExposedPorts: ops.portSet,
	}, hostCfg, nil, nil, ops.name)
	if err != nil {
		return ctx, errors.Wrap(err, "create container")
	}

	ctx.containerId = resp.ID

	if ops.networkId != "" {
		if err := c.c.NetworkConnect(context.Background(), ops.networkId, resp.ID, nil); err != nil {
			return ctx, errors.Wrap(err, "network connect")
		}
	}

	err = c.c.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return ctx, errors.Wrap(err, "start container")
	}

	if ops.networkId != "" {
		containerInfo, err := c.c.ContainerInspect(context.Background(), resp.ID)
		if err != nil {
			return ctx, errors.Wrap(err, "container inspect")
		}
		ctx.ipAddr = containerInfo.NetworkSettings.Networks[ops.networkName].IPAddress
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
		ctx.logger = ops.logger
		go func() {
			_, _ = io.Copy(ops.logger, reader)
		}()
	}
	ctx.started = true
	return ctx, nil
}

func NewClient() (*ispDockerClient, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	return &ispDockerClient{
		c: cli,
	}, nil
}
