package installer

import (
	"encoding/json"
	"errors"

	"github.com/Masterminds/semver/v3"
	"github.tools.sap/D001323/landep/pkg/landep"
)

type serviceManagerAgentInstaller struct {
	target  landep.K8sCloudFoundryBridgingTarget
	version *semver.Version
}

type ServiceManagerAgentResponse struct {
}

func ServiceManagerAgentInstallerFactory(target landep.Target, version *semver.Version) (landep.Installer, error) {
	cTarget, ok := target.(landep.K8sCloudFoundryBridgingTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &serviceManagerAgentInstaller{target: cTarget, version: version}, nil
}

func (s *serviceManagerAgentInstaller) Apply(name string, images map[string]landep.Image, parameter []landep.Parameter, dependencies *landep.Dependencies) (landep.Parameter, error) {
	params, err := landep.JsonMerge(parameter)
	if err != nil {
		return nil, err
	}
	err = s.target.K8sTarget().Helm().Apply(name, "service-manager-agent", s.version, params)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&ServiceManagerAgentResponse{})
}

func (s *serviceManagerAgentInstaller) Delete(name string) error {
	return s.target.K8sTarget().Helm().Delete(name)
}
