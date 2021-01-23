package landep

import (
	"fmt"
	"strings"
)

type DependenciesMissing struct {
	RequestedDependencies map[string]RequestedDependency
}

func (d DependenciesMissing) Error() string {
	var sb strings.Builder
	sb.WriteString("the following dependencies are missing")
	for k, v := range d.RequestedDependencies {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(v.PkgName)
		sb.WriteString(", ")
	}
	return sb.String()
}

type DependencyStillRequired struct {
	Installation *Installation
	UsedBy       string
}

func (d DependencyStillRequired) Error() string {
	return fmt.Sprintf("Installation %s is still required by %s", d.Installation, d.UsedBy)
}

var _ error = (*DependenciesMissing)(nil)
var _ error = (*DependencyStillRequired)(nil)
