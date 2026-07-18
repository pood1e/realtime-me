package app

import (
	"fmt"

	"example.com/cloud-drive/api/internal/domain"
)

func domainNotFound(resource string) error {
	return fmt.Errorf("%w: %s", domain.ErrNotFound, resource)
}
