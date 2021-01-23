package installer

import (
	"encoding/json"
	"errors"

	"github.tools.sap/D001323/landep/pkg/landep"
)

type cloudFoundryInstaller struct {
	k8sTarget landep.K8sTarget
}
type CloudFoundryResponse struct {
	CF  landep.Credentials `json:"cf"`
	UAA landep.Credentials `json:"uaa"`
}

func CloudFoundryInstallerFactory(targets landep.Targets) (landep.Installer, error) {
	target, err := targets.SingleTarget()
	if err != nil {
		return nil, err
	}
	k8sTarget, ok := target.(landep.K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &cloudFoundryInstaller{k8sTarget: k8sTarget}, nil
}

func (s *cloudFoundryInstaller) Apply(name string, images map[string]landep.Image, parameter []landep.Parameter, dependencies *landep.Dependencies) (landep.Parameter, error) {
	dc := landep.NewDependencyChecker(dependencies)
	err := dc.Required("istio", "docker.io/pkgs/istio", ">= 1.0", landep.WithDefaultTarget(landep.NewK8sTarget("istio-system", s.k8sTarget.Config())))
	if err != nil {
		return nil, err
	}
	err = s.k8sTarget.Kapp().Apply(name, "cf-for-k8s-scp", nil)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&CloudFoundryResponse{
		CF: landep.Credentials{
			URL: "https://api.exapmle.com",
			Basic: landep.BasicAuthorization{
				Username: "username",
				Password: "password",
			},
		},
		UAA: landep.Credentials{
			URL: "https://uaa.exapmle.com",
			Basic: landep.BasicAuthorization{
				Username: "username",
				Password: "password",
			},
		},
	})

}

func (s *cloudFoundryInstaller) Delete(name string) error {
	return s.k8sTarget.Kapp().Delete(name)
}
