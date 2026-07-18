package app

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

func validateName(name string) error {
	if strings.TrimSpace(name) == "" || name == "." || name == ".." || len(name) > 255 || !utf8.ValidString(name) {
		return fmt.Errorf("%w: invalid name", domain.ErrInvalidArgument)
	}
	for _, value := range name {
		if value == '/' || value == '\\' || value == 0 || unicode.IsControl(value) {
			return fmt.Errorf("%w: invalid name", domain.ErrInvalidArgument)
		}
	}
	return nil
}

func validateDisplayName(value string) error {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 255 || !utf8.ValidString(value) {
		return fmt.Errorf("%w: invalid display name", domain.ErrInvalidArgument)
	}
	return nil
}

func emptyToNil(value *string) *string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	copy := *value
	return &copy
}
