package installer

import (
	"encoding/json"
	"errors"

	"github.tools.sap/D001323/landep/pkg/landep"
)

type serviceManagerAgentInstaller struct {
	target landep.K8sCloudFoundryBridgingTarget
}

type ServiceManagerAgentResponse struct {
}

func ServiceManagerAgentInstallerFactory(target landep.Target) (landep.Installer, error) {
	cTarget, ok := target.(landep.K8sCloudFoundryBridgingTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &serviceManagerAgentInstaller{target: cTarget}, nil
}

func (s *serviceManagerAgentInstaller) Apply(name string, images map[string]landep.Image, parameter []landep.Parameter, dependencies *landep.Dependencies) (landep.Parameter, error) {
	err := s.target.K8sTarget().Helm().Apply(name, "service-manager-agent", nil)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&ServiceManagerAgentResponse{})
}

func (s *serviceManagerAgentInstaller) Delete(name string) error {
	return s.target.K8sTarget().Helm().Delete(name)
}
