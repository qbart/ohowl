package tea

import (
	"fmt"
	"strings"
)

type EqArgsValidator interface {
	Valid() bool
}

type EqArgsPresenceValidator struct {
	args *EqArgs
	keys []string
}

type EqArgsInclusionValidator struct {
	args   *EqArgs
	key    string
	values []string
}

type EqArgs struct {
	Raw        map[string]string
	Errors     []error
	validators []EqArgsValidator
}

func ParseEqArgs(args []string) *EqArgs {
	aa := EqArgs{
		Raw:        make(map[string]string, 0),
		Errors:     make([]error, 0),
		validators: make([]EqArgsValidator, 0),
	}
	for _, arg := range args {
		kv := strings.SplitN(arg, "=", 2)
		aa.Raw[kv[0]] = kv[1]
	}

	return &aa
}

func (a *EqArgs) ErrorMessages() string {
	sb := strings.Builder{}
	for i, e := range a.Errors {
		sb.WriteString(e.Error())
		sb.WriteString(".")
		if i < len(a.Errors)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
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

func (a *EqArgs) ValidateInclusion(key string, values []string) {
	v := EqArgsInclusionValidator{
		args:   a,
		key:    key,
		values: values,
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
	valid := true
	for _, k := range v.keys {
		if _, ok := v.args.Raw[k]; !ok {
			v.args.Errors = append(v.args.Errors, fmt.Errorf("%s is missing", k))
			valid = false
		}
	}

	return valid
}

func (v *EqArgsInclusionValidator) Valid() bool {
	s := v.args.Raw[v.key]
	for _, f := range v.values {
		if s == f {
			return true
		}
	}

	v.args.Errors = append(v.args.Errors, fmt.Errorf("%s must be one of: %s", v.key, strings.Join(v.values, ", ")))

	return false
}
