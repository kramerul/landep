package installer

import (
	"encoding/json"
	"errors"

	"github.tools.sap/D001323/landep/pkg/landep"
)

type cloudFoundryEnvironmentInstaller struct {
	k8sTarget landep.K8sTarget
}

func CloudFoundryEnvironmentInstallerFactory(targets landep.Targets) (landep.Installer, error) {
	target, err := targets.SingleTarget()
	if err != nil {
		return nil, err
	}
	k8sTarget, ok := target.(landep.K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &cloudFoundryEnvironmentInstaller{k8sTarget: k8sTarget}, nil
}

func (s *cloudFoundryEnvironmentInstaller) Apply(name string, images map[string]landep.Image, parameter []landep.Parameter, dependencies *landep.Dependencies) (landep.Parameter, error) {
	dc := landep.NewDependencyChecker(dependencies)
	err := dc.WithRequired("cluster", "docker.io/pkgs/cluster", ">= 1.0", func(cluster *landep.Installation) error {
		config := &ClusterResponse{}
		err := json.Unmarshal(cluster.Response, config)
		if err != nil {
			return err
		}
		return dc.WithRequired("cloud-foundry", "docker.io/pkgs/cloud-foundry", ">= 2.0", func(cloudFoundry *landep.Installation) error {
			config := &CloudFoundryResponse{}
			err := json.Unmarshal(cloudFoundry.Response, config)
			if err != nil {
				return err
			}
			return dc.Required("organization", "docker.io/pkgs/organization", ">= 1.0", landep.WithDefaultTarget(landep.NewCloudFoundryTarget(&config.CF)))

		}, landep.WithDefaultTarget(landep.NewK8sTarget("cf-system", s.k8sTarget.Config())))

	})
	return []byte("{}"), err

}

func (s *cloudFoundryEnvironmentInstaller) Delete(name string) error {
	return nil
}
