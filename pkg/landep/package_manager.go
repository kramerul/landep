package landep

import (
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

func (s *PackageManager) installer(target Target, name string, constraints *semver.Constraints) (Installer, *semver.Version, error) {
	installerFactory, version, err := s.repository.Get(name, constraints)
	if err != nil {
		return nil, nil, err
	}
	installer, err := installerFactory(target)
	return installer, version, err
}

func qualifiedFinalizer(digest string) string {
	return "landep.kramerul.com/" + digest
}
func installationDigest(target Target, pkgName string) string {
	hash := md5.New()
	hash.Write(target.Digest())
	hash.Write([]byte(pkgName))
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *PackageManager) Apply(target Target, pkgName string, constraint *semver.Constraints, parameter Parameter) (*Installation, error) {
	return s.apply(target, pkgName, constraint, parameter, qualifiedFinalizer("package-manager"))
}

func cloneLabels(labels map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range labels {
		result[k] = v
	}
	return result
}
func (s *PackageManager) apply(target Target, pkgName string, constraint *semver.Constraints, parameter Parameter, finalizer string) (*Installation, error) {
	digest := installationDigest(target, pkgName)
	installation, ok := s.installationsByDigest[digest]
	if ok {
		installation.Finalizers[finalizer] = true
		return installation, nil
	}
	var response Parameter
	installer, version, err := s.installer(target, pkgName, constraint)
	if err != nil {
		return nil, err
	}
	subFinalizer := qualifiedFinalizer(digest)
	dependencies := &Dependencies{}
	for {
		response, err = installer.Apply(digest, nil, []Parameter{parameter}, dependencies)
		if err != nil {
			dependenciesMissing, ok := err.(*DependenciesMissing)
			if ok {
				for k, v := range dependenciesMissing.RequestedDependencies {
					if v.Target == nil {
						v.Target = target
					}
					depInstallation, err := s.apply(v.Target, v.PkgName, v.Constraints, v.Parameter, subFinalizer)
					if err != nil {
						return nil, err
					}
					dependencies.Add(k, depInstallation)
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
	installation = &Installation{Version: *version, Response: response, Finalizers: map[string]bool{finalizer: true}, PkgName: pkgName, Target: target, Digest: digest, Dependencies: dependencies}
	s.installationsByDigest[digest] = installation

	return installation, nil
}

func (s *PackageManager) Delete(target Target, pkgName string) error {
	digest := installationDigest(target, pkgName)
	installation, ok := s.installationsByDigest[digest]
	if !ok {
		return fmt.Errorf("Installation %s not found in target %v", pkgName, target)
	}
	return s.delete(installation, qualifiedFinalizer("package-manager"))
}

func (s *PackageManager) delete(installation *Installation, finalizer string) error {
	delete(installation.Finalizers, finalizer)
	if len(installation.Finalizers) != 0 {
		return nil
	}
	constraint, err := semver.NewConstraint("=" + installation.Version.String())
	if err != nil {
		return err
	}
	installer, _, err := s.installer(installation.Target, installation.PkgName, constraint)
	if err != nil {
		return err
	}
	err = installer.Delete(installation.Digest)
	if err != nil {
		return err
	}
	subFinalizer := qualifiedFinalizer(installation.Digest)
	dependencies := installation.Dependencies.Installations()
	for i := len(dependencies) - 1; i >= 0; i-- {
		err = s.delete(dependencies[i], subFinalizer)
		if err != nil {
			return err
		}
	}
	delete(s.installationsByDigest, installation.Digest)
	return nil
}
