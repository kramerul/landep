package installer

import (
	"errors"

	"github.tools.sap/D001323/landep/pkg/landep"
)

type organizationInstaller struct {
	cfTarget landep.CloudFoundryTarget
}

func OrganizationInstallerFactory(targets landep.Targets) (landep.Installer, error) {
	target, err := targets.SingleTarget()
	if err != nil {
		return nil, err
	}
	cfTarget, ok := target.(landep.CloudFoundryTarget)
	if !ok {
		return nil, errors.New("Not a CloudFoundryTarget")
	}
	return &organizationInstaller{cfTarget: cfTarget}, nil
}

func (s *organizationInstaller) Apply(name string, images map[string]landep.Image, parameter []landep.Parameter, dependencies *landep.Dependencies) (landep.Parameter, error) {
	err := s.cfTarget.CreateOrg(name, "admin")
	if err != nil {
		return nil, err
	}
	return []byte("{}"), nil

}

func (s *organizationInstaller) Delete(name string) error {
	return s.cfTarget.DeleteOrg(name)
}
