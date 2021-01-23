package landep

import (
	"encoding/json"
	"errors"
)

type cloudFoundryEnvironmentInstaller struct {
	k8sTarget K8sTarget
}

func CloudFoundryEnvironmentInstallerFactory(targets Targets) (Installer, error) {
	target, err := targets.SingleTarget()
	if err != nil {
		return nil, err
	}
	k8sTarget, ok := target.(K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &cloudFoundryEnvironmentInstaller{k8sTarget: k8sTarget}, nil
}

func (s *cloudFoundryEnvironmentInstaller) Apply(name string, images map[string]Image, parameter []Parameter, dependencies InstalledDependencies) (Parameter, error) {
	dc := NewDependencyChecker(dependencies)
	err := dc.WithRequired("cluster", "docker.io/pkgs/cluster", ">= 1.0", func(cluster *Installation) error {
		config := &ClusterResponse{}
		err := json.Unmarshal(cluster.Response, config)
		if err != nil {
			return err
		}
		return dc.WithRequired("cloud-foundry", "docker.io/pkgs/cloud-foundry", ">= 2.0", func(cloudFoundry *Installation) error {
			config := &ClusterResponse{}
			err := json.Unmarshal(cloudFoundry.Response, config)
			if err != nil {
				return err
			}
			return dc.Required("organization", "docker.io/pkgs/organization", ">= 1.0", WithDefaultTarget(s.k8sTarget.For("cf-system", config.URL)))

		}, WithDefaultTarget(s.k8sTarget.For("cf-system", config.URL)))

	})
	return []byte("{}"), err

}

func (s *cloudFoundryEnvironmentInstaller) Delete(name string) error {
	return s.k8sTarget.Delete("cluster", name)
}
