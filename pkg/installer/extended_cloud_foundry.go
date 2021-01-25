package installer

import (
	"errors"

	"github.com/Masterminds/semver/v3"

	"github.tools.sap/D001323/landep/pkg/landep"
)

type extendedCloudFoundryEnvironmentInstaller struct {
	k8sTarget landep.K8sTarget
	version   *semver.Version
}

func init() {
	landep.Repository.Register("docker.io/pkgs/extended-cloud-foundry", semver.MustParse("2.0.0"), extendedCloudFoundryEnvironmentInstallerFactory)
}

func extendedCloudFoundryEnvironmentInstallerFactory(target landep.Target, version *semver.Version) (landep.Installer, error) {
	k8sTarget, ok := target.(landep.K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &extendedCloudFoundryEnvironmentInstaller{k8sTarget: k8sTarget, version: version}, nil
}

func (s *extendedCloudFoundryEnvironmentInstaller) Apply(name string, images map[string]landep.Image, helper *landep.InstallationHelper) (landep.Parameter, error) {
	cloudFoundryResponse := CloudFoundryResponse{}
	var parameter landep.Parameter
	dummy := struct{}{}
	helper.
		MergedJsonParameter(&parameter).
		InstallationRequestCb(&cloudFoundryResponse, "cloud-foundry", "docker.io/pkgs/cloud-foundry", ">= 2.0", func() error {
			return helper.
				InstallationRequest(&dummy, "organization", "docker.io/pkgs/organization", ">= 1.0",
					landep.WithTarget(landep.NewCloudFoundryTarget(&cloudFoundryResponse))).
				InstallationRequest(&dummy, "service-manager-agent", "docker.io/pkgs/service-manager-agent", ">= 0.1",
					landep.WithTarget(landep.NewK8sCloudFoundryBridgingTarget(
						landep.NewK8sTarget("service-agent-manager", s.k8sTarget.Config()),
						landep.NewCloudFoundryTarget(&cloudFoundryResponse)))).
				Error()

		})

	return helper.Apply(func() (interface{}, error) {
		return &cloudFoundryResponse, nil
	})
}

func (s *extendedCloudFoundryEnvironmentInstaller) Delete(name string) error {
	return nil
}
