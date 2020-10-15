package tea

import "strings"

type EqArgsValidator interface {
	Valid() bool
}

type EqArgsPresenceValidator struct {
	args *EqArgs
	keys []string
}

type EqArgs struct {
	Raw        map[string]string
	validators []EqArgsValidator
}

func ParseEqArgs(args []string) *EqArgs {
	aa := EqArgs{
		Raw:        make(map[string]string, 0),
		validators: make([]EqArgsValidator, 0),
	}
	for _, arg := range args {
		kv := strings.SplitN(arg, "=", 2)
		aa.Raw[kv[0]] = kv[1]
	}

	return &aa
}

func (a *EqArgs) Valid() bool {
	for _, v := range a.validators {
		if !v.Valid() {
			return false
		}
	}

	return true
}

func (a *EqArgs) ValidatePresence(keys ...string) {
	v := EqArgsPresenceValidator{
		args: a,
		keys: keys,
	}
	a.validators = append(a.validators, &v)
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

func (v *EqArgsPresenceValidator) Valid() bool {
	for _, k := range v.keys {
		if _, ok := v.args.Raw[k]; !ok {
			return false
		}
	}

	return true
}
