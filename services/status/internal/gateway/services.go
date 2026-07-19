package gateway

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	authv1 "github.com/pood1e/realtime-me/gen/go/realtime/me/auth/v1"
	sitev1 "github.com/pood1e/realtime-me/gen/go/realtime/me/site/v1"
	mev1 "github.com/pood1e/realtime-me/gen/go/realtime/me/status/v1"
	"github.com/pood1e/realtime-me/libs/go/authn"
	"github.com/pood1e/realtime-me/libs/go/serviceauth"
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
	targets  ScrapeTargetPolicy
}

func NewIngestServer(store *StatusStore, identity *IdentityStore, github *GitHubStatusPublisher, targets ScrapeTargetPolicy) *IngestServer {
	return &IngestServer{store: store, identity: identity, github: github, targets: targets}
}

func (server *IngestServer) requireEnrolled(uid string) (*EnrolledDevice, error) {
	device, ok := server.identity.Lookup(uid)
	if !ok {
		// not_found is what tells a client its cached uid is stale, so it drops
		// the uid and enrolls again instead of reporting into the void.
		return nil, connect.NewError(connect.CodeNotFound, errors.New("device is not enrolled"))
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

	mobile := &mev1.MobileState{
		DeviceUid:      message.GetDeviceUid(),
		DisplayName:    firstString(strings.TrimSpace(message.GetDisplayName()), device.DisplayName),
		Model:          firstString(strings.TrimSpace(message.GetModel()), device.Model),
		Phone:          message.GetPhone(),
		Watch:          message.GetWatch(),
		SwitchPresence: message.GetSwitchPresence(),
		UpdateTime:     timestamppb.New(time.Now().UTC()),
	}
	server.store.UpdateMobile(mobile)
	if mobile.GetWatch() != nil {
		server.github.Enqueue(mobile)
	}
	return connect.NewResponse(&mev1.ReportMobileStatusResponse{}), nil
}

// RegisterScrapeTargets declares a device's complete target set. Declaring an
// empty set deregisters the device.
func (server *IngestServer) RegisterScrapeTargets(
	_ context.Context,
	request *connect.Request[mev1.RegisterScrapeTargetsRequest],
) (*connect.Response[mev1.RegisterScrapeTargetsResponse], error) {
	message := request.Msg
	if _, err := server.requireEnrolled(message.GetDeviceUid()); err != nil {
		return nil, err
	}
	for _, target := range message.GetTargets() {
		if err := server.targets.Validate(target); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}
	if err := server.store.SetTargets(message.GetDeviceUid(), message.GetTargets()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&mev1.RegisterScrapeTargetsResponse{}), nil
}

// StatusServer implements the Connect StatusService.
type StatusServer struct {
	store      *StatusStore
	prometheus *PrometheusClient
	config     Config
	cache      prometheusCache
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

// GetProfile answers with the owner's identity, or reports that it is unavailable.
// The reason is a filesystem path, which is the operator's business and not a
// visitor's, so it stays in the log the gateway already wrote at startup.
func (server *ProfileServer) GetProfile(
	_ context.Context,
	_ *connect.Request[sitev1.GetProfileRequest],
) (*connect.Response[sitev1.GetProfileResponse], error) {
	profile, err := server.profile.Profile()
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("profile is unavailable"))
	}
	return connect.NewResponse(&sitev1.GetProfileResponse{Profile: profile}), nil
}

// ProjectsServer implements the Connect ProjectsService.
type ProjectsServer struct {
	projects *ProjectsService
}

func NewProjectsServer(projects *ProjectsService) *ProjectsServer {
	return &ProjectsServer{projects: projects}
}

// ListProjects answers with the curated projects, or reports that they are
// unavailable, for the same reason GetProfile does.
func (server *ProjectsServer) ListProjects(
	_ context.Context,
	_ *connect.Request[sitev1.ListProjectsRequest],
) (*connect.Response[sitev1.ListProjectsResponse], error) {
	projects, err := server.projects.List()
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("projects are unavailable"))
	}
	return connect.NewResponse(&sitev1.ListProjectsResponse{Projects: projects}), nil
}

// NewTokenAuthInterceptor rejects unauthenticated workload calls whose bearer
// token is not in the given scope's token set, except for the listed public
// procedures.
func NewTokenAuthInterceptor(tokens map[string]struct{}, publicProcedures ...string) connect.UnaryInterceptorFunc {
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

// NewOwnerInterceptor requires both the private management-plane credential
// and an OIDC identity with one bounded-context permission.
func NewOwnerInterceptor(internalAPIKey serviceauth.Key, verifier *authn.Verifier, permission authv1.Permission, publicProcedures ...string) connect.UnaryInterceptorFunc {
	allow := make(map[string]bool, len(publicProcedures))
	for _, procedure := range publicProcedures {
		allow[procedure] = true
	}
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			if request.Spec().IsClient || allow[request.Spec().Procedure] {
				return next(ctx, request)
			}
			if !internalAPIKey.Matches(request.Header().Get(serviceauth.Header)) {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized"))
			}
			_, err := verifier.Authenticate(ctx, request.Header().Get("Authorization"), permission)
			if errors.Is(err, authn.ErrUnavailable) {
				return nil, connect.NewError(connect.CodeUnavailable, errors.New("identity service unavailable"))
			}
			if errors.Is(err, authn.ErrPermissionDenied) {
				return nil, connect.NewError(connect.CodePermissionDenied, errors.New("permission denied"))
			}
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized"))
			}
			return next(ctx, request)
		}
	}
}
