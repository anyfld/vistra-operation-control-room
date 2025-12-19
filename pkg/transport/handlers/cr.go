package handlers

import (
	"context"
	"math"
	"time"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/usecase"
)

const (
	systemStatusInterval = 500 * time.Millisecond
	cineResultInterval   = 200 * time.Millisecond
)

type CRHandler struct {
	uc usecase.CRInteractor
}

func NewCRHandler(uc usecase.CRInteractor) *CRHandler {
	return &CRHandler{uc: uc}
}

func safeUint32(num int) uint32 {
	if num < 0 {
		return 0
	}

	if num > math.MaxUint32 {
		return math.MaxUint32
	}

	return uint32(num)
}

func (h *CRHandler) RegisterMasterMF(
	ctx context.Context,
	req *connect.Request[protov1.RegisterMasterMFRequest],
) (*connect.Response[protov1.RegisterMasterMFResponse], error) {
	mf, err := h.uc.RegisterMasterMF(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.RegisterMasterMFResponse{MasterMf: mf}), nil
}

func (h *CRHandler) UnregisterMasterMF(
	ctx context.Context,
	req *connect.Request[protov1.UnregisterMasterMFRequest],
) (*connect.Response[protov1.UnregisterMasterMFResponse], error) {
	ok, err := h.uc.UnregisterMasterMF(ctx, req.Msg.GetMasterMfId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.UnregisterMasterMFResponse{Success: ok}), nil
}

func (h *CRHandler) ListMasterMFs(
	ctx context.Context,
	req *connect.Request[protov1.ListMasterMFsRequest],
) (*connect.Response[protov1.ListMasterMFsResponse], error) {
	mfs, err := h.uc.ListMasterMFs(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.ListMasterMFsResponse{
		MasterMfs:     mfs,
		NextPageToken: "",
		TotalCount:    safeUint32(len(mfs)),
	}), nil
}

func (h *CRHandler) GetMasterMF(
	ctx context.Context,
	req *connect.Request[protov1.GetMasterMFRequest],
) (*connect.Response[protov1.GetMasterMFResponse], error) {
	mf, err := h.uc.GetMasterMF(ctx, req.Msg.GetMasterMfId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.GetMasterMFResponse{MasterMf: mf}), nil
}

func (h *CRHandler) GetSystemStatus(
	ctx context.Context,
	req *connect.Request[protov1.GetSystemStatusRequest],
) (*connect.Response[protov1.GetSystemStatusResponse], error) {
	mfs, err := h.uc.ListMasterMFs(ctx)
	if err != nil {
		return nil, err
	}

	cams, err := h.uc.ListAllCameras(ctx)
	if err != nil {
		return nil, err
	}

	status := &protov1.SystemStatus{
		Health:              protov1.SystemHealthStatus_SYSTEM_HEALTH_STATUS_HEALTHY,
		OnlineMasterMfCount: safeUint32(len(mfs)),
		OnlineCameraCount:   safeUint32(len(cams)),
		ActiveStreamCount:   0,
		UpdatedAtMs:         time.Now().UnixMilli(),
	}

	return connect.NewResponse(&protov1.GetSystemStatusResponse{Status: status}), nil
}

func (h *CRHandler) StreamSystemStatus(
	ctx context.Context,
	req *connect.Request[protov1.StreamSystemStatusRequest],
	stream *connect.ServerStream[protov1.StreamSystemStatusResponse],
) error {
	for range 5 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		mfs, err := h.uc.ListMasterMFs(ctx)
		if err != nil {
			return err
		}

		cams, err := h.uc.ListAllCameras(ctx)
		if err != nil {
			return err
		}

		if err := stream.Send(&protov1.StreamSystemStatusResponse{
			Status: &protov1.SystemStatus{
				Health:              protov1.SystemHealthStatus_SYSTEM_HEALTH_STATUS_HEALTHY,
				OnlineMasterMfCount: safeUint32(len(mfs)),
				OnlineCameraCount:   safeUint32(len(cams)),
				ActiveStreamCount:   0,
				UpdatedAtMs:         time.Now().UnixMilli(),
			},
			TimestampMs: time.Now().UnixMilli(),
		}); err != nil {
			return err
		}

		time.Sleep(systemStatusInterval)
	}

	return nil
}

func (h *CRHandler) ListAllCameras(
	ctx context.Context,
	req *connect.Request[protov1.ListAllCamerasRequest],
) (*connect.Response[protov1.ListAllCamerasResponse], error) {
	cams, err := h.uc.ListAllCameras(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.ListAllCamerasResponse{
		Cameras:       cams,
		NextPageToken: "",
		TotalCount:    safeUint32(len(cams)),
	}), nil
}

func (h *CRHandler) GetCameraStatus(
	ctx context.Context,
	req *connect.Request[protov1.GetCameraStatusRequest],
) (*connect.Response[protov1.GetCameraStatusResponse], error) {
	cam, err := h.uc.GetCamera(ctx, req.Msg.GetCameraId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.GetCameraStatusResponse{Camera: cam}), nil
}

func (h *CRHandler) PushConfiguration(
	ctx context.Context,
	req *connect.Request[protov1.PushConfigurationRequest],
) (*connect.Response[protov1.PushConfigurationResponse], error) {
	cfg := req.Msg.GetConfiguration()
	ok, failed := h.uc.PushConfiguration(ctx, cfg, req.Msg.GetTargetMasterMfIds())

	return connect.NewResponse(&protov1.PushConfigurationResponse{Success: ok, FailedMasterMfIds: failed}), nil
}

func (h *CRHandler) GetConfiguration(
	ctx context.Context,
	req *connect.Request[protov1.GetConfigurationRequest],
) (*connect.Response[protov1.GetConfigurationResponse], error) {
	cfg := h.uc.GetConfiguration(ctx, req.Msg.GetMasterMfId())

	return connect.NewResponse(&protov1.GetConfigurationResponse{Configuration: cfg}), nil
}

func (h *CRHandler) SendCinematographyInstruction(
	ctx context.Context,
	req *connect.Request[protov1.SendCinematographyInstructionRequest],
) (*connect.Response[protov1.SendCinematographyInstructionResponse], error) {
	res, err := h.uc.SendCinematographyInstruction(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (h *CRHandler) StreamCinematographyResults(
	ctx context.Context,
	req *connect.Request[protov1.StreamCinematographyResultsRequest],
	stream *connect.ServerStream[protov1.StreamCinematographyResultsResponse],
) error {
	for range 3 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := stream.Send(&protov1.StreamCinematographyResultsResponse{
			Result: &protov1.CinematographyResult{
				InstructionId: "instr-stream",
				CameraId:      "cam-1",
				Success:       true,
				ErrorMessage:  "",
				AppliedPtz:    nil,
				CompletedAtMs: time.Now().UnixMilli(),
			},
			TimestampMs: time.Now().UnixMilli(),
		}); err != nil {
			return err
		}

		time.Sleep(cineResultInterval)
	}

	return nil
}
