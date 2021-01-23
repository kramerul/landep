package landep

import (
	"encoding/json"
	"errors"
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

type Targets map[string]Target

func (s *Targets) SingleTarget() (Target, error) {
	switch len(*s) {
	case 0:
		return nil, errors.New("No target available")
	case 1:
		for _, v := range *s {
			return v, nil
		}
	default:
	}
	return nil, errors.New("Too may targets")
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
	URL string
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

type targetFactory interface {
	K8s(namespace string, config *K8sConfig) K8sTarget
	CloudFoundry(credentials *Credentials) CloudFoundryTarget
}

var tf targetFactory

func NewK8sTarget(namespace string, config *K8sConfig) K8sTarget {
	return tf.K8s(namespace, config)
}

func NewCloudFoundryTarget(credentials *Credentials) CloudFoundryTarget {
	return tf.CloudFoundry(credentials)
}
