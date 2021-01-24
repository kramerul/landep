package installer

import (
	"encoding/json"
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

func (s *organizationInstaller) Apply(name string, images map[string]landep.Image, parameter []landep.Parameter, helper *landep.InstallationHelper) (landep.Parameter, error) {
	params, err := landep.JsonMerge(parameter)
	if err != nil {
		return nil, err
	}
	orgParams := &OrganizationParameter{Username: "admin"}
	if params != nil {
		err = json.Unmarshal(params, orgParams)
		if err != nil {
			return nil, err
		}
	}
	err = s.cfTarget.CreateOrg(name, orgParams.Username)
	if err != nil {
		return nil, err
	}
	return []byte("{}"), nil

}

func (s *organizationInstaller) Delete(name string) error {
	return s.cfTarget.DeleteOrg(name)
}
