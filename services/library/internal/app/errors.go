package app

import (
	"fmt"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

func domainNotFound(resource string) error {
	return fmt.Errorf("%w: %s", domain.ErrNotFound, resource)
}
