package landep

import (
	"errors"
)

type organizationInstaller struct {
	k8sTarget K8sTarget
}

func OrganizationInstallerFactory(targets Targets) (Installer, error) {
	target, err := targets.SingleTarget()
	if err != nil {
		return nil, err
	}
	k8sTarget, ok := target.(K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &organizationInstaller{k8sTarget: k8sTarget}, nil
}

func (s *organizationInstaller) Apply(name string, images map[string]Image, parameter []Parameter, dependencies InstalledDependencies) (Parameter, error) {
	err := s.k8sTarget.CreateOrUpdate("organization", name, []byte("{}"))
	if err != nil {
		return nil, err
	}
	return []byte("{}"), nil

}

func (s *organizationInstaller) Delete(name string) error {
	return s.k8sTarget.Delete("organization", name)
}
