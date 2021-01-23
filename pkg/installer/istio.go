package installer

import (
	"errors"

	"github.tools.sap/D001323/landep/pkg/landep"
)

type istioInstaller struct {
	k8sTarget landep.K8sTarget
}

type IstioResponse struct {
}

func IstioInstallerFactory(target landep.Target) (landep.Installer, error) {
	k8sTarget, ok := target.(landep.K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &istioInstaller{k8sTarget: k8sTarget}, nil
}

func (s *istioInstaller) Apply(name string, images map[string]landep.Image, parameter []landep.Parameter, dependencies *landep.Dependencies) (landep.Parameter, error) {
	err := s.k8sTarget.Helm().Apply(name, "istio", nil)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *istioInstaller) Delete(name string) error {
	return s.k8sTarget.Helm().Delete(name)
}
