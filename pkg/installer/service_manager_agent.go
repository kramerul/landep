package installer

import (
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

type ImagePullSecrets struct {
	Repository string `json:"repository"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

func ServiceManagerAgentInstallerFactory(target landep.Target, version *semver.Version) (landep.Installer, error) {
	cTarget, ok := target.(landep.K8sCloudFoundryBridgingTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &serviceManagerAgentInstaller{target: cTarget, version: version}, nil
}

func (s *serviceManagerAgentInstaller) Apply(name string, images map[string]landep.Image, parameter []landep.Parameter, helper *landep.InstallationHelper) (landep.Parameter, error) {
	var artifactory ImagePullSecrets
	return helper.SecretRequest("artifactory", "ARTIFACTORY", &artifactory).
		Apply(parameter, func(parameter landep.Parameter) (interface{}, error) {
			return &ServiceManagerAgentResponse{}, s.target.K8sTarget().Helm().Apply(name, "service-manager-agent", s.version, parameter)
		})
}

func (s *serviceManagerAgentInstaller) Delete(name string) error {
	return s.target.K8sTarget().Helm().Delete(name)
}
