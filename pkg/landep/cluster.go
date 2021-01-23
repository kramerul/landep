package landep

import (
	"encoding/json"
	"errors"
	"fmt"
)

type clusterInstaller struct {
	k8sTarget K8sTarget
}

type ClusterResponse struct {
	URL    string          `json:"url"`
	Config json.RawMessage `json:"config"`
}

func ClusterInstallerFactory(targets Targets) (Installer, error) {
	target, err := targets.SingleTarget()
	if err != nil {
		return nil, err
	}
	k8sTarget, ok := target.(K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &clusterInstaller{k8sTarget: k8sTarget}, nil
}

func (s *clusterInstaller) Apply(name string, images map[string]Image, parameter []Parameter, dependencies InstalledDependencies) (Parameter, error) {
	err := s.k8sTarget.CreateOrUpdate("cluster", name, []byte("{}"))
	if err != nil {
		return nil, err
	}
	return json.Marshal(&ClusterResponse{
		URL: fmt.Sprintf("https://%s.cluster.hana-ondemand.com", name),
	})

}

func (s *clusterInstaller) Delete(name string) error {
	return s.k8sTarget.Delete("cluster", name)
}
