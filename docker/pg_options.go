package docker

import "io"

type runPgOptions struct {
	logger io.Writer
}

type RunPGOption func(opts *runPgOptions)

// redirect docker container pulling and running logs
func WithLogger(logger io.Writer) RunPGOption {
	return func(opts *runPgOptions) {
		opts.logger = logger
	}
}
