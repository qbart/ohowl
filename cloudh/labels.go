package cloudh

import (
	"fmt"
	"strings"
)

// LabelSelector
// https://docs.hetzner.cloud/#label-selector
type LabelSelector struct {
	MustLabels    map[string]string
	MustNotLabels map[string]string
}

func (ls *LabelSelector) String() string {
	selectors := make([]string, 0, 5)

	for k, v := range ls.MustLabels {
		selectors = append(selectors, fmt.Sprint(k, "==", v))
	}
	for k, v := range ls.MustNotLabels {
		selectors = append(selectors, fmt.Sprint(k, "!=", v))
	}

	return strings.Join(selectors, ",")
}
