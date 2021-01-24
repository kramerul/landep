package installer

import (
	"encoding/json"
	"errors"

	"github.com/Masterminds/semver/v3"

	"github.tools.sap/D001323/landep/pkg/landep"
)

type cloudFoundryEnvironmentInstaller struct {
	k8sTarget landep.K8sTarget
	version   *semver.Version
}

func CloudFoundryEnvironmentInstallerFactory(target landep.Target, version *semver.Version) (landep.Installer, error) {
	k8sTarget, ok := target.(landep.K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &cloudFoundryEnvironmentInstaller{k8sTarget: k8sTarget, version: version}, nil
}

func (s *cloudFoundryEnvironmentInstaller) Apply(name string, images map[string]landep.Image, parameter []landep.Parameter, helper *landep.InstallationHelper) (landep.Parameter, error) {
	err := helper.WithInstallationRequest("cluster", "docker.io/pkgs/cluster", ">= 1.0", func(response landep.Response) error {
		shootK8sConfig := &ClusterResponse{}
		err := json.Unmarshal(response, shootK8sConfig)
		if err != nil {
			return err
		}
		return helper.WithInstallationRequest("cloud-foundry", "docker.io/pkgs/cloud-foundry", ">= 2.0", func(response landep.Response) error {
			config := &CloudFoundryResponse{}
			err := json.Unmarshal(response, config)
			if err != nil {
				return err
			}
			return helper.InstallationRequest("organization", "docker.io/pkgs/organization", ">= 1.0", landep.WithTarget(landep.NewCloudFoundryTarget(&config.CF))).
				InstallationRequest("service-manager-agent", "docker.io/pkgs/service-manager-agent", ">= 0.1", landep.WithTarget(landep.NewK8sCloudFoundryBridingTarget("service-agent-manager", shootK8sConfig, &config.CF))).
				Error()

		}, landep.WithTarget(landep.NewK8sTarget("cf-system", shootK8sConfig)))

	})
	return []byte("{}"), err

}

func (s *cloudFoundryEnvironmentInstaller) Delete(name string) error {
	return nil
}
