package v1alpha1

import (
	"fmt"
	"io/ioutil"

	"sigs.k8s.io/yaml"
)

func (pm *PackageTransformer) Marshal() ([]byte, error) {
	return yaml.Marshal(pm)
}

func FromYAML(bytes []byte) (*PackageTransformer, error) {
	pm := &PackageTransformer{}
	return pm, yaml.Unmarshal(bytes, pm)
}

func FromFile(path string) (*PackageTransformer, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}
	return FromYAML(bytes)
}
