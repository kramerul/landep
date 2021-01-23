package landep

import "github.com/Masterminds/semver/v3"

type dependencyChecker struct {
	requestedDependencies map[string]RequestedDependency
	dependencies          *Dependencies
	err                   error
}

func NewDependencyChecker(dependencies *Dependencies) *dependencyChecker {
	return &dependencyChecker{requestedDependencies: make(map[string]RequestedDependency), dependencies: dependencies}
}

type DependencyOption = func(dep *RequestedDependency)

func WithTarget(target Target) DependencyOption {
	return func(dep *RequestedDependency) {
		dep.Target = target
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
	installation, ok := s.dependencies.Get(name)
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
	s.err = cb(installation)
	return s.Error()
}

func (s *dependencyChecker) Required(name string, pkgName string, constraints string, options ...DependencyOption) *dependencyChecker {
	s.WithRequired(name, pkgName, constraints, func(installation *Installation) error { return nil }, options...)
	return s
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
