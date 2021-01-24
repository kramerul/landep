package landep

import (
	"crypto/md5"
	"encoding/json"
	"fmt"

	"github.com/Masterminds/semver/v3"
)

func InitFakeTargetFactory(log func(message string)) {
	tf = &fakeTargetFactory{log: log}
}

type fakeTargetFactory struct {
	log func(message string)
}

func (s *fakeTargetFactory) K8sCloudFoundryBridgingTarget(namespace string, config *K8sConfig, credentials *Credentials) K8sCloudFoundryBridgingTarget {
	return &k8sCloudFoundryBridgingTargetFake{k8sTarget: &k8sTargetFake{namespace: namespace, config: config, log: s.log},
		cloudFoundryTarget: &cloudFoundryTargetFake{url: credentials.URL, log: s.log}}
}

func (s *fakeTargetFactory) K8s(namespace string, config *K8sConfig) K8sTarget {
	return &k8sTargetFake{namespace: namespace, config: config, log: s.log}
}

func (s *fakeTargetFactory) CloudFoundry(credentials *Credentials) CloudFoundryTarget {
	return &cloudFoundryTargetFake{url: credentials.URL, log: s.log}
}

type k8sTargetFake struct {
	namespace string
	config    *K8sConfig
	log       func(message string)
}

func (s *k8sTargetFake) Config() *K8sConfig {
	return s.config
}

type helmFake struct {
	log       func(message string)
	namespace string
}

func (s *helmFake) Apply(name string, chart string, version *semver.Version, parameter json.RawMessage) error {
	s.log(fmt.Sprintf("helm upgrade -i -n %s --version %s %s %s %s", s.namespace, version.String(), name, chart, string(parameter)))
	return nil
}

func (s *helmFake) Delete(name string) error {
	s.log(fmt.Sprintf("helm delete -n %s %s", s.namespace, name))
	return nil
}

type kappFake struct {
	log       func(message string)
	namespace string
}

func (s *kappFake) Apply(name string, chart string, version *semver.Version, parameter json.RawMessage) error {
	s.log(fmt.Sprintf("kapp deploy -n %s -a %s %s %s", s.namespace, name, chart, string(parameter)))
	return nil
}

func (s *kappFake) Delete(name string) error {
	s.log(fmt.Sprintf("kapp delete -n %s -a %s", s.namespace, name))
	return nil
}

func (s *k8sTargetFake) Helm() Helm {
	return &helmFake{log: s.log, namespace: s.namespace}
}

func (s *k8sTargetFake) Kapp() Kapp {
	return &kappFake{log: s.log, namespace: s.namespace}
}

func (s *k8sTargetFake) Digest() []byte {
	hash := md5.New()
	hash.Write([]byte(s.namespace))
	hash.Write([]byte(s.config.URL))
	return hash.Sum(nil)
}

type cloudFoundryTargetFake struct {
	url string
	log func(message string)
}

func (s *cloudFoundryTargetFake) Digest() []byte {
	hash := md5.New()
	hash.Write([]byte(s.url))
	return hash.Sum(nil)
}

func (s *cloudFoundryTargetFake) DeleteOrg(name string) error {
	s.log(fmt.Sprintf("cf delete org %s", name))
	return nil
}

func (s *cloudFoundryTargetFake) CreateOrg(name string, user string) error {
	s.log(fmt.Sprintf("cf create org %s", name))
	return nil
}

type k8sCloudFoundryBridgingTargetFake struct {
	k8sTarget          K8sTarget
	cloudFoundryTarget CloudFoundryTarget
}

func (s k8sCloudFoundryBridgingTargetFake) Digest() []byte {
	hash := md5.New()
	hash.Write(s.k8sTarget.Digest())
	hash.Write(s.cloudFoundryTarget.Digest())
	return hash.Sum(nil)
}

func (s k8sCloudFoundryBridgingTargetFake) K8sTarget() K8sTarget {
	return s.k8sTarget
}

func (s k8sCloudFoundryBridgingTargetFake) CloudFoundryTarget() CloudFoundryTarget {
	return s.cloudFoundryTarget
}
