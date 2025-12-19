package handlers

import (
	"context"
	"errors"
	"fmt"
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

func (h *FDHandler) SendControlCommand(
	ctx context.Context,
	req *connect.Request[protov1.SendControlCommandRequest],
) (*connect.Response[protov1.SendControlCommandResponse], error) {
	result, err := h.uc.SendControlCommand(ctx, req.Msg.GetCommand())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&protov1.SendControlCommandResponse{
		Result: result,
	}), nil
}

func (h *FDHandler) StreamControlCommands(
	ctx context.Context,
	stream *connect.BidiStream[protov1.StreamControlCommandsRequest, protov1.StreamControlCommandsResponse],
) error {
	var cameraID string
	var commandCh <-chan *infrastructure.PTZCommandEvent

	for {
		req, err := stream.Receive()
		if err != nil {
			break
		}
		if req == nil {
			continue
		}

		if init := req.GetInit(); init != nil {
			cameraID = init.GetCameraId()
			if cameraID == "" {
				return connect.NewError(
					connect.CodeInvalidArgument,
					errors.New("camera_id is required"),
				)
			}

			var err error
			commandCh, err = h.uc.SubscribePTZCommands(ctx, cameraID)
			if err != nil {
				return err
			}

			if err := stream.Send(&protov1.StreamControlCommandsResponse{
				Message: &protov1.StreamControlCommandsResponse_Status{
					Status: &protov1.StreamControlCommandsStatus{
						Connected: true,
						Message:   fmt.Sprintf("Subscribed to PTZ commands for camera: %s", cameraID),
					},
				},
				TimestampMs: time.Now().UnixMilli(),
			}); err != nil {
				h.uc.UnsubscribePTZCommands(ctx, cameraID, commandCh)
				return err
			}

			go func() {
				defer func() {
					if commandCh != nil {
						h.uc.UnsubscribePTZCommands(ctx, cameraID, commandCh)
					}
				}()

				for {
					select {
					case <-ctx.Done():
						return
					case event, ok := <-commandCh:
						if !ok {
							return
						}

						if event.Command != nil {
							if err := stream.Send(&protov1.StreamControlCommandsResponse{
								Message: &protov1.StreamControlCommandsResponse_Command{
									Command: event.Command,
								},
								TimestampMs: event.TimestampMs,
							}); err != nil {
								return
							}
						}

						if event.Result != nil {
							if err := stream.Send(&protov1.StreamControlCommandsResponse{
								Message: &protov1.StreamControlCommandsResponse_Result{
									Result: event.Result,
								},
								TimestampMs: event.TimestampMs,
							}); err != nil {
								return
							}
						}
					}
				}
			}()
		} else if req.GetCommand() != nil {
			result, err := h.uc.SendControlCommand(ctx, req.GetCommand())
			if err != nil {
				return err
			}
			if result != nil {
				if err := stream.Send(&protov1.StreamControlCommandsResponse{
					Message: &protov1.StreamControlCommandsResponse_Result{
						Result: result,
					},
					TimestampMs: time.Now().UnixMilli(),
				}); err != nil {
					return err
				}
			}
		} else if req.GetResult() != nil {
		} else if req.GetState() != nil {
			state := req.GetState()
			_, err := h.uc.ReportCameraState(ctx, state)
			if err != nil {
				return err
			}

			if h.cameraUC != nil {
				_, err := h.cameraUC.UpdateCameraState(
					ctx,
					state.GetCameraId(),
					state.GetCurrentPtz(),
					state.GetStatus(),
				)
				if err != nil {
					return err
				}
			}
		}
	}

	if commandCh != nil {
		h.uc.UnsubscribePTZCommands(ctx, cameraID, commandCh)
	}

	return nil
}
