package landep

import (
	"encoding/json"
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
	Apply(name string, chart string, parameter json.RawMessage) error
	Delete(name string) error
}

type Kapp interface {
	Apply(name string, chart string, parameter json.RawMessage) error
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

type CloudFoundryTarget interface {
	Target
	CreateOrg(name string, user string) error
	DeleteOrg(name string) error
}

type K8sCloudFoundryBridgingTarget interface {
	Target
	K8sTarget() K8sTarget
	CloudFoundryTarget() CloudFoundryTarget
}

type targetFactory interface {
	K8s(namespace string, config *K8sConfig) K8sTarget
	CloudFoundry(credentials *Credentials) CloudFoundryTarget
	K8sCloudFoundryBridgingTarget(namespace string, config *K8sConfig, credentials *Credentials) K8sCloudFoundryBridgingTarget
}

var tf targetFactory

func NewK8sTarget(namespace string, config *K8sConfig) K8sTarget {
	return tf.K8s(namespace, config)
}

func NewCloudFoundryTarget(credentials *Credentials) CloudFoundryTarget {
	return tf.CloudFoundry(credentials)
}

func NewK8sCloudFoundryBridingTarget(namespace string, config *K8sConfig, credentials *Credentials) K8sCloudFoundryBridgingTarget {
	return tf.K8sCloudFoundryBridgingTarget(namespace, config, credentials)
}
