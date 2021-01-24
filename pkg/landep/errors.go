package landep

import (
	"strings"
)

type DependenciesMissing struct {
	DependencyRequests map[string]DependencyRequest
}

func (d DependenciesMissing) Error() string {
	var sb strings.Builder
	sb.WriteString("the following dependencies are missing")
	for k, v := range d.DependencyRequests {
		if v.Installation != nil {
			sb.WriteString(k)
			sb.WriteString(": ")
			sb.WriteString(v.Installation.PkgName)
			sb.WriteString(", ")
		}
		if v.Secret != nil {
			sb.WriteString(v.Secret.Name)
		}
	}
	return sb.String()
}

var _ error = (*DependenciesMissing)(nil)
