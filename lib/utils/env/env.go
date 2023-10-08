package env

import (
	"lib/tlog"
	"os"
	"strconv"
	"strings"
)

type types interface {
	string | bool | int | float64 | []string
}

// Get gets variable from env. If not found returns default value.
// If defaultValue is set and variable not found, then panics.
func Get[T types](envName string, defaultValue ...T) T {
	value := os.Getenv(envName)

	var ret any = value
	var err error

	var def T
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}

	switch any(def).(type) {
	case string:
		ret = value

	case bool:
		ret, err = strconv.ParseBool(value)

	case int:
		ret, err = strconv.Atoi(value)

	case float64:
		ret, err = strconv.ParseFloat(value, 64)

	case []string:
		if strings.Contains(value, ";") {
			ret = strings.Split(value, ";")
		} else {
			ret = strings.Split(value, ",")
		}
	}

	switch {
	case value == "" && len(defaultValue) == 0:
		tlog.Fatal("Required variable {{name}} is not set - type: {{type}}", tlog.Vars{
			"name": envName,
			"type": def,
		})

	case value == "":
		ret = def

	case err != nil:
		tlog.Fatal("Variable {{name}} could not be parsed - type: {{type}}, value: {{value}}", tlog.Vars{
			"name":  envName,
			"type":  def,
			"value": value,
		})
	}

	return ret.(T)
}
