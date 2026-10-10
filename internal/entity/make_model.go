package entity

import (
	"fmt"
	"strings"
	"unicode"
)

// checkMakeModel returns an error if the make or model name of a camera or lens contains a control character.
func checkMakeModel(makeName, modelName string) error {
	if strings.IndexFunc(makeName, unicode.IsControl) >= 0 || strings.IndexFunc(modelName, unicode.IsControl) >= 0 {
		return fmt.Errorf("%w: make and model must not contain control characters", ErrInvalidValue)
	}

	return nil
}
