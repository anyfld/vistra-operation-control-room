package handlers

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
)

type FDHandler struct{}

func (h *FDHandler) ExecuteCinematography(
	ctx context.Context,
	req *connect.Request[protov1.ExecuteCinematographyRequest],
) (*connect.Response[protov1.ExecuteCinematographyResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.FDService.ExecuteCinematography is not implemented"),
	)
}

func (h *FDHandler) StreamCinematographyInstructions(
	ctx context.Context,
	req *connect.Request[protov1.FDServiceStreamCinematographyInstructionsRequest],
	stream *connect.ServerStream[protov1.FDServiceStreamCinematographyInstructionsResponse],
) error {
	return connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.FDService.StreamCinematographyInstructions is not implemented"),
	)
}

func (h *FDHandler) ProcessImage(
	ctx context.Context,
	req *connect.Request[protov1.ProcessImageRequest],
) (*connect.Response[protov1.ProcessImageResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.FDService.ProcessImage is not implemented"),
	)
}

func (h *FDHandler) StartPatternMatching(
	ctx context.Context,
	req *connect.Request[protov1.StartPatternMatchingRequest],
) (*connect.Response[protov1.StartPatternMatchingResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.FDService.StartPatternMatching is not implemented"),
	)
}

func (h *FDHandler) StopPatternMatching(
	ctx context.Context,
	req *connect.Request[protov1.StopPatternMatchingRequest],
) (*connect.Response[protov1.StopPatternMatchingResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.FDService.StopPatternMatching is not implemented"),
	)
}

func (h *FDHandler) StreamPatternMatchResults(
	ctx context.Context,
	req *connect.Request[protov1.StreamPatternMatchResultsRequest],
	stream *connect.ServerStream[protov1.StreamPatternMatchResultsResponse],
) error {
	return connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.FDService.StreamPatternMatchResults is not implemented"),
	)
}

func (h *FDHandler) CalculateFraming(
	ctx context.Context,
	req *connect.Request[protov1.CalculateFramingRequest],
) (*connect.Response[protov1.CalculateFramingResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.FDService.CalculateFraming is not implemented"),
	)
}

func (h *FDHandler) SendControlCommand(
	ctx context.Context,
	req *connect.Request[protov1.SendControlCommandRequest],
) (*connect.Response[protov1.SendControlCommandResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.FDService.SendControlCommand is not implemented"),
	)
}

func (h *FDHandler) StreamControlCommands(
	ctx context.Context,
	req *connect.Request[protov1.StreamControlCommandsRequest],
	stream *connect.ServerStream[protov1.StreamControlCommandsResponse],
) error {
	return connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.FDService.StreamControlCommands is not implemented"),
	)
}

func (h *FDHandler) ReportCameraState(
	ctx context.Context,
	req *connect.Request[protov1.ReportCameraStateRequest],
) (*connect.Response[protov1.ReportCameraStateResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.FDService.ReportCameraState is not implemented"),
	)
}

func (h *FDHandler) GetCameraState(
	ctx context.Context,
	req *connect.Request[protov1.GetCameraStateRequest],
) (*connect.Response[protov1.GetCameraStateResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.FDService.GetCameraState is not implemented"),
	)
}
