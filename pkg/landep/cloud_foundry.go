package landep

import (
	"encoding/json"
	"errors"
)

type cloudFoundryInstaller struct {
	k8sTarget K8sTarget
}
type BasicAuthorization struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Credentials struct {
	URL   string             `json:"url"`
	Basic BasicAuthorization `json:"basic"`
}
type CloudFoundryResponse struct {
	CF  Credentials `json:"cf"`
	UAA Credentials `json:"uaa"`
}

func CloudFoundryInstallerFactory(targets Targets) (Installer, error) {
	target, err := targets.SingleTarget()
	if err != nil {
		return nil, err
	}
	k8sTarget, ok := target.(K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &cloudFoundryInstaller{k8sTarget: k8sTarget}, nil
}

func (s *cloudFoundryInstaller) Apply(name string, images map[string]Image, parameter []Parameter, dependencies InstalledDependencies) (Parameter, error) {
	err := s.k8sTarget.CreateOrUpdate("secret", name, []byte("{}"))
	if err != nil {
		return nil, err
	}
	return json.Marshal(&CloudFoundryResponse{
		CF: Credentials{
			URL: "https://api.exapmle.com",
			Basic: BasicAuthorization{
				Username: "username",
				Password: "password",
			},
		},
		UAA: Credentials{
			URL: "https://uaa.exapmle.com",
			Basic: BasicAuthorization{
				Username: "username",
				Password: "password",
			},
		},
	})

}

func (s *cloudFoundryInstaller) Delete(name string) error {
	return s.k8sTarget.Delete("secret", name)
}
