package installer

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.tools.sap/D001323/landep/pkg/landep"
)

type clusterInstaller struct {
	k8sTarget landep.K8sTarget
	version   *semver.Version
}

type ClusterResponse = landep.K8sConfig

func ClusterInstallerFactory(target landep.Target, version *semver.Version) (landep.Installer, error) {
	k8sTarget, ok := target.(landep.K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &clusterInstaller{k8sTarget: k8sTarget, version: version}, nil
}

func (s *clusterInstaller) Apply(name string, images map[string]landep.Image, helper *landep.InstallationHelper) (landep.Parameter, error) {
	var params landep.Response
	return helper.
		MergedJsonParameter(&params).
		Apply(func() (interface{}, error) {
			return &ClusterResponse{
				URL: fmt.Sprintf("https://%s.cluster.hana-ondemand.com", name),
			}, s.k8sTarget.Helm().Apply(name, "cluster", s.version, params)

		})
}

func (s *clusterInstaller) Delete(name string) error {
	return s.k8sTarget.Helm().Delete(name)
}
