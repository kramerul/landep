package landep

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	semver "github.com/Masterminds/semver/v3"
)

type Target interface {
	Digest() []byte
}

type K8sTarget interface {
	Target
	KubeConfig() (string, error)
	// This is only for simplicity. Normally, you would use KubeConfig and then run helm
	CreateOrUpdate(kind string, name string, message json.RawMessage) error
	Delete(kind string, name string) error
	List(kind string) (map[string]json.RawMessage, error)
	For(namespace string, url string) K8sTarget
}
type Targets map[string]Target

func (s *Targets) SingleTarget() (Target, error) {
	switch len(*s) {
	case 0:
		return nil, errors.New("No target available")
	case 1:
		for _, v := range *s {
			return v, nil
		}
	default:
	}
	return nil, errors.New("Too may targets")
}



type Image struct {
	Repo string
	SHA  string
}

type Parameter = json.RawMessage

type Installation struct {
	Response Parameter
	Version  semver.Version
	Labels   map[string]string
	PkgName  string
	Targets  Targets
	Digest   string
}

type RequestedDependency struct {
	PkgName     string
	Constraints *semver.Constraints
	Targets     Targets
	Parameter   Parameter
}

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
	UsedBy string
}

func (d DependencyStillRequired) Error() string {
	return fmt.Sprintf("Installation %s is still required by %s",d.Installation,d.UsedBy)
}

var _ error = (*DependenciesMissing)(nil)
var _ error = (*DependencyStillRequired)(nil)

type dependencyChecker struct {
	requestedDependencies map[string]RequestedDependency
	installedDependencies InstalledDependencies
	err                   error
}

func NewDependencyChecker(dependencies InstalledDependencies) *dependencyChecker {
	return &dependencyChecker{requestedDependencies: make(map[string]RequestedDependency), installedDependencies: dependencies}
}

type DependencyOption = func(dep *RequestedDependency)

func WithDefaultTarget(target Target) DependencyOption {
	return WithTargets(Targets{"default": target})
}

func WithTargets(targets Targets) DependencyOption {
	return func(dep *RequestedDependency) {
		dep.Targets = targets
	}
}
func WithParameter(parameter Parameter) DependencyOption {
	return func(dep *RequestedDependency) {
		dep.Parameter = parameter
	}
}

func (s *dependencyChecker) WithRequired(name string, pkgName string, constraints string, cb func(installation *Installation) error, options ...DependencyOption) error {
	if s.err != nil {
		return s.err
	}
	installation, ok := s.installedDependencies[name]
	if !ok {
		c, err := semver.NewConstraint(constraints)
		if err != nil {
			s.err = err
			return s.Error()
		}
		requestedDependency := RequestedDependency{PkgName: pkgName, Constraints: c}
		for _, o := range options {
			o(&requestedDependency)
		}
		s.requestedDependencies[name] = requestedDependency
		return s.Error()
	}
	s.err = cb(&installation)
	return s.Error()
}

func (s *dependencyChecker) Required(name string, pkgName string, constraints string, options ...DependencyOption) error {
	return s.WithRequired(name, pkgName, constraints, func(installation *Installation) error { return nil }, options...)
}

func (s *dependencyChecker) Error() error {
	if s.err != nil {
		return s.err
	}
	if len(s.requestedDependencies) != 0 {
		return &DependenciesMissing{RequestedDependencies: s.requestedDependencies}
	}
	return nil
}

type InstalledDependencies map[string]Installation

type Installer interface {
	Apply(name string, images map[string]Image, parameter []Parameter, dependencies InstalledDependencies) (Parameter, error)
	Delete(name string) error
}

type InstallerFactory = func(targets Targets) (Installer, error)

type versionedInstaller struct {
	version   *semver.Version
	installer InstallerFactory
}

type InstallerRepository map[string][]versionedInstaller

func (s *InstallerRepository) Register(name string, version *semver.Version, installer InstallerFactory) {
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

func (s *InstallerRepository) Get(name string, contraints *semver.Constraints) (InstallerFactory, *semver.Version, error) {
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

type LabelIndex map[string]map[string]map[string]bool

func (s* LabelIndex) Add(label string, value string, id string) {
	if *s == nil {
		*s = make(map[string]map[string]map[string]bool)
	}
	values, ok  := (*s)[label]
	if !ok {
		values = make(map[string]map[string]bool)
		(*s)[label] = values
	}
	ids, ok := values[value]
	if !ok {
		ids = make(map[string]bool)
		values[value] = ids
	}
	ids[id] = true
}

func (s* LabelIndex) Remove(label string, value string, id string) {
	if *s == nil {
		*s = make(map[string]map[string]map[string]bool)
	}
	values, ok  := (*s)[label]
	if !ok {
		return
	}
	ids, ok := values[value]
	if !ok {
		return
	}
	delete(ids,id)
}

func (s* LabelIndex) Each(label string, value string, cb func(id string) error) error {
	if *s == nil {
		*s = make(map[string]map[string]map[string]bool)
	}
	values, ok  := (*s)[label]
	if !ok {
		return nil
	}
	ids, ok := values[value]
	if !ok {
		return nil
	}
	var keys []string
	for k := range ids {
		keys = append(keys,k)
	}
	for _, k := range keys {
		err := cb(k)
		if err != nil {
			return err
		}
	}
	return nil
}

type packageManager struct {
	repository    InstallerRepository
	installationsByDigest map[string]*Installation
	labelsIndex LabelIndex
}

func (s *packageManager) installer(targets Targets, name string, constraints *semver.Constraints) (Installer, *semver.Version, error) {
	installerFactory, version, err := s.repository.Get(name, constraints)
	if err != nil {
		return nil, nil, err
	}
	installer, err := installerFactory(targets)
	return installer, version, err
}

func installationDigest(targets Targets, pkgName string) string {
	hash := md5.New()
	for k, v := range targets {
		hash.Write([]byte(k))
		hash.Write(v.Digest())
	}
	hash.Write([]byte(pkgName))
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *packageManager) Apply(targets Targets, pkgName string, constraint *semver.Constraints, parameter Parameter, labels map[string]string) (*Installation, error) {
	return s.apply(targets,pkgName,constraint,parameter,labels)
}

func cloneLabels(labels map[string]string) map[string]string {
	result := make(map[string]string)
	for k,v := range labels {
		result[k] = v
	}
	return result
}
func (s *packageManager) apply(targets Targets, pkgName string, constraint *semver.Constraints, parameter Parameter, labels map[string]string) (*Installation, error) {
	digest := installationDigest(targets, pkgName)
	installation, ok := s.installationsByDigest[digest]
	if ok {
		return installation, nil
	}
	var response Parameter
	installer, version, err := s.installer(targets, pkgName, constraint)
	if err != nil {
		return nil, err
	}
	sublabels := cloneLabels(labels)
	sublabels["installer.com/used-by/" +digest] = "true"
	dependencies := make(map[string]Installation)
	for {
		response, err = installer.Apply(digest, nil, []Parameter{parameter}, dependencies)
		if err != nil {
			dependenciesMissing, ok := err.(*DependenciesMissing)
			if ok {
				for k, v := range dependenciesMissing.RequestedDependencies {
					if v.Targets == nil {
						v.Targets = targets
					}
					depInstallation, err := s.apply(v.Targets, v.PkgName, v.Constraints, v.Parameter,sublabels)
					if err != nil {
						return nil, err
					}
					dependencies[k] = *depInstallation
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
	installation = &Installation{Version: *version, Response: response, Labels: make(map[string]string), PkgName: pkgName, Targets: targets, Digest: digest}
	s.installationsByDigest[digest] = installation

	for k,v := range labels {
		s.labelsIndex.Add(k,v,digest)
	}
	return installation, nil
}

func (s *packageManager) EachInstallation(label string, value string, cb func(installation *Installation) error) error {
	return s.labelsIndex.Each(label,value, func(digest string) error {
		installation , ok := s.installationsByDigest[digest]
		if !ok {
			return fmt.Errorf("Missing installation %s",digest)
		}
		return cb(installation)
	})
}

func (s *packageManager) Delete(targets Targets, pkgName string,) error {
	digest := installationDigest(targets, pkgName)
	installation, ok := s.installationsByDigest[digest]
	if !ok {
		return fmt.Errorf("Installation %s not found in targets %v", pkgName, targets)
	}
	err := s.delete(installation)
	if err != nil {
		return err
	}
	return s.EachInstallation("installer.com/used-by/" + digest, "true", func(installation *Installation) error {
		s.removeLabel("installer.com/used-by/" + digest,"true",installation)
		err := s.delete(installation)
		if err != nil {
			_, ok :=  err.(*DependencyStillRequired)
			if !ok {
				return err
			}
		}
		return nil
	})
}

func (s *packageManager) delete(installation *Installation) error {
	for k,_ := range installation.Labels {
		if strings.HasPrefix(k,"installer.com/used-by/") {
			digest := strings.TrimPrefix(k,"installer.com/used-by/")
			userBy, ok := s.installationsByDigest[digest]
			if ok {
				return DependencyStillRequired{Installation: installation, UsedBy: fmt.Sprintf("%s on targets %v", userBy.PkgName, userBy.Targets)}
			} else {
				return DependencyStillRequired{Installation: installation, UsedBy: digest}
			}
		}
	}
	constraint, err := semver.NewConstraint("=" + installation.Version.String())
	if err != nil {
		return err
	}
	installer, _, err := s.installer(installation.Targets, installation.PkgName, constraint)
	if err != nil {
		return err
	}
	err = installer.Delete(installation.Digest)
	if err != nil {
		return err
	}
	for k,v := range installation.Labels {
		values := s.labelsIndex[k]
		if values != nil {
			delete(values,v)
		}
	}
	delete(s.installationsByDigest, installation.Digest)
	return nil
}


func (s *packageManager) removeLabel(label string, value string, installation *Installation)  {
	delete(installation.Labels,label)
	s.labelsIndex.Remove(label,value,installation.Digest)
}