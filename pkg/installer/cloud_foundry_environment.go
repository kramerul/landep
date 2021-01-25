package installer

import (
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

func (s *cloudFoundryEnvironmentInstaller) Apply(name string, images map[string]landep.Image, helper *landep.InstallationHelper) (landep.Parameter, error) {
	clusterResponse := ClusterResponse{}
	cloudFoundryResponse := CloudFoundryResponse{}
	var parameter landep.Parameter
	dummy := struct{}{}
	helper.
		MergedJsonParameter(&parameter).
		InstallationRequestCb(&clusterResponse, "cluster", "docker.io/pkgs/cluster", ">= 1.0", func() error {
			return helper.
				InstallationRequestCb(&cloudFoundryResponse, "cloud-foundry", "docker.io/pkgs/cloud-foundry", ">= 2.0", func() error {
					return helper.
						InstallationRequest(&dummy, "organization", "docker.io/pkgs/organization", ">= 1.0",
							landep.WithTarget(landep.NewCloudFoundryTarget(&cloudFoundryResponse.CF))).
						InstallationRequest(&dummy, "service-manager-agent", "docker.io/pkgs/service-manager-agent", ">= 0.1",
							landep.WithTarget(landep.NewK8sCloudFoundryBridingTarget("service-agent-manager", &clusterResponse, &cloudFoundryResponse.CF))).
						Error()

				}, landep.WithTarget(landep.NewK8sTarget("cf-system", &clusterResponse)))

		})
	return helper.Apply(func() (interface{}, error) {
		return &struct{}{}, nil
	})
}

func (s *cloudFoundryEnvironmentInstaller) Delete(name string) error {
	return nil
}
