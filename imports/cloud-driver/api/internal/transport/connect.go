package transport

import (
	"errors"

	"connectrpc.com/connect"
	"example.com/cloud-drive/api/internal/auth"
	"example.com/cloud-drive/api/internal/domain"
)

func sessionError(err error) error {
	if errors.Is(err, auth.ErrUnauthenticated) {
		return connect.NewError(connect.CodeUnauthenticated, auth.ErrUnauthenticated)
	}
	return connect.NewError(connect.CodeInternal, errors.New("internal server error"))
}

func connectError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidArgument):
		return connect.NewError(connect.CodeInvalidArgument, errors.New("invalid request"))
	case errors.Is(err, domain.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, errors.New("resource not found"))
	case errors.Is(err, domain.ErrForbidden):
		return connect.NewError(connect.CodePermissionDenied, errors.New("access denied"))
	case errors.Is(err, domain.ErrResourceExhausted):
		return connect.NewError(connect.CodeResourceExhausted, errors.New("storage capacity is unavailable"))
	case errors.Is(err, domain.ErrConflict):
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("request cannot be applied"))
	case errors.Is(err, domain.ErrProviderReconnectRequired):
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("provider account must be reconnected"))
	case errors.Is(err, domain.ErrUnavailable):
		return connect.NewError(connect.CodeUnavailable, errors.New("service temporarily unavailable"))
	default:
		return connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}
}
