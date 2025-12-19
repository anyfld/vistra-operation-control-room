package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/infrastructure"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/usecase"
	"google.golang.org/protobuf/encoding/protojson"
)

type FDFallbackHandler struct {
	uc       usecase.FDInteractor
	cameraUC usecase.CameraInteractor
}

func NewFDFallbackHandler(uc usecase.FDInteractor, cameraUC usecase.CameraInteractor) *FDFallbackHandler {
	return &FDFallbackHandler{
		uc:       uc,
		cameraUC: cameraUC,
	}
}

func (h *FDHandler) ReportCameraStateHTTP() http.Handler {
	fallback := NewFDFallbackHandler(h.uc, h.cameraUC)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)

			return
		}

		defer func() {
			if closeErr := request.Body.Close(); closeErr != nil {
				_ = closeErr
			}
		}()

		body, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, "failed to read request body", http.StatusBadRequest)

			return
		}

		state := &protov1.CameraState{} //nolint:exhaustruct
		if err := protojson.Unmarshal(body, state); err != nil {
			http.Error(writer, "invalid request body", http.StatusBadRequest)

			return
		}

		if state.GetCameraId() == "" {
			http.Error(writer, "camera_id is required", http.StatusBadRequest)

			return
		}

		if _, err := fallback.uc.ReportCameraState(request.Context(), state); err != nil {
			http.Error(writer, "failed to report camera state", http.StatusInternalServerError)

			return
		}

		if fallback.cameraUC != nil {
			_, err := fallback.cameraUC.UpdateCameraState(
				request.Context(),
				state.GetCameraId(),
				state.GetCurrentPtz(),
				state.GetStatus(),
			)
			if err != nil {
				http.Error(writer, "failed to update camera state", http.StatusInternalServerError)

				return
			}
		}

		writer.WriteHeader(http.StatusNoContent)
	})
}

func (h *FDHandler) SendControlCommandHTTP() http.Handler {
	fallback := NewFDFallbackHandler(h.uc, h.cameraUC)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)

			return
		}

		defer func() {
			if closeErr := request.Body.Close(); closeErr != nil {
				_ = closeErr
			}
		}()

		body, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, "failed to read request body", http.StatusBadRequest)

			return
		}

		command := &protov1.ControlCommand{} //nolint:exhaustruct
		if err := protojson.Unmarshal(body, command); err != nil {
			http.Error(writer, "invalid request body", http.StatusBadRequest)

			return
		}

		if command.GetCameraId() == "" {
			http.Error(writer, "camera_id is required", http.StatusBadRequest)

			return
		}

		result, err := fallback.uc.SendControlCommand(request.Context(), command)
		if err != nil {
			http.Error(writer, "failed to send control command", http.StatusInternalServerError)

			return
		}

		writer.Header().Set("Content-Type", "application/json")

		encoded, err := protojson.MarshalOptions{ //nolint:exhaustruct
			UseProtoNames: true,
		}.Marshal(result)
		if err != nil {
			http.Error(writer, "failed to encode response", http.StatusInternalServerError)

			return
		}

		if _, err := writer.Write(encoded); err != nil {
			http.Error(writer, "failed to write response", http.StatusInternalServerError)
		}
	})
}

type pollControlCommandsResponse struct {
	Command     *protov1.ControlCommand       `json:"command,omitempty"`
	Result      *protov1.ControlCommandResult `json:"result,omitempty"`
	TimestampMs int64                         `json:"timestampMs,omitempty"`
}

func (h *FDHandler) PollControlCommandsHTTP() http.Handler {
	fallback := NewFDFallbackHandler(h.uc, h.cameraUC)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !h.validatePollRequest(writer, request) {
			return
		}

		cameraID := request.URL.Query().Get("camera_id")
		timeoutMs := h.parseTimeoutMs(request)

		ctx := request.Context()

		commandCh, err := fallback.uc.SubscribePTZCommands(ctx, cameraID)
		if err != nil {
			http.Error(writer, "failed to subscribe to commands", http.StatusInternalServerError)

			return
		}

		defer func() {
			unsubscribeErr := fallback.uc.UnsubscribePTZCommands(ctx, cameraID, commandCh)
			if unsubscribeErr != nil {
				_ = unsubscribeErr
			}
		}()

		h.handlePollResponse(writer, request, commandCh, timeoutMs)
	})
}

func (h *FDHandler) validatePollRequest(
	writer http.ResponseWriter,
	request *http.Request,
) bool {
	if request.Method != http.MethodGet {
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)

		return false
	}

	cameraID := request.URL.Query().Get("camera_id")
	if cameraID == "" {
		http.Error(writer, "camera_id is required", http.StatusBadRequest)

		return false
	}

	return true
}

func (h *FDHandler) parseTimeoutMs(request *http.Request) int {
	const defaultTimeoutMs = 30000

	timeoutMsParam := request.URL.Query().Get("timeout_ms")
	if timeoutMsParam == "" {
		return defaultTimeoutMs
	}

	value, err := strconv.Atoi(timeoutMsParam)
	if err != nil || value <= 0 {
		return defaultTimeoutMs
	}

	return value
}

func (h *FDHandler) handlePollResponse(
	writer http.ResponseWriter,
	request *http.Request,
	commandCh <-chan *infrastructure.PTZCommandEvent,
	timeoutMs int,
) {
	timeout := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	defer timeout.Stop()

	select {
	case <-request.Context().Done():
		http.Error(writer, "request cancelled", http.StatusRequestTimeout)

		return
	case <-timeout.C:
		writer.WriteHeader(http.StatusNoContent)

		return
	case event, ok := <-commandCh:
		if !ok || event == nil || (event.Command == nil && event.Result == nil) {
			writer.WriteHeader(http.StatusNoContent)

			return
		}

		h.writePollResponse(writer, event)
	}
}

func (h *FDHandler) writePollResponse(
	writer http.ResponseWriter,
	event *infrastructure.PTZCommandEvent,
) {
	response := &pollControlCommandsResponse{
		Command:     event.Command,
		Result:      event.Result,
		TimestampMs: event.TimestampMs,
	}

	writer.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(writer).Encode(response); err != nil {
		http.Error(writer, "failed to write response", http.StatusInternalServerError)
	}
}
