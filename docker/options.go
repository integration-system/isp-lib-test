package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"io"
)

type options struct {
	logger io.Writer

	pullImage     bool
	registryCreds string
	portMapping   nat.PortMap

	env []string

	name string

	network string
}

type Option func(opts *options)

// redirect docker container pulling and running logs
func WithLogger(logger io.Writer) Option {
	return func(opts *options) {
		opts.logger = logger
	}
}

func PullImage(registryLogin string, registryPassword string) Option {
	return func(opts *options) {
		opts.pullImage = true
		if registryLogin != "" {
			authConfig := types.AuthConfig{
				Username: registryLogin,
				Password: registryPassword,
			}
			encodedJSON, _ := json.Marshal(authConfig)
			opts.registryCreds = base64.URLEncoding.EncodeToString(encodedJSON)
		}
	}
}

func WithPortMapping(mapping map[string]string) Option {
	arr := make([]string, 0, len(mapping))
	for pub, priv := range mapping {
		arr = append(arr, fmt.Sprintf("%s:%s", pub, priv))
	}
	_, m, _ := nat.ParsePortSpecs(arr)
	return func(opts *options) {
		opts.portMapping = m
	}
}

func WithEnv(vars map[string]string) Option {
	arr := make([]string, 0, len(vars))
	for k, v := range vars {
		arr = append(arr, fmt.Sprintf("%s=%s", k, v))
	}
	return func(opts *options) {
		opts.env = arr
	}
}

func WithName(containerName string) Option {
	return func(opts *options) {
		opts.name = containerName
	}
}

func WithNetwork(ctx *NetworkContext) Option {
	return func(opts *options) {
		opts.network = ctx.id
	}
}
