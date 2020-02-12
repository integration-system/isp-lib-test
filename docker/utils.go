package docker

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/integration-system/bellows"
	"github.com/integration-system/isp-lib/v2/config"
)

func configToEnvVariables(config interface{}, prefix string, withTypes bool) []string {
	m := bellows.Flatten(config)
	vars := make([]string, 0, len(m))
	for k, v := range m {
		if v == nil {
			continue
		}
		value, t := toString(v)
		if value != "" {
			env := ""
			if withTypes {
				env = fmt.Sprintf("%s_%s=%s#{%s}", strings.ToUpper(prefix), strings.ToUpper(k), value, t)
			} else {
				env = fmt.Sprintf("%s_%s=%s", strings.ToUpper(prefix), strings.ToUpper(k), value)
			}
			vars = append(vars, env)
		}
	}
	return vars
}

// avoid zero values
func toString(value interface{}) (string, config.PropertyType) {
	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.String:
		s := rv.String()
		return s, config.String
	case
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		if rv.Int() == 0 {
			return "", config.Int
		}
		s := strconv.Itoa(int(rv.Int()))
		return s, config.Int
	case reflect.Float32:
		if rv.Float() == 0 {
			return "", config.Float32
		}
		s := strconv.FormatFloat(rv.Float(), 'f', -1, 32)
		return s, config.Float32
	case reflect.Float64:
		if rv.Float() == 0 {
			return "", config.Float64
		}
		s := strconv.FormatFloat(rv.Float(), 'f', -1, 64)
		return s, config.Float64
	case reflect.Bool:
		if !rv.Bool() {
			return "", config.Bool
		}
		s := strconv.FormatBool(rv.Bool())
		return s, config.Bool
	case reflect.Interface, reflect.Ptr:
		if rv.IsNil() {
			return "", ""
		}
		el := rv.Elem().Interface()
		return toString(el)
	default:
		return "", ""
	}
}
