package landep

import (
	"encoding/json"

	"github.com/Masterminds/semver/v3"
)

type BasicAuthorization struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Credentials struct {
	URL   string             `json:"url"`
	Basic BasicAuthorization `json:"basic"`
}

type Target interface {
	Digest() []byte
}

type Helm interface {
	Apply(name string, chart string, version *semver.Version, parameter json.RawMessage) error
	Delete(name string) error
}

type Kapp interface {
	Apply(name string, chart string, version *semver.Version, parameter json.RawMessage) error
	Delete(name string) error
}

type K8sConfig struct {
	URL string `json "url"`
}
type K8sTarget interface {
	Target
	Helm() Helm
	Kapp() Kapp
	Config() *K8sConfig
}

type CloudFoundryConfig struct {
	CloudFoundryCredentials Credentials `json "cf"`
	UAACredentials          Credentials `json "uaa"`
}

type CloudFoundryTarget interface {
	Target
	CreateOrg(name string, user string) error
	DeleteOrg(name string) error
	Config() *CloudFoundryConfig
}

type K8sCloudFoundryBridgingTarget interface {
	Target
	K8sTarget() K8sTarget
	CloudFoundryTarget() CloudFoundryTarget
}

type targetFactory interface {
	K8s(namespace string, config *K8sConfig) K8sTarget
	CloudFoundry(config *CloudFoundryConfig) CloudFoundryTarget
	K8sCloudFoundryBridgingTarget(k8s K8sTarget, cf CloudFoundryTarget) K8sCloudFoundryBridgingTarget
}

var tf targetFactory

func NewK8sTarget(namespace string, config *K8sConfig) K8sTarget {
	return tf.K8s(namespace, config)
}

func NewCloudFoundryTarget(credentials *CloudFoundryConfig) CloudFoundryTarget {
	return tf.CloudFoundry(credentials)
}

func NewK8sCloudFoundryBridgingTarget(k8s K8sTarget, cf CloudFoundryTarget) K8sCloudFoundryBridgingTarget {
	return tf.K8sCloudFoundryBridgingTarget(k8s, cf)
}
