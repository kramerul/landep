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

func init() {
	landep.Repository.Register("docker.io/pkgs/cloud-foundry-environment", semver.MustParse("1.0.0"), cloudFoundryEnvironmentInstallerFactory)
}

func cloudFoundryEnvironmentInstallerFactory(target landep.Target, version *semver.Version) (landep.Installer, error) {
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
	helper.
		MergedJsonParameter(&parameter).
		InstallationRequestCb(&clusterResponse, "cluster", "docker.io/pkgs/cluster", ">= 1.0", func() error {
			return helper.
				InstallationRequest(&cloudFoundryResponse, "cloud-foundry", "docker.io/pkgs/extended-cloud-foundry", ">= 2.0",
					landep.WithTarget(landep.NewK8sTarget("cf-system", &clusterResponse))).
				Error()

		})
	return helper.Apply(func() (interface{}, error) {
		return cloudFoundryResponse, nil
	})
}

func (s *cloudFoundryEnvironmentInstaller) Delete(name string) error {
	return nil
}
