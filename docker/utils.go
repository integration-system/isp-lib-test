package docker

import (
	"fmt"
	"github.com/doublerebel/bellows"
	"reflect"
	"strconv"
	"strings"
)

func configToEnvVariables(config interface{}, prefix string) []string {
	m := bellows.Flatten(config)
	vars := make([]string, 0, len(m))
	for k, v := range m {
		if v == nil {
			continue
		}
		value := toString(v)
		if value != "" {
			vars = append(vars, fmt.Sprintf("%s_%s=%s", strings.ToUpper(prefix), strings.ToUpper(k), value))
		}
	}
	return vars
}

// avoid zero values
func toString(value interface{}) string {
	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.String:
		s := rv.String()
		return s
	case
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		if rv.Int() == 0 {
			return ""
		}
		s := strconv.Itoa(int(rv.Int()))
		return s
	case reflect.Float32:
		if rv.Float() == 0 {
			return ""
		}
		s := strconv.FormatFloat(rv.Float(), 'f', -1, 32)
		return s
	case reflect.Float64:
		if rv.Float() == 0 {
			return ""
		}
		s := strconv.FormatFloat(rv.Float(), 'f', -1, 64)
		return s
	case reflect.Bool:
		if !rv.Bool() {
			return ""
		}
		s := strconv.FormatBool(rv.Bool())
		return s
	case reflect.Interface, reflect.Ptr:
		if rv.IsNil() {
			return ""
		}
		el := rv.Elem().Interface()
		return toString(el)
	default:
		return ""
	}
}
