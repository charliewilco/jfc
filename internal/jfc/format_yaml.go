package jfc

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

func formatYAML(input []byte, config Config) ([]byte, error) {
	if !utf8.Valid(input) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}

	var document yaml.Node
	if err := yaml.Unmarshal(input, &document); err != nil {
		return nil, err
	}

	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(config.TabWidth)
	if err := encoder.Encode(&document); err != nil {
		_ = encoder.Close()
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return applyOutputConventions(output.String(), config), nil
}
