package handlers

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
)

type MDHandler struct{}

func (h *MDHandler) ReceiveCinematographyInstruction(
	ctx context.Context,
	req *connect.Request[protov1.ReceiveCinematographyInstructionRequest],
) (*connect.Response[protov1.ReceiveCinematographyInstructionResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.ReceiveCinematographyInstruction is not implemented"),
	)
}

func (h *MDHandler) StreamCinematographyInstructions(
	ctx context.Context,
	req *connect.Request[protov1.MDServiceStreamCinematographyInstructionsRequest],
	stream *connect.ServerStream[protov1.MDServiceStreamCinematographyInstructionsResponse],
) error {
	return connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.StreamCinematographyInstructions is not implemented"),
	)
}

func (h *MDHandler) ForwardToFD(
	ctx context.Context,
	req *connect.Request[protov1.ForwardToFDRequest],
) (*connect.Response[protov1.ForwardToFDResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.ForwardToFD is not implemented"),
	)
}

func (h *MDHandler) ConfigureVideoOutput(
	ctx context.Context,
	req *connect.Request[protov1.ConfigureVideoOutputRequest],
) (*connect.Response[protov1.ConfigureVideoOutputResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.ConfigureVideoOutput is not implemented"),
	)
}

func (h *MDHandler) GetVideoOutputStatus(
	ctx context.Context,
	req *connect.Request[protov1.GetVideoOutputStatusRequest],
) (*connect.Response[protov1.GetVideoOutputStatusResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.GetVideoOutputStatus is not implemented"),
	)
}

func (h *MDHandler) ListVideoOutputs(
	ctx context.Context,
	req *connect.Request[protov1.ListVideoOutputsRequest],
) (*connect.Response[protov1.ListVideoOutputsResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.ListVideoOutputs is not implemented"),
	)
}

func (h *MDHandler) StartStreaming(
	ctx context.Context,
	req *connect.Request[protov1.StartStreamingRequest],
) (*connect.Response[protov1.StartStreamingResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.StartStreaming is not implemented"),
	)
}

func (h *MDHandler) StopStreaming(
	ctx context.Context,
	req *connect.Request[protov1.StopStreamingRequest],
) (*connect.Response[protov1.StopStreamingResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.StopStreaming is not implemented"),
	)
}

func (h *MDHandler) SwitchSource(
	ctx context.Context,
	req *connect.Request[protov1.SwitchSourceRequest],
) (*connect.Response[protov1.SwitchSourceResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.SwitchSource is not implemented"),
	)
}

func (h *MDHandler) GetStreamingStatus(
	ctx context.Context,
	req *connect.Request[protov1.GetStreamingStatusRequest],
) (*connect.Response[protov1.GetStreamingStatusResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.GetStreamingStatus is not implemented"),
	)
}

func (h *MDHandler) StreamStreamingEvents(
	ctx context.Context,
	req *connect.Request[protov1.StreamStreamingEventsRequest],
	stream *connect.ServerStream[protov1.StreamStreamingEventsResponse],
) error {
	return connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.StreamStreamingEvents is not implemented"),
	)
}

func (h *MDHandler) SendToLLM(
	ctx context.Context,
	req *connect.Request[protov1.SendToLLMRequest],
) (*connect.Response[protov1.SendToLLMResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.SendToLLM is not implemented"),
	)
}

func (h *MDHandler) ReceiveFromLLM(
	ctx context.Context,
	req *connect.Request[protov1.ReceiveFromLLMRequest],
	stream *connect.ServerStream[protov1.ReceiveFromLLMResponse],
) error {
	return connect.NewError(
		connect.CodeUnimplemented,
		errors.New("v1.MDService.ReceiveFromLLM is not implemented"),
	)
}
