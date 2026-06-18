package jfc

import (
	"fmt"
	"slices"
	"unicode/utf8"

	"github.com/tailscale/hujson"
)

func formatJSONC(input []byte, config Config) ([]byte, error) {
	if !utf8.Valid(input) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}

	root, err := hujson.Parse(input)
	if err != nil {
		return nil, err
	}

	if config.SortKeys {
		sortJSONCKeys(&root)
	}
	root.Format()
	return applyOutputConventions(string(root.Pack()), config), nil
}

func sortJSONCKeys(value *hujson.Value) {
	switch typed := value.Value.(type) {
	case *hujson.Object:
		for i := range typed.Members {
			sortJSONCKeys(&typed.Members[i].Value)
		}
		slices.SortStableFunc(typed.Members, func(a hujson.ObjectMember, b hujson.ObjectMember) int {
			aKey := jsonCObjectKey(a)
			bKey := jsonCObjectKey(b)
			switch {
			case aKey < bKey:
				return -1
			case aKey > bKey:
				return 1
			default:
				return 0
			}
		})
	case *hujson.Array:
		for i := range typed.Elements {
			sortJSONCKeys(&typed.Elements[i])
		}
	}
}

func jsonCObjectKey(member hujson.ObjectMember) string {
	if literal, ok := member.Name.Value.(hujson.Literal); ok && literal.Kind() == '"' {
		return literal.String()
	}
	return string(member.Name.Pack())
}
