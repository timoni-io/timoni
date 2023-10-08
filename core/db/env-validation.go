package db

import (
	log "lib/tlog"
	"strings"
)

func (env *EnvironmentS) CheckName() *log.RecordS {
	if env.Name == "" {
		return log.Error("name is required")
	}

	if !reSimpleName1.MatchString(env.Name) {
		return log.Error("name contains characters that are not allowed: " + env.Name)
	}
	if strings.Contains(env.Name, "--") {
		return log.Error("name contains `--` that are not allowed")
	}
	if strings.HasPrefix(env.Name, "-") {
		return log.Error("name starts with `-` which is not allowed")
	}
	if strings.HasSuffix(env.Name, "-") {
		return log.Error("name ends with `-` which is not allowed")
	}
	if len(env.Name) > 40 {
		return log.Error("name is too long (max 40 chars)")
	}

	return nil
}

func (env *EnvironmentS) CheckID() *log.RecordS {

	errMsg := "ID max length is 16 chars, allowed chars: a-z, 0-9 and '-' ('--' is not!) and must start with a-z or 0-9"

	if !reSimpleName1.MatchString(env.ID) {
		return log.Error(errMsg)
	}
	if strings.Contains(env.ID, "--") {
		return log.Error(errMsg)
	}
	if strings.HasPrefix(env.ID, "-") {
		return log.Error(errMsg)
	}
	if strings.HasSuffix(env.ID, "-") {
		return log.Error(errMsg)
	}
	if len(env.ID) > 16 {
		return log.Error(errMsg)
	}

	return nil
}
