package handlers

import (
	"context"
	"errors"
	"log"
	"time"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
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

func (h *FDHandler) StreamControlCommands(
	ctx context.Context,
	req *connect.Request[protov1.StreamControlCommandsRequest],
) (*connect.Response[protov1.StreamControlCommandsResponse], error) {
	message := req.Msg
	if message == nil {
		return connect.NewResponse(&protov1.StreamControlCommandsResponse{}), nil
	}

	if init := message.GetInit(); init != nil {
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

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case event, ok := <-commandCh:
			if !ok || event == nil {
				log.Printf(
					"fd stream control commands subscription closed: camera_id=%s",
					cameraID,
				)

				return connect.NewResponse(&protov1.StreamControlCommandsResponse{
					Status: &protov1.StreamControlCommandsStatus{
						Connected: false,
						Message:   "subscription closed for camera: " + cameraID,
					},
					TimestampMs: time.Now().UnixMilli(),
				}), nil
			}

			if event.Result != nil {
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

			if event.Command != nil {
				log.Printf(
					"fd stream control commands command event: camera_id=%s command_id=%s",
					cameraID,
					event.Command.GetCommandId(),
				)

				return connect.NewResponse(&protov1.StreamControlCommandsResponse{
					Command: event.Command,
					Status: &protov1.StreamControlCommandsStatus{
						Connected: true,
						Message:   "command event",
					},
					TimestampMs: event.TimestampMs,
				}), nil
			}

			log.Printf(
				"fd stream control commands event without payload: camera_id=%s",
				cameraID,
			)

			return connect.NewResponse(&protov1.StreamControlCommandsResponse{
				Status: &protov1.StreamControlCommandsStatus{
					Connected: true,
					Message:   "no PTZ command payload for camera: " + cameraID,
				},
				TimestampMs: time.Now().UnixMilli(),
			}), nil
		case <-time.After(fdPollingIntervalMs * time.Millisecond):
			return connect.NewResponse(&protov1.StreamControlCommandsResponse{
				Status: &protov1.StreamControlCommandsStatus{
					Connected: true,
					Message:   "no new PTZ events for camera: " + cameraID,
				},
				TimestampMs: time.Now().UnixMilli(),
			}), nil
		}
	}

	if command := message.GetCommand(); command != nil {
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
			return connect.NewResponse(&protov1.StreamControlCommandsResponse{}), nil
		}

		return connect.NewResponse(&protov1.StreamControlCommandsResponse{
			Result: result,
			Status: &protov1.StreamControlCommandsStatus{
				Connected: true,
				Message:   "command accepted",
			},
			TimestampMs: time.Now().UnixMilli(),
		}), nil
	}

	if message.GetResult() != nil {
		log.Printf("fd stream control commands result received")

		return connect.NewResponse(&protov1.StreamControlCommandsResponse{
			Status: &protov1.StreamControlCommandsStatus{
				Connected: true,
				Message:   "command result received",
			},
			TimestampMs: time.Now().UnixMilli(),
		}), nil
	}

	if state := message.GetState(); state != nil {
		log.Printf(
			"fd stream control commands state: camera_id=%s",
			state.GetCameraId(),
		)

		_, err := h.uc.ReportCameraState(ctx, state)
		if err != nil {
			return nil, err
		}

		if h.cameraUC != nil {
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
		}

		return connect.NewResponse(&protov1.StreamControlCommandsResponse{
			Status: &protov1.StreamControlCommandsStatus{
				Connected: true,
				Message:   "camera state updated",
			},
			TimestampMs: time.Now().UnixMilli(),
		}), nil
	}

	return connect.NewResponse(&protov1.StreamControlCommandsResponse{
		Status: &protov1.StreamControlCommandsStatus{
			Connected: true,
			Message:   "no operation",
		},
		TimestampMs: time.Now().UnixMilli(),
	}), nil
}
