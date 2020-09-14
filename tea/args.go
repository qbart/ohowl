package tea

import "strings"

type EqArgs struct {
	Raw map[string]string
}

func ParseEqArgs(args []string) *EqArgs {
	aa := EqArgs{
		Raw: make(map[string]string, 0),
	}
	for _, arg := range args {
		kv := strings.SplitN(arg, "=", 2)
		aa.Raw[kv[0]] = kv[1]
	}

	return &aa
}

func (a *EqArgs) GetBoolDefault(key string, defaultValue bool) bool {
	v := a.Raw[key]
	switch v {
	case "true":
		return true
	case "false":
		return false
	default:
		return defaultValue
	}
}

func (a *EqArgs) GetString(key string) string {
	return a.Raw[key]
}

func (a *EqArgs) GetStrings(key string, sep string) []string {
	return strings.Split(a.Raw[key], sep)
}

func (a *EqArgs) Exist(keys ...string) bool {
	for _, k := range keys {
		if _, ok := a.Raw[k]; !ok {
			return false
		}
	}

	return true
}
