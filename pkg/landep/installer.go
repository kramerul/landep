package landep

import (
	"encoding/json"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type Image struct {
	Repo string
	SHA  string
}

type Parameter = json.RawMessage

type Response = json.RawMessage

type Secret = json.RawMessage

type IntersectedConstrains []*semver.Constraints

func (s IntersectedConstrains) Check(version *semver.Version) bool {
	for _, c := range s {
		if !c.Check(version) {
			return false
		}
	}
	return true
}

func (s IntersectedConstrains) String() string {
	sb := strings.Builder{}
	for i, c := range s {
		if i > 0 {
			sb.WriteString(" and ")
		}
		sb.WriteString(c.String())
	}
	return sb.String()
}

type Installation struct {
	Response  Parameter                      `json:"-"`
	Version   *semver.Version                `json:"version"`
	Requests  map[string]InstallationRequest `json:"-"`
	PkgName   string                         `json:"pkgName"`
	Target    Target                         `json:"-"`
	Digest    string                         `json:"-"`
	Children  []*Installation                `json:"-"`
	Responses map[string]Response            `json:"-"`
}

func (s *Installation) IntersectedConstraints() IntersectedConstrains {
	intersectedConstraints := []*semver.Constraints{}
	for _, r := range s.Requests {
		intersectedConstraints = append(intersectedConstraints, r.Constraints)
	}
	return intersectedConstraints
}

type InstallationRequest struct {
	PkgName     string              `json:"pkgName"`
	Constraints *semver.Constraints `json:"constraints"`
	Target      Target              `json:"target,omitempty"`
	Parameter   Parameter           `json:"parameter,omitempty"`
}

type SecretRequest struct {
	Name string `json:"name"`
}

type DependencyRequest struct {
	Installation *InstallationRequest `json:"installation,omitempty"`
	Secret       *SecretRequest       `json:"secret,omitempty"`
}

type Installer interface {
	Apply(name string, images map[string]Image, parameter []Parameter, helper *InstallationHelper) (Parameter, error)
	Delete(name string) error
}

type InstallerFactory = func(target Target, version *semver.Version) (Installer, error)
