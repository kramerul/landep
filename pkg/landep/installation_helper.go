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

func (s *InstallationHelper) WithInstallationRequest(response interface{}, name string, pkgName string, constraints string, cb func() error, options ...InstallationOption) error {
	if s.err != nil {
		return s.err
	}
	jsonResponse, ok := s.responses[name]
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
	s.err = json.Unmarshal(jsonResponse, response)
	if s.err != nil {
		return s.err
	}
	s.err = cb()
	return s.Error()
}

func (s *InstallationHelper) InstallationRequest(response interface{}, name string, pkgName string, constraints string, options ...InstallationOption) *InstallationHelper {
	s.WithInstallationRequest(response, name, pkgName, constraints, func() error { return nil }, options...)
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

func (s *InstallationHelper) ApplyJson(params *Parameter, cb func() (interface{}, error), options ...JsonMergeOption) (Response, error) {
	if err := s.Error(); err != nil {
		if err != nil {
			return nil, err
		}
	}
	var err error
	(*params), err = JsonMerge(s.parameter, options...)
	if err != nil {
		return nil, err
	}
	response, err := cb()
	if err != nil {
		return nil, err
	}
	return json.Marshal(response)
}

func (s *InstallationHelper) Apply(parameter interface{}, cb func() (interface{}, error), options ...JsonMergeOption) (Response, error) {
	var params Parameter
	return s.ApplyJson(&params, func() (interface{}, error) {
		if params != nil {
			err := json.Unmarshal(params, parameter)
			if err != nil {
				return nil, err
			}
		}
		return cb()
	}, options...)
}

func (s *InstallationHelper) ApplyVoid(cb func() (interface{}, error), options ...JsonMergeOption) (Response, error) {
	var params Parameter
	return s.ApplyJson(&params, func() (interface{}, error) {
		return cb()
	}, options...)
}
