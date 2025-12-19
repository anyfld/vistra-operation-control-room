package handlers

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/usecase"
)

const (
	cameraPollingIntervalMs = 500
)

type CameraHandler struct {
	uc usecase.CameraInteractor
}

func NewCameraHandler(uc usecase.CameraInteractor) *CameraHandler {
	return &CameraHandler{uc: uc}
}

func (h *CameraHandler) RegisterCamera(
	ctx context.Context,
	req *connect.Request[protov1.RegisterCameraRequest],
) (*connect.Response[protov1.RegisterCameraResponse], error) {
	camera, err := h.uc.RegisterCamera(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	if camera == nil {
		return nil, connect.NewError(
			connect.CodeInternal,
			errors.New("failed to register camera"),
		)
	}

	return connect.NewResponse(&protov1.RegisterCameraResponse{Camera: camera}), nil
}

func (h *CameraHandler) UnregisterCamera(
	ctx context.Context,
	req *connect.Request[protov1.UnregisterCameraRequest],
) (*connect.Response[protov1.UnregisterCameraResponse], error) {
	success, err := h.uc.UnregisterCamera(ctx, req.Msg.GetCameraId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.UnregisterCameraResponse{Success: success}), nil
}

func (h *CameraHandler) UpdateCamera(
	ctx context.Context,
	req *connect.Request[protov1.UpdateCameraRequest],
) (*connect.Response[protov1.UpdateCameraResponse], error) {
	camera, err := h.uc.UpdateCamera(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	if camera == nil {
		return nil, connect.NewError(
			connect.CodeNotFound,
			errors.New("camera not found"),
		)
	}

	return connect.NewResponse(&protov1.UpdateCameraResponse{Camera: camera}), nil
}

func (h *CameraHandler) GetCamera(
	ctx context.Context,
	req *connect.Request[protov1.GetCameraRequest],
) (*connect.Response[protov1.GetCameraResponse], error) {
	camera, connection, capabilities, err := h.uc.GetCamera(ctx, req.Msg.GetCameraId())
	if err != nil {
		return nil, err
	}

	if camera == nil {
		return nil, connect.NewError(
			connect.CodeNotFound,
			errors.New("camera not found"),
		)
	}

	return connect.NewResponse(&protov1.GetCameraResponse{
		Camera:       camera,
		Connection:   connection,
		Capabilities: capabilities,
	}), nil
}

func (h *CameraHandler) ListCameras(
	ctx context.Context,
	req *connect.Request[protov1.ListCamerasRequest],
) (*connect.Response[protov1.ListCamerasResponse], error) {
	cameras, err := h.uc.ListCameras(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	count := len(cameras)

	var totalCount uint32

	if count > 0 && count <= int(^uint32(0)) {
		totalCount = uint32(count)
	} else {
		totalCount = ^uint32(0)
	}

	return connect.NewResponse(&protov1.ListCamerasResponse{
		Cameras:       cameras,
		NextPageToken: "",
		TotalCount:    totalCount,
	}), nil
}

func (h *CameraHandler) SwitchCameraMode(
	ctx context.Context,
	req *connect.Request[protov1.SwitchCameraModeRequest],
) (*connect.Response[protov1.SwitchCameraModeResponse], error) {
	success, err := h.uc.SwitchCameraMode(ctx, req.Msg.GetCameraId(), req.Msg.GetTargetMode())
	if err != nil {
		return nil, err
	}

	camera, _, _, err := h.uc.GetCamera(ctx, req.Msg.GetCameraId())
	if err != nil {
		return nil, err
	}

	if !success {
		return connect.NewResponse(&protov1.SwitchCameraModeResponse{
			Success:      false,
			Camera:       camera,
			ErrorMessage: "camera not found",
		}), nil
	}

	return connect.NewResponse(&protov1.SwitchCameraModeResponse{
		Success:      true,
		Camera:       camera,
		ErrorMessage: "",
	}), nil
}

func (h *CameraHandler) StreamConnectionStatus(
	ctx context.Context,
	req *connect.Request[protov1.StreamConnectionStatusRequest],
	stream *connect.ServerStream[protov1.StreamConnectionStatusResponse],
) error {
	cameraIDs := req.Msg.GetCameraIds()

	previousStatuses := make(map[string]protov1.CameraStatus)

	for range 10 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		statuses, err := h.uc.GetAllConnectionStatuses(ctx, cameraIDs)
		if err != nil {
			return err
		}

		for cameraID, currentStatus := range statuses {
			previousStatus, hadPrevious := previousStatuses[cameraID]

			if !hadPrevious || previousStatus != currentStatus {
				if err := stream.Send(&protov1.StreamConnectionStatusResponse{
					CameraId:         cameraID,
					PreviousStatus:   previousStatus,
					CurrentStatus:    currentStatus,
					TimestampMs:      time.Now().UnixMilli(),
					DisconnectReason: "",
				}); err != nil {
					return err
				}

				previousStatuses[cameraID] = currentStatus
			}
		}

		time.Sleep(cameraPollingIntervalMs * time.Millisecond)
	}

	return nil
}
