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

type ServiceManagerAgentParams struct {
	ServiceManagerCredentials landep.Credentials `json:"smCredentials"`
}

type ServiceManagerAgentResponse struct {
}

type ImagePullSecrets struct {
	Repository string `json:"repository"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

func init() {
	landep.Repository.Register("docker.io/pkgs/service-manager-agent", semver.MustParse("0.1.0"), serviceManagerAgentInstallerFactory)
}

func serviceManagerAgentInstallerFactory(target landep.Target, version *semver.Version) (landep.Installer, error) {
	cTarget, ok := target.(landep.K8sCloudFoundryBridgingTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &serviceManagerAgentInstaller{target: cTarget, version: version}, nil
}

func (s *serviceManagerAgentInstaller) Apply(name string, images map[string]landep.Image, helper *landep.InstallationHelper) (landep.Parameter, error) {
	var artifactory ImagePullSecrets
	var params ServiceManagerAgentParams

	return helper.
		MergedParameter(&params).
		SecretRequest(&artifactory, "artifactory", "ARTIFACTORY").
		Apply(func() (interface{}, error) {
			params := map[string]interface{}{
				"SM_USER":                params.ServiceManagerCredentials.Basic.Username,
				"SM_PASSWORD":            params.ServiceManagerCredentials.Basic.Password,
				"CF_CLIENT_USERNAME":     s.target.CloudFoundryTarget().Config().CloudFoundryCredentials.Basic.Username,
				"CF_CLIENT_PASSWORD":     s.target.CloudFoundryTarget().Config().CloudFoundryCredentials.Basic.Password,
				"AUTHZ_CLIENT_ID":        s.target.CloudFoundryTarget().Config().UAACredentials.Basic.Username,
				"AUTHZ_CLIENT_SECRET":    s.target.CloudFoundryTarget().Config().UAACredentials.Basic.Password,
				"AUTHZ_CLIENT_ID_SUFFIX": "",
			}
			jsonParams, err := json.Marshal(params)
			if err != nil {
				return nil, err
			}
			return &ServiceManagerAgentResponse{}, s.target.K8sTarget().Helm().Apply(name, "service-manager-agent", s.version, jsonParams)
		})
}

func (s *serviceManagerAgentInstaller) Delete(name string) error {
	return s.target.K8sTarget().Helm().Delete(name)
}
