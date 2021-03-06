package installer

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.tools.sap/D001323/landep/pkg/landep"
)

type istioInstaller struct {
	k8sTarget landep.K8sTarget
	version   *semver.Version
}

type IstioResponse struct {
}

type Pilot struct {
	Instances int `json:"instances"`
}

type IstioParameter struct {
	Pilot Pilot `json:"pilot"`
}

func init() {
	landep.Repository.Register("docker.io/pkgs/istio", semver.MustParse("1.7.0"), istioInstallerFactory)
}

func istioInstallerFactory(target landep.Target, version *semver.Version) (landep.Installer, error) {
	k8sTarget, ok := target.(landep.K8sTarget)
	if !ok {
		return nil, errors.New("Not a K8sTarget")
	}
	return &istioInstaller{k8sTarget: k8sTarget, version: version}, nil
}

func istioConflictSolver(path string, j1 json.RawMessage, j2 json.RawMessage) (json.RawMessage, error) {
	if path == ".pilot.instances" {
		return landep.MaximumConflictSolver(path, j1, j2)
	}
	return nil, fmt.Errorf("Incompatible jsons at %s: '%s' '%s'", path, string(j1), string(j2))
}

func (s *istioInstaller) Apply(name string, images map[string]landep.Image, helper *landep.InstallationHelper) (landep.Parameter, error) {
	var params landep.Parameter
	return helper.
		MergedJsonParameter(&params, landep.WithConflictSolver(istioConflictSolver)).
		Apply(func() (interface{}, error) {
			return &IstioResponse{}, s.k8sTarget.Helm().Apply(name, "istio", s.version, params)
		})
}

func (s *istioInstaller) Delete(name string) error {
	return s.k8sTarget.Helm().Delete(name)
}
