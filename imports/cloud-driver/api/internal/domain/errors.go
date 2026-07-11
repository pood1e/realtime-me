package domain

import "errors"

var (
	// ErrInvalidArgument marks invalid client input.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrNotFound marks a missing resource.
	ErrNotFound = errors.New("not found")
	// ErrConflict marks an invalid state transition or hierarchy conflict.
	ErrConflict = errors.New("conflict")
	// ErrForbidden marks a resource outside an authorized share scope.
	ErrForbidden = errors.New("forbidden")
	// ErrUnavailable marks a temporary dependency failure.
	ErrUnavailable = errors.New("unavailable")
	// ErrResourceExhausted marks insufficient storage capacity.
	ErrResourceExhausted = errors.New("resource exhausted")
	// ErrProviderReconnectRequired marks an expired or revoked external account credential.
	ErrProviderReconnectRequired = errors.New("provider reconnect required")
)
