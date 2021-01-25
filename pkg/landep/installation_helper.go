package landep

import (
	"encoding/json"

	"github.com/Masterminds/semver/v3"
)

type InstallationHelper struct {
	requestedDependencies map[string]DependencyRequest
	responses             map[string]Response
	parameter             []Parameter
	err                   error
}

func NewDependencyChecker(parameter []Parameter, responses map[string]Response) *InstallationHelper {
	return &InstallationHelper{requestedDependencies: make(map[string]DependencyRequest), responses: responses, parameter: parameter}
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

func (s *InstallationHelper) InstallationRequestCb(response interface{}, name string, pkgName string, constraints string, cb func() error, options ...InstallationOption) error {
	if s.err != nil {
		return s.err
	}
	jsonResponse, ok := s.responses[name]
	if !ok {
		c, err := semver.NewConstraint(constraints)
		if err != nil {
			s.err = err
			return s.err
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
	s.err = json.Unmarshal(jsonResponse, response)
	if s.err != nil {
		return s.err
	}
	s.err = cb()
	return s.Error()
}

func (s *InstallationHelper) InstallationRequest(response interface{}, name string, pkgName string, constraints string, options ...InstallationOption) *InstallationHelper {
	s.InstallationRequestCb(response, name, pkgName, constraints, func() error { return nil }, options...)
	return s
}

func (s *InstallationHelper) SecretRequest(secret interface{}, name string, secretName string) *InstallationHelper {
	if s.err != nil {
		return s
	}
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

func (s *InstallationHelper) MergedJsonParameter(parameter *Parameter, options ...JsonMergeOption) *InstallationHelper {
	if s.err != nil {
		return s
	}
	(*parameter), s.err = JsonMerge(s.parameter, options...)
	return s
}

func (s *InstallationHelper) MergedParameter(parameter interface{}, options ...JsonMergeOption) *InstallationHelper {
	if s.err != nil {
		return s
	}
	parameterJson, err := JsonMerge(s.parameter, options...)
	if err != nil {
		s.err = err
		return s
	}
	if parameterJson != nil {
		s.err = json.Unmarshal(parameterJson, parameter)
		if s.err != nil {
			return s
		}
	}
	return s
}

func (s *InstallationHelper) Apply(cb func() (interface{}, error)) (Response, error) {
	err := s.Error()
	if err != nil {
		return nil, err
	}
	response, err := cb()
	if err != nil {
		return nil, err
	}
	return json.Marshal(response)
}
