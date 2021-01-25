package installer

import (
	"errors"

	"github.com/Masterminds/semver/v3"
	"github.tools.sap/D001323/landep/pkg/landep"
)

type kymaInstaller struct {
	k8sTarget landep.K8sTarget
	version   *semver.Version
}
type KymaResponse struct {
}

func KymaInstallerFactory(target landep.Target, version *semver.Version) (landep.Installer, error) {
	k8sTarget, ok := target.(landep.K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &kymaInstaller{k8sTarget: k8sTarget, version: version}, nil
}

func (s *kymaInstaller) Apply(name string, images map[string]landep.Image, helper *landep.InstallationHelper) (landep.Parameter, error) {
	var params landep.Parameter
	var istioResponse IstioResponse
	return helper.
		InstallationRequest(&istioResponse, "istio", "docker.io/pkgs/istio", "~ 1.7",
			landep.WithTarget(landep.NewK8sTarget("istio-system", s.k8sTarget.Config())),
			landep.WithJsonParameter(&IstioParameter{Pilot: Pilot{Instances: 3}}),
		).
		ApplyJson(&params, func() (interface{}, error) {
			return &KymaResponse{}, s.k8sTarget.Helm().Apply(name, "kyma", s.version, params)
		})
}

func (s *kymaInstaller) Delete(name string) error {
	return s.k8sTarget.Helm().Delete(name)
}
