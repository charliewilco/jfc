package jfc

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"unicode/utf8"
)

func formatCSV(input []byte, config Config) ([]byte, error) {
	if err := validateDelimited(input, ','); err != nil {
		return nil, err
	}
	return applyFinalNewlineOnly(input, config), nil
}

func formatTSV(input []byte, config Config) ([]byte, error) {
	if err := validateDelimited(input, '\t'); err != nil {
		return nil, err
	}
	return applyFinalNewlineOnly(input, config), nil
}

func validateDelimited(input []byte, comma rune) error {
	if !utf8.Valid(input) {
		return fmt.Errorf("input is not valid UTF-8")
	}

	reader := csv.NewReader(bytes.NewReader(input))
	reader.Comma = comma
	reader.FieldsPerRecord = 0
	reader.ReuseRecord = true

	for {
		if _, err := reader.Read(); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}
