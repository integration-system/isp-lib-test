package docker

import (
	"fmt"
	"github.com/doublerebel/bellows"
	"strings"
)

func configToEnvVariables(config interface{}, prefix string) []string {
	m := bellows.Flatten(config)
	vars := make([]string, 0, len(m))
	for k, v := range m {
		if v == nil {
			continue
		}
		value := fmt.Sprintf("%v", v)
		if value != "" {
			vars = append(vars, fmt.Sprintf("%s_%s=%s", strings.ToUpper(prefix), strings.ToUpper(k), value))
		}
	}
	return vars
}
