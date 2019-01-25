package docker

import (
	"encoding/base64"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"io"
)

type options struct {
	logger io.Writer

	pullImage     bool
	registryCreds string
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
