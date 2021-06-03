package wyaml

import (
	"fmt"
	"gopkg.in/yaml.v2"

	"joao/wasm-manipulator/pkg/wfile"
)

// Read reads a yaml file and returns the content on a transformation model.
func Read(filename string) (*BaseYAML, error) {
	yamlContent, err := wfile.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading yaml input content: %w", err)
	}
	var data BaseYAML
	if err := yaml.Unmarshal([]byte(yamlContent), &data); err != nil {
		return nil, fmt.Errorf("unmarshal yaml input content: %w", err)
	}
	return &data, nil
}
