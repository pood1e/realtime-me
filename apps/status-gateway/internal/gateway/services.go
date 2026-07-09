package gateway

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

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

// IngestServer implements the Connect IngestService. Every report is keyed by an
// enrolled device uid; unknown uids are rejected.
type IngestServer struct {
	store    *StatusStore
	identity *IdentityStore
	github   *GitHubStatusPublisher
}

func NewIngestServer(store *StatusStore, identity *IdentityStore, github *GitHubStatusPublisher) *IngestServer {
	return &IngestServer{store: store, identity: identity, github: github}
}

func (server *IngestServer) requireEnrolled(uid string) (*EnrolledDevice, error) {
	device, ok := server.identity.Lookup(uid)
	if !ok {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("device is not enrolled"))
	}
	return device, nil
}

func (server *IngestServer) ReportMobileStatus(
	ctx context.Context,
	request *connect.Request[mev1.ReportMobileStatusRequest],
) (*connect.Response[mev1.ReportMobileStatusResponse], error) {
	message := request.Msg
	device, err := server.requireEnrolled(message.GetDeviceUid())
	if err != nil {
		return nil, err
	}

	watch := message.GetWatch()
	if watch.GetWatchState().GetWristState() == mev1.WristState_WRIST_STATE_OFF_WRIST {
		watch.HeartRate = nil
	}
	mobile := &mev1.MobileState{
		DeviceUid:   message.GetDeviceUid(),
		DisplayName: firstString(strings.TrimSpace(message.GetDisplayName()), device.DisplayName),
		Model:       firstString(strings.TrimSpace(message.GetModel()), device.Model),
		Phone:       message.GetPhone(),
		Watch:       watch,
		UpdateTime:  timestamppb.New(time.Now().UTC()),
	}
	server.store.UpdateMobile(mobile)
	server.github.Enqueue(mobile)
	return connect.NewResponse(&mev1.ReportMobileStatusResponse{}), nil
}

func (server *IngestServer) RegisterScrapeTargets(
	_ context.Context,
	request *connect.Request[mev1.RegisterScrapeTargetsRequest],
) (*connect.Response[mev1.RegisterScrapeTargetsResponse], error) {
	if err := server.store.RegisterTargets(request.Msg.GetTargets()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&mev1.RegisterScrapeTargetsResponse{}), nil
}

// StatusServer implements the Connect StatusService.
type StatusServer struct {
	store      *StatusStore
	prometheus *PrometheusClient
	config     Config
}

func NewStatusServer(store *StatusStore, prometheus *PrometheusClient, config Config) *StatusServer {
	return &StatusServer{store: store, prometheus: prometheus, config: config}
}

func (server *StatusServer) GetPublicStatus(
	ctx context.Context,
	_ *connect.Request[mev1.GetPublicStatusRequest],
) (*connect.Response[mev1.GetPublicStatusResponse], error) {
	return connect.NewResponse(&mev1.GetPublicStatusResponse{Status: server.buildPublicStatus(ctx)}), nil
}

func (server *StatusServer) GetInternalStatus(
	ctx context.Context,
	_ *connect.Request[mev1.GetInternalStatusRequest],
) (*connect.Response[mev1.GetInternalStatusResponse], error) {
	return connect.NewResponse(&mev1.GetInternalStatusResponse{Status: server.buildInternalStatus(ctx)}), nil
}

// ProfileServer implements the Connect ProfileService.
type ProfileServer struct {
	profile *ProfileService
}

func NewProfileServer(profile *ProfileService) *ProfileServer {
	return &ProfileServer{profile: profile}
}

func (server *ProfileServer) GetProfilePage(
	_ context.Context,
	_ *connect.Request[mev1.GetProfilePageRequest],
) (*connect.Response[mev1.GetProfilePageResponse], error) {
	return connect.NewResponse(&mev1.GetProfilePageResponse{Page: server.profile.Page(time.Now())}), nil
}

// NewAuthInterceptor rejects unauthenticated server-side calls whose bearer
// token is not in the given scope's token set, except for the listed public
// procedures.
func NewAuthInterceptor(tokens map[string]struct{}, publicProcedures ...string) connect.UnaryInterceptorFunc {
	allow := make(map[string]bool, len(publicProcedures))
	for _, procedure := range publicProcedures {
		allow[procedure] = true
	}
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			if request.Spec().IsClient || allow[request.Spec().Procedure] {
				return next(ctx, request)
			}
			if !authorizedWith(tokens, request.Header().Get("Authorization")) {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized"))
			}
			return next(ctx, request)
		}
	}
}
