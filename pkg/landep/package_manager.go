package landep

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	semver "github.com/Masterminds/semver/v3"
)

type PackageManager struct {
	repository            Repository
	installationsByDigest map[string]*Installation
}

func NewPackageManager(repository Repository) *PackageManager {
	return &PackageManager{repository: repository}
}

func (s *PackageManager) installer(target Target, name string, constraints IntersectedConstrains) (Installer, *semver.Version, error) {
	installerFactory, version, err := s.repository.Get(name, constraints)
	if err != nil {
		return nil, nil, err
	}
	installer, err := installerFactory(target, version)
	return installer, version, err
}

func installationDigest(target Target, pkgName string) string {
	hash := md5.New()
	hash.Write(target.Digest())
	hash.Write([]byte(pkgName))
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *PackageManager) Apply(target Target, pkgName string, constraint *semver.Constraints, parameter Parameter) (*Installation, error) {
	return s.apply(target, pkgName, constraint, parameter, "package-manager")
}

func requesterName(pkgName string, digest string) string {
	return pkgName + "/" + digest
}

func (s *PackageManager) apply(target Target, pkgName string, constraints *semver.Constraints, parameter Parameter, requester string) (*Installation, error) {
	digest := installationDigest(target, pkgName)
	installation, ok := s.installationsByDigest[digest]
	requestedDependency := RequestedDependency{
		PkgName:     pkgName,
		Constraints: constraints,
		Target:      target,
		Parameter:   parameter,
	}

	if ok {
		request, ok := installation.Requests[requester]
		if ok {
			if bytes.Compare(request.Parameter, requestedDependency.Parameter) == 0 && request.Constraints == requestedDependency.Constraints {
				return installation, nil
			}
		}
		installation.Requests[requester] = requestedDependency
	} else {
		installation = &Installation{PkgName: pkgName, Target: target, Digest: digest, Requests: map[string]RequestedDependency{requester: requestedDependency}, Dependencies: &Dependencies{}}
	}
	installer, version, err := s.installer(target, pkgName, installation.IntersectedConstraints())
	installation.Version = version
	if err != nil {
		return nil, err
	}
	subRequester := requesterName(pkgName, digest)
	for {
		joinedParamater := []Parameter{}
		for _, r := range installation.Requests {
			if r.Parameter != nil {
				joinedParamater = append(joinedParamater, r.Parameter)
			}
		}
		installation.Response, err = installer.Apply(digest, nil, joinedParamater, installation.Dependencies)
		if err != nil {
			dependenciesMissing, ok := err.(*DependenciesMissing)
			if ok {
				for k, v := range dependenciesMissing.RequestedDependencies {
					if v.Target == nil {
						v.Target = target
					}
					depInstallation, err := s.apply(v.Target, v.PkgName, v.Constraints, v.Parameter, subRequester)
					if err != nil {
						return nil, err
					}
					installation.Dependencies.Add(k, depInstallation)
				}
			} else {
				return nil, err
			}
		} else {
			break
		}
	}
	if s.installationsByDigest == nil {
		s.installationsByDigest = make(map[string]*Installation)
	}
	s.installationsByDigest[digest] = installation

	return installation, nil
}

func (s *PackageManager) Delete(target Target, pkgName string) error {
	digest := installationDigest(target, pkgName)
	installation, ok := s.installationsByDigest[digest]
	if !ok {
		return fmt.Errorf("Installation %s not found in target %v", pkgName, target)
	}
	return s.delete(installation, "package-manager")
}

func (s *PackageManager) delete(installation *Installation, requester string) error {
	delete(installation.Requests, requester)
	if len(installation.Requests) != 0 {
		return nil
	}
	installer, _, err := s.installer(installation.Target, installation.PkgName, installation.IntersectedConstraints())
	if err != nil {
		return err
	}
	err = installer.Delete(installation.Digest)
	if err != nil {
		return err
	}
	subRequester := requesterName(installation.PkgName, installation.Digest)
	dependencies := installation.Dependencies.Installations()
	for i := len(dependencies) - 1; i >= 0; i-- {
		err = s.delete(dependencies[i], subRequester)
		if err != nil {
			return err
		}
	}
	delete(s.installationsByDigest, installation.Digest)
	return nil
}
