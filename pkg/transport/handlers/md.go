package handlers

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/usecase"
)

type MDHandler struct {
	uc usecase.MDInteractor
}

func NewMDHandler(uc usecase.MDInteractor) *MDHandler {
	return &MDHandler{uc: uc}
}

func (h *MDHandler) ReceiveCinematographyInstruction(
	ctx context.Context,
	req *connect.Request[protov1.ReceiveCinematographyInstructionRequest],
) (*connect.Response[protov1.ReceiveCinematographyInstructionResponse], error) {
	res, err := h.uc.ReceiveCinematographyInstruction(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (h *MDHandler) StreamCinematographyInstructions(
	ctx context.Context,
	req *connect.Request[protov1.MDServiceStreamCinematographyInstructionsRequest],
	stream *connect.ServerStream[protov1.MDServiceStreamCinematographyInstructionsResponse],
) error {
	cameraID := req.Msg.GetCameraId()

	for range 5 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		instruction, err := h.uc.GetCinematographyInstruction(ctx, cameraID)
		if err != nil {
			return err
		}

		if instruction != nil {
			if err := stream.Send(&protov1.MDServiceStreamCinematographyInstructionsResponse{
				Instruction: instruction,
				TimestampMs: time.Now().UnixMilli(),
			}); err != nil {
				return err
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func (h *MDHandler) ForwardToFD(
	ctx context.Context,
	req *connect.Request[protov1.ForwardToFDRequest],
) (*connect.Response[protov1.ForwardToFDResponse], error) {
	return connect.NewResponse(&protov1.ForwardToFDResponse{
		Success:      true,
		ErrorMessage: "",
		Result: &protov1.CinematographyResult{
			InstructionId: req.Msg.GetInstruction().GetInstructionId(),
			CameraId:      req.Msg.GetInstruction().GetCameraId(),
			Success:       true,
			ErrorMessage:  "",
			CompletedAtMs: time.Now().UnixMilli(),
		},
	}), nil
}

func (h *MDHandler) ConfigureVideoOutput(
	ctx context.Context,
	req *connect.Request[protov1.ConfigureVideoOutputRequest],
) (*connect.Response[protov1.ConfigureVideoOutputResponse], error) {
	output, err := h.uc.ConfigureVideoOutput(ctx, req.Msg.GetConfig())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.ConfigureVideoOutputResponse{
		Success: true,
		Output:  output,
	}), nil
}

func (h *MDHandler) GetVideoOutputStatus(
	ctx context.Context,
	req *connect.Request[protov1.GetVideoOutputStatusRequest],
) (*connect.Response[protov1.GetVideoOutputStatusResponse], error) {
	output, err := h.uc.GetVideoOutput(ctx, req.Msg.GetOutputId())
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, connect.NewError(
			connect.CodeNotFound,
			errors.New("video output not found"),
		)
	}

	return connect.NewResponse(&protov1.GetVideoOutputStatusResponse{
		Output: output,
	}), nil
}

func (h *MDHandler) ListVideoOutputs(
	ctx context.Context,
	req *connect.Request[protov1.ListVideoOutputsRequest],
) (*connect.Response[protov1.ListVideoOutputsResponse], error) {
	outputs, err := h.uc.ListVideoOutputs(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.ListVideoOutputsResponse{
		Outputs: outputs,
	}), nil
}

func (h *MDHandler) StartStreaming(
	ctx context.Context,
	req *connect.Request[protov1.StartStreamingRequest],
) (*connect.Response[protov1.StartStreamingResponse], error) {
	success, err := h.uc.StartStreaming(ctx, req.Msg.GetOutputId(), req.Msg.GetSourceCameraId())
	if err != nil {
		return nil, err
	}

	if !success {
		return connect.NewResponse(&protov1.StartStreamingResponse{
			Success:      false,
			ErrorMessage: "video output not found",
		}), nil
	}

	return connect.NewResponse(&protov1.StartStreamingResponse{
		Success:      true,
		ErrorMessage: "",
	}), nil
}

func (h *MDHandler) StopStreaming(
	ctx context.Context,
	req *connect.Request[protov1.StopStreamingRequest],
) (*connect.Response[protov1.StopStreamingResponse], error) {
	success, err := h.uc.StopStreaming(ctx, req.Msg.GetOutputId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.StopStreamingResponse{
		Success: success,
	}), nil
}

func (h *MDHandler) SwitchSource(
	ctx context.Context,
	req *connect.Request[protov1.SwitchSourceRequest],
) (*connect.Response[protov1.SwitchSourceResponse], error) {
	success, err := h.uc.SwitchSource(ctx, req.Msg.GetOutputId(), req.Msg.GetNewSourceCameraId())
	if err != nil {
		return nil, err
	}

	if !success {
		return connect.NewResponse(&protov1.SwitchSourceResponse{
			Success:      false,
			ErrorMessage: "video output not found",
		}), nil
	}

	return connect.NewResponse(&protov1.SwitchSourceResponse{
		Success:      true,
		ErrorMessage: "",
	}), nil
}

func (h *MDHandler) GetStreamingStatus(
	ctx context.Context,
	req *connect.Request[protov1.GetStreamingStatusRequest],
) (*connect.Response[protov1.GetStreamingStatusResponse], error) {
	outputs, err := h.uc.GetStreamingStatus(ctx, req.Msg.GetOutputId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.GetStreamingStatusResponse{
		Outputs: outputs,
	}), nil
}

func (h *MDHandler) StreamStreamingEvents(
	ctx context.Context,
	req *connect.Request[protov1.StreamStreamingEventsRequest],
	stream *connect.ServerStream[protov1.StreamStreamingEventsResponse],
) error {
	outputIDs := req.Msg.GetOutputIds()

	for range 5 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var outputs []*protov1.VideoOutput
		if len(outputIDs) == 0 {
			var err error
			outputs, err = h.uc.GetStreamingStatus(ctx, "")
			if err != nil {
				return err
			}
		} else {
			for _, outputID := range outputIDs {
				output, err := h.uc.GetVideoOutput(ctx, outputID)
				if err != nil {
					continue
				}
				if output != nil {
					outputs = append(outputs, output)
				}
			}
		}

		for _, output := range outputs {
			eventType := protov1.StreamingEventType_STREAMING_EVENT_TYPE_UNSPECIFIED
			if output.Status == protov1.VideoOutputStatus_VIDEO_OUTPUT_STATUS_STREAMING {
				eventType = protov1.StreamingEventType_STREAMING_EVENT_TYPE_STARTED
			}

			if err := stream.Send(&protov1.StreamStreamingEventsResponse{
				OutputId:    output.Config.GetId(),
				Type:        eventType,
				Output:      output,
				TimestampMs: time.Now().UnixMilli(),
				Details:     "",
			}); err != nil {
				return err
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func (h *MDHandler) SendToLLM(
	ctx context.Context,
	req *connect.Request[protov1.SendToLLMRequest],
) (*connect.Response[protov1.SendToLLMResponse], error) {
	requestID, err := h.uc.CreateLLMRequest(ctx, req.Msg.GetPrompt(), req.Msg.GetContext())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.SendToLLMResponse{
		Accepted:  true,
		RequestId: requestID,
	}), nil
}

func (h *MDHandler) ReceiveFromLLM(
	ctx context.Context,
	req *connect.Request[protov1.ReceiveFromLLMRequest],
	stream *connect.ServerStream[protov1.ReceiveFromLLMResponse],
) error {
	requestID := req.Msg.GetRequestId()

	for range 3 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		prompt, llmContext, err := h.uc.GetLLMRequest(ctx, requestID)
		if err != nil {
			return err
		}

		if prompt != "" {
			if err := stream.Send(&protov1.ReceiveFromLLMResponse{
				RequestId:    requestID,
				Text:         prompt,
				Instruction:  nil,
				IsComplete:   true,
				TimestampMs:  time.Now().UnixMilli(),
			}); err != nil {
				return err
			}
		}

		_ = llmContext

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}
