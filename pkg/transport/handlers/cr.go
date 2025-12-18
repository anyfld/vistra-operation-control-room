package handlers

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
)

type CRHandler struct{}

func (h *CRHandler) RegisterMasterMF(
	ctx context.Context,
	req *connect.Request[protov1.RegisterMasterMFRequest],
) (*connect.Response[protov1.RegisterMasterMFResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CRService.RegisterMasterMF is not implemented"),
	)
}

func (h *CRHandler) UnregisterMasterMF(
	ctx context.Context,
	req *connect.Request[protov1.UnregisterMasterMFRequest],
) (*connect.Response[protov1.UnregisterMasterMFResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CRService.UnregisterMasterMF is not implemented"),
	)
}

func (h *CRHandler) ListMasterMFs(
	ctx context.Context,
	req *connect.Request[protov1.ListMasterMFsRequest],
) (*connect.Response[protov1.ListMasterMFsResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CRService.ListMasterMFs is not implemented"),
	)
}

func (h *CRHandler) GetMasterMF(
	ctx context.Context,
	req *connect.Request[protov1.GetMasterMFRequest],
) (*connect.Response[protov1.GetMasterMFResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CRService.GetMasterMF is not implemented"),
	)
}

func (h *CRHandler) GetSystemStatus(
	ctx context.Context,
	req *connect.Request[protov1.GetSystemStatusRequest],
) (*connect.Response[protov1.GetSystemStatusResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CRService.GetSystemStatus is not implemented"),
	)
}

func (h *CRHandler) StreamSystemStatus(
	ctx context.Context,
	req *connect.Request[protov1.StreamSystemStatusRequest],
	stream *connect.ServerStream[protov1.StreamSystemStatusResponse],
) error {
	return connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CRService.StreamSystemStatus is not implemented"),
	)
}

func (h *CRHandler) ListAllCameras(
	ctx context.Context,
	req *connect.Request[protov1.ListAllCamerasRequest],
) (*connect.Response[protov1.ListAllCamerasResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CRService.ListAllCameras is not implemented"),
	)
}

func (h *CRHandler) GetCameraStatus(
	ctx context.Context,
	req *connect.Request[protov1.GetCameraStatusRequest],
) (*connect.Response[protov1.GetCameraStatusResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CRService.GetCameraStatus is not implemented"),
	)
}

func (h *CRHandler) PushConfiguration(
	ctx context.Context,
	req *connect.Request[protov1.PushConfigurationRequest],
) (*connect.Response[protov1.PushConfigurationResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CRService.PushConfiguration is not implemented"),
	)
}

func (h *CRHandler) GetConfiguration(
	ctx context.Context,
	req *connect.Request[protov1.GetConfigurationRequest],
) (*connect.Response[protov1.GetConfigurationResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CRService.GetConfiguration is not implemented"),
	)
}

func (h *CRHandler) SendCinematographyInstruction(
	ctx context.Context,
	req *connect.Request[protov1.SendCinematographyInstructionRequest],
) (*connect.Response[protov1.SendCinematographyInstructionResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CRService.SendCinematographyInstruction is not implemented"),
	)
}

func (h *CRHandler) StreamCinematographyResults(
	ctx context.Context,
	req *connect.Request[protov1.StreamCinematographyResultsRequest],
	stream *connect.ServerStream[protov1.StreamCinematographyResultsResponse],
) error {
	return connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.CRService.StreamCinematographyResults is not implemented"),
	)
}
