package landep

import (
	"encoding/json"

	"github.com/Masterminds/semver/v3"
)

type InstallationHelper struct {
	requestedDependencies map[string]DependencyRequest
	responses             map[string]Response
	err                   error
}

func NewDependencyChecker(responses map[string]Response) *InstallationHelper {
	return &InstallationHelper{requestedDependencies: make(map[string]DependencyRequest), responses: responses}
}

type InstallationOption = func(dep *InstallationRequest) error

func WithTarget(target Target) InstallationOption {
	return func(dep *InstallationRequest) error {
		dep.Target = target
		return nil
	}
}

func WithParameter(parameter Parameter) InstallationOption {
	return func(dep *InstallationRequest) error {
		dep.Parameter = parameter
		return nil
	}
}

func WithJsonParameter(parameter interface{}) InstallationOption {
	return func(dep *InstallationRequest) (err error) {
		dep.Parameter, err = json.Marshal(parameter)
		return
	}
}

func (s *InstallationHelper) WithInstallationRequest(name string, pkgName string, constraints string, cb func(response Parameter) error, options ...InstallationOption) error {
	if s.err != nil {
		return s.err
	}
	response, ok := s.responses[name]
	if !ok {
		c, err := semver.NewConstraint(constraints)
		if err != nil {
			s.err = err
			return s.Error()
		}
		installationRequest := InstallationRequest{PkgName: pkgName, Constraints: c}
		for _, o := range options {
			err := o(&installationRequest)
			if err != nil {
				return err
			}
		}
		s.requestedDependencies[name] = DependencyRequest{Installation: &installationRequest}
		return s.Error()
	}
	s.err = cb(response)
	return s.Error()
}

func (s *InstallationHelper) InstallationRequest(name string, pkgName string, constraints string, options ...InstallationOption) *InstallationHelper {
	s.WithInstallationRequest(name, pkgName, constraints, func(response Parameter) error { return nil }, options...)
	return s
}

func (s *InstallationHelper) SecretRequest(name string, secretName string, secret interface{}) *InstallationHelper {
	response, ok := s.responses[name]
	if !ok {
		secretRequest := SecretRequest{Name: secretName}
		s.requestedDependencies[name] = DependencyRequest{Secret: &secretRequest}
		return s
	}
	s.err = json.Unmarshal(response, secret)
	return s
}

func (s *InstallationHelper) Error() error {
	if s.err != nil {
		return s.err
	}
	if len(s.requestedDependencies) != 0 {
		return &DependenciesMissing{DependencyRequests: s.requestedDependencies}
	}
	return nil
}

func (s *InstallationHelper) Apply(parameter []Parameter, cb func(Parameter) (interface{}, error), options ...JsonMergeOption) (Response, error) {
	if err := s.Error(); err != nil {
		if err != nil {
			return nil, err
		}
	}
	params, err := JsonMerge(parameter, options...)
	if err != nil {
		return nil, err
	}
	response, err := cb(params)
	if err != nil {
		return nil, err
	}
	return json.Marshal(response)
}
