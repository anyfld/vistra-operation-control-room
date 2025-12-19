package infrastructure

import (
	"fmt"
	"sync"
	"time"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
)

type MDRepo struct {
	mu                         sync.RWMutex
	videoOutputs               map[string]*protov1.VideoOutput
	cinematographyInstructions map[string]*protov1.CinematographyInstruction
	llmRequests                map[string]*LLMRequest
}

type LLMRequest struct {
	RequestID string
	Prompt    string
	Context   *protov1.LLMContext
	CreatedAt time.Time
}

func NewMDRepo() *MDRepo {
	return &MDRepo{
		mu:                         sync.RWMutex{},
		videoOutputs:               make(map[string]*protov1.VideoOutput),
		cinematographyInstructions: make(map[string]*protov1.CinematographyInstruction),
		llmRequests:                make(map[string]*LLMRequest),
	}
}

func (r *MDRepo) ConfigureVideoOutput(config *protov1.VideoOutputConfig) *protov1.VideoOutput {
	r.mu.Lock()
	defer r.mu.Unlock()

	outputID := config.GetId()
	if outputID == "" {
		outputID = fmt.Sprintf("output-%d", time.Now().UnixNano())
	}

	output := &protov1.VideoOutput{
		Config:                config,
		Status:                protov1.VideoOutputStatus_VIDEO_OUTPUT_STATUS_IDLE,
		CurrentSourceCameraId: "",
		StreamingStartedAtMs:  0,
		BytesSent:             0,
		ErrorMessage:          "",
	}

	r.videoOutputs[outputID] = output

	return output
}

func (r *MDRepo) GetVideoOutput(outputID string) *protov1.VideoOutput {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.videoOutputs[outputID]
}

func (r *MDRepo) ListVideoOutputs(
	typeFilter []protov1.VideoOutputType,
	statusFilter []protov1.VideoOutputStatus,
) []*protov1.VideoOutput {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*protov1.VideoOutput

	for _, output := range r.videoOutputs {
		if len(typeFilter) > 0 {
			found := false
			for _, t := range typeFilter {
				if output.Config.GetType() == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if len(statusFilter) > 0 {
			found := false
			for _, s := range statusFilter {
				if output.Status == s {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		result = append(result, output)
	}

	return result
}

func (r *MDRepo) StartStreaming(outputID string, sourceCameraID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	output, ok := r.videoOutputs[outputID]
	if !ok {
		return false
	}

	output.Status = protov1.VideoOutputStatus_VIDEO_OUTPUT_STATUS_STREAMING
	output.CurrentSourceCameraId = sourceCameraID
	output.StreamingStartedAtMs = time.Now().UnixMilli()
	output.ErrorMessage = ""

	return true
}

func (r *MDRepo) StopStreaming(outputID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	output, ok := r.videoOutputs[outputID]
	if !ok {
		return false
	}

	output.Status = protov1.VideoOutputStatus_VIDEO_OUTPUT_STATUS_IDLE
	output.CurrentSourceCameraId = ""
	output.StreamingStartedAtMs = 0

	return true
}

func (r *MDRepo) SwitchSource(outputID string, newSourceCameraID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	output, ok := r.videoOutputs[outputID]
	if !ok {
		return false
	}

	output.CurrentSourceCameraId = newSourceCameraID

	return true
}

func (r *MDRepo) ReceiveCinematographyInstruction(
	instruction *protov1.CinematographyInstruction,
	source string,
) (*protov1.ReceiveCinematographyInstructionResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	instructionID := instruction.GetInstructionId()
	if instructionID == "" {
		instructionID = fmt.Sprintf("instr-%d", time.Now().UnixNano())
		instruction.InstructionId = instructionID
	}

	r.cinematographyInstructions[instructionID] = instruction

	return &protov1.ReceiveCinematographyInstructionResponse{
		Accepted:        true,
		InstructionId:   instructionID,
		RejectionReason: "",
	}, nil
}

func (r *MDRepo) GetCinematographyInstruction(cameraID string) *protov1.CinematographyInstruction {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, instr := range r.cinematographyInstructions {
		if instr.GetCameraId() == cameraID {
			return instr
		}
	}

	return nil
}

func (r *MDRepo) CreateLLMRequest(prompt string, context *protov1.LLMContext) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	requestID := fmt.Sprintf("llm-%d", time.Now().UnixNano())
	r.llmRequests[requestID] = &LLMRequest{
		RequestID: requestID,
		Prompt:    prompt,
		Context:   context,
		CreatedAt: time.Now(),
	}

	return requestID
}

func (r *MDRepo) GetLLMRequest(requestID string) *LLMRequest {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.llmRequests[requestID]
}
