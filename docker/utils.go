package docker

import (
	"fmt"
	"github.com/doublerebel/bellows"
)

func configToEnvVariables(localConfig interface{}, prefix string) []string {
	m := bellows.FlattenPrefixed(localConfig, prefix)
	vars := make([]string, 0, len(m))
	for k, v := range m {
		value := fmt.Sprintf("%v", v)
		if value != "" {
			vars = append(vars, fmt.Sprintf("%s=%s", k, value))
		}
	}
	return vars
}
