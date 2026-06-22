package jfc

import (
	"fmt"
	"unicode/utf8"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func formatHCL(input []byte, config Config) ([]byte, error) {
	if !utf8.Valid(input) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}
	if _, diagnostics := hclwrite.ParseConfig(input, "input.hcl", hcl.Pos{Line: 1, Column: 1}); diagnostics.HasErrors() {
		return nil, diagnostics
	}

	return applyOutputConventions(string(hclwrite.Format(input)), config), nil
}
