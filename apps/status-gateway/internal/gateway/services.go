package gateway

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

// EnrollmentServer implements the Connect EnrollmentService. It mints the
// gateway-owned device identities that all subsequent reports are keyed by.
type EnrollmentServer struct {
	identity *IdentityStore
}

func NewEnrollmentServer(identity *IdentityStore) *EnrollmentServer {
	return &EnrollmentServer{identity: identity}
}

func (server *EnrollmentServer) EnrollDevice(
	_ context.Context,
	request *connect.Request[mev1.EnrollDeviceRequest],
) (*connect.Response[mev1.EnrollDeviceResponse], error) {
	message := request.Msg
	if message.GetKind() == mev1.DeviceKind_DEVICE_KIND_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("device kind is required"))
	}
	device, err := server.identity.Enroll(
		message.GetKind(),
		message.GetRole(),
		strings.TrimSpace(message.GetDisplayName()),
		strings.TrimSpace(message.GetModel()),
		time.Now(),
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&mev1.EnrollDeviceResponse{DeviceUid: device.UID}), nil
}

// NewAuthInterceptor rejects unauthenticated server-side calls whose bearer
// token is not a configured ingest token.
func NewAuthInterceptor(config Config) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			if request.Spec().IsClient {
				return next(ctx, request)
			}
			if !config.Authorized(request.Header().Get("Authorization")) {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized"))
			}
			return next(ctx, request)
		}
	}
}
