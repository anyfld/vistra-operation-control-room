package handlers

import (
	"context"
	"errors"
	"log"
	"time"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/infrastructure"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/usecase"
)

const (
	fdPollingIntervalMs   = 500
	fdDefaultConfidence   = 0.85
	fdDefaultBoundingBoxX = 0.2
	fdDefaultBoundingBoxY = 0.2
	fdDefaultBoundingBoxW = 0.3
	fdDefaultBoundingBoxH = 0.3
)

type FDHandler struct {
	uc       usecase.FDInteractor
	cameraUC usecase.CameraInteractor
}

func NewFDHandler(uc usecase.FDInteractor, cameraUC usecase.CameraInteractor) *FDHandler {
	return &FDHandler{
		uc:       uc,
		cameraUC: cameraUC,
	}
}

func (h *FDHandler) ExecuteCinematography(
	ctx context.Context,
	req *connect.Request[protov1.ExecuteCinematographyRequest],
) (*connect.Response[protov1.ExecuteCinematographyResponse], error) {
	result, err := h.uc.ExecuteCinematography(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.ExecuteCinematographyResponse{
		Result: result,
	}), nil
}

func (h *FDHandler) StreamCinematographyInstructions(
	ctx context.Context,
	req *connect.Request[protov1.FDServiceStreamCinematographyInstructionsRequest],
	stream *connect.ServerStream[protov1.FDServiceStreamCinematographyInstructionsResponse],
) error {
	sourceFilter := req.Msg.GetSourceFilter()

	for range 5 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		instruction, err := h.uc.GetCinematographyInstruction(ctx, sourceFilter)
		if err != nil {
			return err
		}

		if instruction != nil {
			if err := stream.Send(&protov1.FDServiceStreamCinematographyInstructionsResponse{
				Instruction: instruction,
				Source:      "md",
				TimestampMs: time.Now().UnixMilli(),
			}); err != nil {
				return err
			}
		}

		time.Sleep(fdPollingIntervalMs * time.Millisecond)
	}

	return nil
}

func (h *FDHandler) ProcessImage(
	ctx context.Context,
	req *connect.Request[protov1.ProcessImageRequest],
) (*connect.Response[protov1.ProcessImageResponse], error) {
	detected, processingTime, err := h.uc.ProcessImage(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.ProcessImageResponse{
		DetectedSubjects: detected,
		ProcessingTimeMs: processingTime,
	}), nil
}

func (h *FDHandler) StartPatternMatching(
	ctx context.Context,
	req *connect.Request[protov1.StartPatternMatchingRequest],
) (*connect.Response[protov1.StartPatternMatchingResponse], error) {
	sessionID, err := h.uc.StartPatternMatching(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.StartPatternMatchingResponse{
		Success:   true,
		SessionId: sessionID,
	}), nil
}

func (h *FDHandler) StopPatternMatching(
	ctx context.Context,
	req *connect.Request[protov1.StopPatternMatchingRequest],
) (*connect.Response[protov1.StopPatternMatchingResponse], error) {
	success, err := h.uc.StopPatternMatching(ctx, req.Msg.GetSessionId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.StopPatternMatchingResponse{
		Success: success,
	}), nil
}

func (h *FDHandler) StreamPatternMatchResults(
	ctx context.Context,
	req *connect.Request[protov1.StreamPatternMatchResultsRequest],
	stream *connect.ServerStream[protov1.StreamPatternMatchResultsResponse],
) error {
	sessionID := req.Msg.GetSessionId()

	sessID, cameraID, targetSubjects, intervalMs, err := h.uc.GetPatternMatchingSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if sessID == "" {
		return connect.NewError(
			connect.CodeNotFound,
			errors.New("pattern matching session not found"),
		)
	}

	for range 5 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		detected := []*protov1.DetectedSubject{}
		for _, subject := range targetSubjects {
			detected = append(detected, &protov1.DetectedSubject{
				Subject:    subject,
				Confidence: fdDefaultConfidence,
				DetectedBox: &protov1.BoundingBox{
					X:      fdDefaultBoundingBoxX,
					Y:      fdDefaultBoundingBoxY,
					Width:  fdDefaultBoundingBoxW,
					Height: fdDefaultBoundingBoxH,
				},
			})
		}

		if err := stream.Send(&protov1.StreamPatternMatchResultsResponse{
			SessionId:        sessionID,
			CameraId:         cameraID,
			DetectedSubjects: detected,
			TimestampMs:      time.Now().UnixMilli(),
		}); err != nil {
			return err
		}

		time.Sleep(time.Duration(intervalMs) * time.Millisecond)
	}

	return nil
}

func (h *FDHandler) CalculateFraming(
	ctx context.Context,
	req *connect.Request[protov1.CalculateFramingRequest],
) (*connect.Response[protov1.CalculateFramingResponse], error) {
	ptz, timeMs, success, errMsg, err := h.uc.CalculateFraming(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.CalculateFramingResponse{
		CalculatedPtz:       ptz,
		EstimatedMoveTimeMs: timeMs,
		Success:             success,
		ErrorMessage:        errMsg,
	}), nil
}

// StreamControlCommands は廃止されました。
// 新しいPTZServiceのPollingとSendPTZCommandを使用してください。
func (h *FDHandler) StreamControlCommands(
	ctx context.Context,
	req *connect.Request[protov1.StreamControlCommandsRequest],
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New(
			"StreamControlCommands is deprecated. "+
				"Please use PTZService.Polling and PTZService.SendPTZCommand instead",
		),
	)
}

func (h *FDHandler) handleInitMessage(
	ctx context.Context,
	init *protov1.StreamControlCommandsInit,
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	cameraID := init.GetCameraId()
	if cameraID == "" {
		log.Printf(
			"fd stream control commands init failed: missing camera_id",
		)

		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("camera_id is required"),
		)
	}

	log.Printf(
		"fd stream control commands init: camera_id=%s",
		cameraID,
	)

	commandCh, err := h.uc.SubscribePTZCommands(ctx, cameraID)
	if err != nil {
		return nil, err
	}

	defer func() {
		if unsubscribeErr := h.uc.UnsubscribePTZCommands(ctx, cameraID, commandCh); unsubscribeErr != nil {
			_ = unsubscribeErr
		}
	}()

	return h.waitForPTZEvent(ctx, cameraID, commandCh)
}

func (h *FDHandler) waitForPTZEvent(
	ctx context.Context,
	cameraID string,
	commandCh <-chan *infrastructure.PTZCommandEvent,
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case event, ok := <-commandCh:
		if !ok || event == nil {
			return h.handleSubscriptionClosed(cameraID)
		}

		return h.handlePTZEvent(cameraID, event)
	case <-time.After(fdPollingIntervalMs * time.Millisecond):
		return h.handleNoEvents(cameraID)
	}
}

func (h *FDHandler) handleSubscriptionClosed(
	cameraID string,
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	log.Printf(
		"fd stream control commands subscription closed: camera_id=%s",
		cameraID,
	)

	return connect.NewResponse(&protov1.StreamControlCommandsResponse{ //nolint:exhaustruct
		Status: &protov1.StreamControlCommandsStatus{
			Connected: false,
			Message:   "subscription closed for camera: " + cameraID,
		},
		TimestampMs: time.Now().UnixMilli(),
		// Command, Result are optional
	}), nil
}

func (h *FDHandler) handlePTZEvent(
	cameraID string,
	event *infrastructure.PTZCommandEvent,
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	if event.Result != nil {
		return h.handleResultEvent(cameraID, event)
	}

	if event.Command != nil {
		return h.handleCommandEvent(cameraID, event)
	}

	return h.handleEmptyEvent(cameraID)
}

func (h *FDHandler) handleResultEvent(
	cameraID string,
	event *infrastructure.PTZCommandEvent,
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	log.Printf(
		"fd stream control commands result event: camera_id=%s command_id=%s",
		cameraID,
		event.Result.GetCommandId(),
	)

	return connect.NewResponse(&protov1.StreamControlCommandsResponse{
		Command: event.Command,
		Result:  event.Result,
		Status: &protov1.StreamControlCommandsStatus{
			Connected: true,
			Message:   "command result event",
		},
		TimestampMs: event.TimestampMs,
	}), nil
}

func (h *FDHandler) handleCommandEvent(
	cameraID string,
	event *infrastructure.PTZCommandEvent,
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	log.Printf(
		"fd stream control commands command event: camera_id=%s command_id=%s",
		cameraID,
		event.Command.GetCommandId(),
	)

	return connect.NewResponse(&protov1.StreamControlCommandsResponse{ //nolint:exhaustruct
		Command: event.Command,
		Status: &protov1.StreamControlCommandsStatus{
			Connected: true,
			Message:   "command event",
		},
		TimestampMs: event.TimestampMs,
		// Result is optional
	}), nil
}

func (h *FDHandler) handleEmptyEvent(
	cameraID string,
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	log.Printf(
		"fd stream control commands event without payload: camera_id=%s",
		cameraID,
	)

	return connect.NewResponse(&protov1.StreamControlCommandsResponse{ //nolint:exhaustruct
		Status: &protov1.StreamControlCommandsStatus{
			Connected: true,
			Message:   "no PTZ command payload for camera: " + cameraID,
		},
		TimestampMs: time.Now().UnixMilli(),
		// Command, Result are optional
	}), nil
}

func (h *FDHandler) handleNoEvents(
	cameraID string,
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	return connect.NewResponse(&protov1.StreamControlCommandsResponse{ //nolint:exhaustruct
		Status: &protov1.StreamControlCommandsStatus{
			Connected: true,
			Message:   "no new PTZ events for camera: " + cameraID,
		},
		TimestampMs: time.Now().UnixMilli(),
		// Command, Result are optional
	}), nil
}

func (h *FDHandler) handleCommandMessage(
	ctx context.Context,
	command *protov1.ControlCommand,
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	log.Printf(
		"fd stream control commands command: camera_id=%s command_id=%s",
		command.GetCameraId(),
		command.GetCommandId(),
	)

	result, err := h.uc.SendControlCommand(ctx, command)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return connect.NewResponse(&protov1.StreamControlCommandsResponse{ //nolint:exhaustruct
			// Command, Result, Status, TimestampMs are optional for empty response
		}), nil
	}

	return connect.NewResponse(&protov1.StreamControlCommandsResponse{ //nolint:exhaustruct
		Result: result,
		Status: &protov1.StreamControlCommandsStatus{
			Connected: true,
			Message:   "command accepted",
		},
		TimestampMs: time.Now().UnixMilli(),
		// Command is optional
	}), nil
}

func (h *FDHandler) handleResultMessage() (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	log.Printf("fd stream control commands result received")

	return connect.NewResponse(&protov1.StreamControlCommandsResponse{ //nolint:exhaustruct
		Status: &protov1.StreamControlCommandsStatus{
			Connected: true,
			Message:   "command result received",
		},
		TimestampMs: time.Now().UnixMilli(),
		// Command, Result are optional
	}), nil
}

func (h *FDHandler) handleStateMessage(
	ctx context.Context,
	state *protov1.CameraState,
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	log.Printf(
		"fd stream control commands state: camera_id=%s",
		state.GetCameraId(),
	)

	_, err := h.uc.ReportCameraState(ctx, state)
	if err != nil {
		return nil, err
	}

	if h.cameraUC == nil {
		return h.createStateUpdateResponse()
	}

	return h.updateCameraState(ctx, state)
}

func (h *FDHandler) updateCameraState(
	ctx context.Context,
	state *protov1.CameraState,
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	success, err := h.cameraUC.UpdateCameraState(
		ctx,
		state.GetCameraId(),
		state.GetCurrentPtz(),
		state.GetStatus(),
	)
	if err != nil {
		if errors.Is(err, usecase.ErrCameraNotFound) {
			log.Printf(
				"fd stream control commands state failed: camera not found: camera_id=%s",
				state.GetCameraId(),
			)

			return nil, connect.NewError(
				connect.CodeNotFound,
				err,
			)
		}

		return nil, err
	}

	if !success {
		log.Printf(
			"fd stream control commands state failed: camera not found: camera_id=%s",
			state.GetCameraId(),
		)

		return nil, connect.NewError(
			connect.CodeNotFound,
			errors.New("camera not found"),
		)
	}

	return h.createStateUpdateResponse()
}

func (h *FDHandler) createStateUpdateResponse() (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	return connect.NewResponse(&protov1.StreamControlCommandsResponse{ //nolint:exhaustruct
		Status: &protov1.StreamControlCommandsStatus{
			Connected: true,
			Message:   "camera state updated",
		},
		TimestampMs: time.Now().UnixMilli(),
		// Command, Result are optional
	}), nil
}
