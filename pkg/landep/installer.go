package landep

import (
	"encoding/json"

	"github.com/Masterminds/semver/v3"
)

type Image struct {
	Repo string
	SHA  string
}

type Parameter = json.RawMessage

type Installation struct {
	Response     Parameter
	Version      semver.Version
	Finalizers   map[string]bool
	PkgName      string
	Target       Target
	Digest       string
	Dependencies *Dependencies
}

// Needs to preserve insertion sequence
type Dependencies struct {
	installations       []*Installation
	installationsByName map[string]*Installation
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
	PkgName     string
	Constraints *semver.Constraints
	Target      Target
	Parameter   Parameter
}

type Installer interface {
	Apply(name string, images map[string]Image, parameter []Parameter, dependencies *Dependencies) (Parameter, error)
	Delete(name string) error
}

type InstallerFactory = func(target Target) (Installer, error)
