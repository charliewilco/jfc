package jfc

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

func formatYAML(input []byte, config Config) ([]byte, error) {
	if !utf8.Valid(input) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}

	decoder := yaml.NewDecoder(bytes.NewReader(input))
	var documents []yaml.Node
	for {
		var document yaml.Node
		err := decoder.Decode(&document)
		if err == nil {
			documents = append(documents, document)
			continue
		}
		if err == io.EOF {
			break
		}
		return nil, err
	}
	if len(documents) == 0 {
		documents = append(documents, yaml.Node{})
	}

	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(config.TabWidth)
	for i := range documents {
		if err := encoder.Encode(&documents[i]); err != nil {
			_ = encoder.Close()
			return nil, err
		}
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return applyOutputConventions(output.String(), config), nil
}
