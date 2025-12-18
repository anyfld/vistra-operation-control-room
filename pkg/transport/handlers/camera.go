package handlers

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
)

type CameraHandler struct{}

func (h *CameraHandler) RegisterCamera(
	ctx context.Context,
	req *connect.Request[protov1.RegisterCameraRequest],
) (*connect.Response[protov1.RegisterCameraResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CameraService.RegisterCamera is not implemented"),
	)
}

func (h *CameraHandler) UnregisterCamera(
	ctx context.Context,
	req *connect.Request[protov1.UnregisterCameraRequest],
) (*connect.Response[protov1.UnregisterCameraResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CameraService.UnregisterCamera is not implemented"),
	)
}

func (h *CameraHandler) UpdateCamera(
	ctx context.Context,
	req *connect.Request[protov1.UpdateCameraRequest],
) (*connect.Response[protov1.UpdateCameraResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CameraService.UpdateCamera is not implemented"),
	)
}

func (h *CameraHandler) GetCamera(
	ctx context.Context,
	req *connect.Request[protov1.GetCameraRequest],
) (*connect.Response[protov1.GetCameraResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CameraService.GetCamera is not implemented"),
	)
}

func (h *CameraHandler) ListCameras(
	ctx context.Context,
	req *connect.Request[protov1.ListCamerasRequest],
) (*connect.Response[protov1.ListCamerasResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CameraService.ListCameras is not implemented"),
	)
}

func (h *CameraHandler) SwitchCameraMode(
	ctx context.Context,
	req *connect.Request[protov1.SwitchCameraModeRequest],
) (*connect.Response[protov1.SwitchCameraModeResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CameraService.SwitchCameraMode is not implemented"),
	)
}

func (h *CameraHandler) Heartbeat(
	ctx context.Context,
	req *connect.Request[protov1.HeartbeatRequest],
) (*connect.Response[protov1.HeartbeatResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CameraService.Heartbeat is not implemented"),
	)
}

func (h *CameraHandler) StreamConnectionStatus(
	ctx context.Context,
	req *connect.Request[protov1.StreamConnectionStatusRequest],
	stream *connect.ServerStream[protov1.StreamConnectionStatusResponse],
) error {
	return connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CameraService.StreamConnectionStatus is not implemented"),
	)
}

func (h *CameraHandler) GetCameraCapabilities(
	ctx context.Context,
	req *connect.Request[protov1.GetCameraCapabilitiesRequest],
) (*connect.Response[protov1.GetCameraCapabilitiesResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CameraService.GetCameraCapabilities is not implemented"),
	)
}
