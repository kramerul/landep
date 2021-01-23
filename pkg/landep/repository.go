package landep

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
)

type versionedInstaller struct {
	version   *semver.Version
	installer InstallerFactory
}

type Repository map[string][]versionedInstaller

func (s *Repository) Register(name string, version *semver.Version, installer InstallerFactory) {
	if *s == nil {
		*s = make(map[string][]versionedInstaller)
	}
	installers := (*s)[name]
	installers = append(installers, versionedInstaller{version: version, installer: installer})
	sort.Slice(installers, func(i, j int) bool {
		return installers[i].version.GreaterThan(installers[j].version)
	})
	(*s)[name] = installers
}

func (s *Repository) Get(name string, contraints *semver.Constraints) (InstallerFactory, *semver.Version, error) {
	installers, ok := (*s)[name]
	if !ok {
		return nil, nil, fmt.Errorf("Installer for name %s not found", name)
	}
	var installersMatchingVersion []versionedInstaller
	for _, i := range installers {
		if contraints.Check(i.version) {
			installersMatchingVersion = append(installersMatchingVersion, i)
		}
	}
	if len(installersMatchingVersion) == 0 {
		return nil, nil, fmt.Errorf("Installer for name %s constraints %s not found", name, contraints.String())
	}
	// first one contains highest version
	return installersMatchingVersion[0].installer, installersMatchingVersion[0].version, nil
}
