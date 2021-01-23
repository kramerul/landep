package landep

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

type K8sFake struct {
	objects   map[string]map[string]json.RawMessage
	namespace string
	url       string
}

func (s K8sFake) KubeConfig() (string, error) {
	return "KUBECONFIG", nil
}

func (s *K8sFake) CreateOrUpdate(kind string, name string, object json.RawMessage) error {
	if s.objects == nil {
		s.objects = make(map[string]map[string]json.RawMessage)
	}
	kinds, ok := s.objects[kind]
	if !ok {
		kinds = make(map[string]json.RawMessage)
		s.objects[kind] = kinds
	}
	kinds[name] = object
	return nil
}

func (s *K8sFake) Delete(kind string, name string) error {
	kinds, ok := s.objects[kind]
	if !ok {
		return fmt.Errorf("object kind %s not found ", kind)
	}
	delete(kinds, name)
	return nil
}

func (s *K8sFake) Digest() []byte {
	hash := md5.New()
	hash.Write([]byte(s.namespace))
	hash.Write([]byte(s.url))
	return hash.Sum(nil)
}

func (s *K8sFake) List(kind string) (map[string]json.RawMessage, error) {
	kinds, ok := s.objects[kind]
	if !ok {
		return nil, fmt.Errorf("object kind %s not found ", kind)
	}
	return kinds, nil
}

func (s *K8sFake) For(namespace string, url string) K8sTarget {
	return &K8sFake{namespace: namespace, url: url}
}

func NewK8sFake(namespace string, url string) *K8sFake {
	return &K8sFake{namespace: namespace, url: url}
}
