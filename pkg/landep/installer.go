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
	Response     Parameter                      `json:"-"`
	Version      *semver.Version                `json:"version"`
	Requests     map[string]RequestedDependency `json:"-"`
	PkgName      string                         `json:"pkgName"`
	Target       Target                         `json:"-"`
	Digest       string                         `json:"-"`
	Dependencies *Dependencies                  `json:"-"`
}

func (s *Installation) IntersectedConstraints() IntersectedConstrains {
	intersectedConstraints := []*semver.Constraints{}
	for _, r := range s.Requests {
		intersectedConstraints = append(intersectedConstraints, r.Constraints)
	}
	return intersectedConstraints
}

// Needs to preserve insertion sequence
type Dependencies struct {
	installations       []*Installation          `json:"-"`
	installationsByName map[string]*Installation `json:",inline"`
}

func (s *Dependencies) Add(name string, installation *Installation) {
	_, existing := s.installationsByName[name]
	if existing {
		return
	}
	if s.installationsByName == nil {
		s.installationsByName = make(map[string]*Installation)
	}
	s.installationsByName[name] = installation
	s.installations = append(s.installations, installation)
}

func (s *Dependencies) Get(name string) (*Installation, bool) {
	i, ok := s.installationsByName[name]
	return i, ok
}

func (s *Dependencies) Installations() []*Installation {
	result := make([]*Installation, len(s.installations))
	copy(result, s.installations)
	return result
}

type RequestedDependency struct {
	PkgName     string              `json:"pkgName"`
	Constraints *semver.Constraints `json:"constraints"`
	Target      Target              `json:"target,omitempty"`
	Parameter   Parameter           `json:"parameter,omitempty"`
}

type Installer interface {
	Apply(name string, images map[string]Image, parameter []Parameter, dependencies *Dependencies) (Parameter, error)
	Delete(name string) error
}

type InstallerFactory = func(target Target, version *semver.Version) (Installer, error)
