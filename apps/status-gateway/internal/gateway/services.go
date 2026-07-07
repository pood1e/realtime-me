package gateway

import (
	"context"
	"errors"
	"log/slog"
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
	if err := server.github.Publish(ctx, mobile); err != nil {
		slog.Error("failed to publish github status", "error", err)
	}
	return connect.NewResponse(&mev1.ReportMobileStatusResponse{}), nil
}

func (server *IngestServer) ReportDeviceStatus(
	_ context.Context,
	request *connect.Request[mev1.ReportDeviceStatusRequest],
) (*connect.Response[mev1.ReportDeviceStatusResponse], error) {
	report := request.Msg.GetDevice()
	if _, err := server.requireEnrolled(report.GetDeviceUid()); err != nil {
		return nil, err
	}
	server.store.UpdateHost(deviceReportToState(report, timestamppb.New(time.Now().UTC())))
	return connect.NewResponse(&mev1.ReportDeviceStatusResponse{}), nil
}

func (server *IngestServer) ReportAgentStatus(
	_ context.Context,
	request *connect.Request[mev1.ReportAgentStatusRequest],
) (*connect.Response[mev1.ReportAgentStatusResponse], error) {
	message := request.Msg
	device, err := server.requireEnrolled(message.GetDeviceUid())
	if err != nil {
		return nil, err
	}
	kind := strings.TrimSpace(message.GetKind())
	if kind == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("agent kind is required"))
	}
	agent := &mev1.Agent{
		Uid:                    agentUID(message.GetDeviceUid(), kind),
		Kind:                   kind,
		DeviceUid:              message.GetDeviceUid(),
		DisplayName:            device.DisplayName,
		State:                  message.GetState(),
		BudgetRemainingPercent: message.BudgetRemainingPercent,
		UpdateTime:             timestamppb.New(time.Now().UTC()),
	}
	server.store.UpdateAgent(agent)
	return connect.NewResponse(&mev1.ReportAgentStatusResponse{}), nil
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

func deviceReportToState(report *mev1.DeviceReport, now *timestamppb.Timestamp) *mev1.DeviceState {
	children := make([]*mev1.DeviceState, 0, len(report.GetChildren()))
	for _, child := range report.GetChildren() {
		children = append(children, deviceReportToState(child, now))
	}
	return &mev1.DeviceState{
		DeviceUid:   report.GetDeviceUid(),
		DisplayName: report.GetDisplayName(),
		Model:       report.GetModel(),
		Kind:        report.GetKind(),
		Role:        report.GetRole(),
		State:       report.GetState(),
		Metrics:     report.GetMetrics(),
		Media:       report.GetMedia(),
		Accessories: report.GetAccessories(),
		Children:    children,
		UpdateTime:  now,
	}
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
// token is not a configured ingest token, except for the listed public
// procedures.
func NewAuthInterceptor(config Config, publicProcedures ...string) connect.UnaryInterceptorFunc {
	allow := make(map[string]bool, len(publicProcedures))
	for _, procedure := range publicProcedures {
		allow[procedure] = true
	}
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			if request.Spec().IsClient || allow[request.Spec().Procedure] {
				return next(ctx, request)
			}
			if !config.Authorized(request.Header().Get("Authorization")) {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized"))
			}
			return next(ctx, request)
		}
	}
}
