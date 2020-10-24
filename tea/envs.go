package tea

import (
	"os"
	"strconv"
)

// EnvGetOr returns env value if present otherwise returns fallback value.
//
func EnvGetOr(key string, fallback string) string {
	env := os.Getenv(key)
	if env != "" {
		return env
	}
	return fallback
}

// EnvGetIntOr returns env value if present otherwise returns fallback value.
//
func EnvGetIntOr(key string, fallback int) int {
	env := os.Getenv(key)
	if i, err := strconv.ParseInt(env, 10, 64); err != nil {
		return fallback
	} else {
		return int(i)
	}
}
