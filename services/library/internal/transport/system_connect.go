package transport

import (
	"context"

	"connectrpc.com/connect"
	systemv1 "github.com/pood1e/realtime-me/gen/go/realtime/me/library/system/v1"
	"github.com/pood1e/realtime-me/services/library/internal/app"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type systemServer struct{ service *app.SystemService }

func (s *systemServer) Check(ctx context.Context, _ *connect.Request[systemv1.CheckRequest]) (*connect.Response[systemv1.CheckResponse], error) {
	healthy, workerHealthy, health, freeBytes, err := s.service.Check(ctx)
	if err != nil {
		return connect.NewResponse(&systemv1.CheckResponse{Healthy: false}), nil
	}
	response := &systemv1.CheckResponse{Healthy: healthy, WorkerHealthy: workerHealthy, PendingJobs: health.PendingJobs, AvailableBytes: freeBytes}
	if health.HeartbeatTime != nil {
		response.WorkerHeartbeatTime = timestamppb.New(*health.HeartbeatTime)
	}
	return connect.NewResponse(response), nil
}
