package docker

import (
	"fmt"
	"github.com/doublerebel/bellows"
	"strings"
)

func configToEnvVariables(localConfig interface{}, prefix string) []string {
	m := bellows.Flatten(localConfig)
	vars := make([]string, 0, len(m))
	for k, v := range m {
		value := fmt.Sprintf("%v", v)
		if value != "" {
			vars = append(vars, fmt.Sprintf("%s_%s=%s", strings.ToUpper(prefix), strings.ToUpper(k), value))
		}
	}
	return vars
}
