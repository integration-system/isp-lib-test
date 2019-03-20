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
	portBinding   nat.PortMap
	portSet       nat.PortSet

	env []string

	name string

	network string

	volume []string
}

type Option func(opts *options)

// redirect docker container runtime logs
func WithLogger(logger io.Writer) Option {
	return func(opts *options) {
		opts.logger = logger
	}
}

// if set image pulls first, registryLogin and registryPassword are optional
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

// bind ports host_machine_port -> container_exposed_port
func WithPortBindings(mapping map[string]string) Option {
	arr := make([]string, 0, len(mapping))
	for pub, priv := range mapping {
		arr = append(arr, fmt.Sprintf("%s:%s", pub, priv))
	}
	portSet, bindings, _ := nat.ParsePortSpecs(arr)
	return func(opts *options) {
		opts.portBinding = bindings
		opts.portSet = portSet
	}
}

// set environments variables
func WithEnv(vars map[string]string) Option {
	arr := make([]string, 0, len(vars))
	for k, v := range vars {
		arr = append(arr, fmt.Sprintf("%s=%s", k, v))
	}
	return func(opts *options) {
		opts.env = arr
	}
}

// set container name and hostname in docker network by default
func WithName(containerName string) Option {
	return func(opts *options) {
		opts.name = containerName
	}
}

// if set, container joins to specified network
func WithNetwork(ctx *NetworkContext) Option {
	return func(opts *options) {
		opts.network = ctx.id
	}
}

//
func WithVolumes(volume map[string]string) Option {
	arr := make([]string, 0, len(volume))
	for pub, priv := range volume {
		arr = append(arr, fmt.Sprintf("%s:%s", pub, priv))
	}
	return func(opts *options) {
		opts.volume = arr
	}
}
