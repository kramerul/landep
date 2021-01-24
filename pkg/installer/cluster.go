package installer

import (
	"encoding/json"
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

func (s *clusterInstaller) Apply(name string, images map[string]landep.Image, parameter []landep.Parameter, dependencies *landep.Dependencies) (landep.Parameter, error) {
	params, err := landep.JsonMerge(parameter)
	if err != nil {
		return nil, err
	}
	err = s.k8sTarget.Helm().Apply(name, "cluster", s.version, params)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&ClusterResponse{
		URL: fmt.Sprintf("https://%s.cluster.hana-ondemand.com", name),
	})

}

func (s *clusterInstaller) Delete(name string) error {
	return s.k8sTarget.Helm().Delete(name)
}
