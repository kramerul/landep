package installer

import (
	"errors"

	"github.com/Masterminds/semver/v3"
	"github.tools.sap/D001323/landep/pkg/landep"
)

type organizationInstaller struct {
	cfTarget landep.CloudFoundryTarget
	version  *semver.Version
}

type OrganizationParameter struct {
	Username string `json:"username"`
}

func OrganizationInstallerFactory(target landep.Target, version *semver.Version) (landep.Installer, error) {
	cfTarget, ok := target.(landep.CloudFoundryTarget)
	if !ok {
		return nil, errors.New("Not a CloudFoundryTarget")
	}
	return &organizationInstaller{cfTarget: cfTarget, version: version}, nil
}

func (s *organizationInstaller) Apply(name string, images map[string]landep.Image, helper *landep.InstallationHelper) (landep.Parameter, error) {
	orgParams := OrganizationParameter{Username: "admin"}
	return helper.
		MergedParameter(&orgParams).
		Apply(func() (interface{}, error) {
			err := s.cfTarget.CreateOrg(name, orgParams.Username)
			return &struct{}{}, err
		})
}

func (s *organizationInstaller) Delete(name string) error {
	return s.cfTarget.DeleteOrg(name)
}
