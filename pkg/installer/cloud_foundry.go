package installer

import (
	"errors"

	"github.com/Masterminds/semver/v3"

	"github.tools.sap/D001323/landep/pkg/landep"
)

type cloudFoundryInstaller struct {
	k8sTarget landep.K8sTarget
	version   *semver.Version
}
type CloudFoundryResponse struct {
	CF  landep.Credentials `json:"cf"`
	UAA landep.Credentials `json:"uaa"`
}

func CloudFoundryInstallerFactory(target landep.Target, version *semver.Version) (landep.Installer, error) {
	k8sTarget, ok := target.(landep.K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &cloudFoundryInstaller{k8sTarget: k8sTarget, version: version}, nil
}

func (s *cloudFoundryInstaller) Apply(name string, images map[string]landep.Image, helper *landep.InstallationHelper) (landep.Parameter, error) {
	var params landep.Parameter
	var istioResponse IstioResponse
	return helper.
		InstallationRequest(&istioResponse, "istio", "docker.io/pkgs/istio", ">= 1.6",
			landep.WithTarget(landep.NewK8sTarget("istio-system", s.k8sTarget.Config())),
			landep.WithJsonParameter(&IstioParameter{Pilot: Pilot{Instances: 1}})).
		ApplyJson(&params, func() (interface{}, error) {
			err := s.k8sTarget.Kapp().Apply(name, "cf-for-k8s-scp", s.version, params)
			if err != nil {
				return nil, err
			}
			return &CloudFoundryResponse{
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
			}, nil
		})

}

func (s *cloudFoundryInstaller) Delete(name string) error {
	return s.k8sTarget.Kapp().Delete(name)
}
