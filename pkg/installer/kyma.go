package installer

import (
	"encoding/json"
	"errors"

	"github.tools.sap/D001323/landep/pkg/landep"
)

type kymaInstaller struct {
	k8sTarget landep.K8sTarget
}
type KymaResponse struct {
}

func KymaInstallerFactory(target landep.Target) (landep.Installer, error) {
	k8sTarget, ok := target.(landep.K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &kymaInstaller{k8sTarget: k8sTarget}, nil
}

func (s *kymaInstaller) Apply(name string, images map[string]landep.Image, parameter []landep.Parameter, dependencies *landep.Dependencies) (landep.Parameter, error) {
	dc := landep.NewDependencyChecker(dependencies)
	err := dc.Required("istio", "docker.io/pkgs/istio", ">= 1.0", landep.WithTarget(landep.NewK8sTarget("istio-system", s.k8sTarget.Config()))).Error()
	if err != nil {
		return nil, err
	}
	err = s.k8sTarget.Helm().Apply(name, "kyma", nil)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&KymaResponse{})

}

func (s *kymaInstaller) Delete(name string) error {
	return s.k8sTarget.Helm().Delete(name)
}
