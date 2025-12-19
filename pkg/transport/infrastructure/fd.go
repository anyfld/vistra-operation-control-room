package infrastructure

import (
	"fmt"
	"sync"
	"time"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
)

const (
	defaultConfidence     = 0.85
	defaultBoundingBoxX   = 0.2
	defaultBoundingBoxY   = 0.2
	defaultBoundingBoxW   = 0.3
	defaultBoundingBoxH   = 0.3
	defaultProcessingTime = 50
	panAdjustment         = 10.0
	tiltAdjustment        = 5.0
	zoomAdjustment        = 0.5
	ptzSpeed              = 0.5
	estimatedMoveTimeMs   = 1000
	executionTimeMs       = 100
)

type PTZCommandEvent struct {
	Command     *protov1.ControlCommand
	Result      *protov1.ControlCommandResult
	TimestampMs int64
}

type FDRepo struct {
	mu                         sync.RWMutex
	patternMatchingSessions    map[string]*PatternMatchingSession
	controlCommands            map[string]*protov1.ControlCommand
	cameraStates               map[string]*protov1.CameraState
	cinematographyInstructions map[string]*protov1.CinematographyInstruction
	lastPTZEvents              map[string]*PTZCommandEvent
	ptzSubscribers             map[string][]chan *PTZCommandEvent
	ptzSubscribersMu           sync.RWMutex
}

type PatternMatchingSession struct {
	SessionID      string
	CameraID       string
	TargetSubjects []*protov1.Subject
	IntervalMs     uint32
	CreatedAt      time.Time
}

func NewFDRepo() *FDRepo {
	return &FDRepo{
		mu:                         sync.RWMutex{},
		patternMatchingSessions:    make(map[string]*PatternMatchingSession),
		controlCommands:            make(map[string]*protov1.ControlCommand),
		cameraStates:               make(map[string]*protov1.CameraState),
		cinematographyInstructions: make(map[string]*protov1.CinematographyInstruction),
		lastPTZEvents:              make(map[string]*PTZCommandEvent),
		ptzSubscribers:             make(map[string][]chan *PTZCommandEvent),
		ptzSubscribersMu:           sync.RWMutex{},
	}
}

func (r *FDRepo) ProcessImage(
	image *protov1.ImageData,
	targetSubjects []*protov1.Subject,
) ([]*protov1.DetectedSubject, uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	detected := make([]*protov1.DetectedSubject, 0, len(targetSubjects))

	for _, target := range targetSubjects {
		detected = append(detected, &protov1.DetectedSubject{
			Subject:    target,
			Confidence: defaultConfidence,
			DetectedBox: &protov1.BoundingBox{
				X:      defaultBoundingBoxX,
				Y:      defaultBoundingBoxY,
				Width:  defaultBoundingBoxW,
				Height: defaultBoundingBoxH,
			},
		})
	}

	return detected, defaultProcessingTime
}

func (r *FDRepo) StartPatternMatching(
	cameraID string,
	targetSubjects []*protov1.Subject,
	intervalMs uint32,
) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())
	r.patternMatchingSessions[sessionID] = &PatternMatchingSession{
		SessionID:      sessionID,
		CameraID:       cameraID,
		TargetSubjects: targetSubjects,
		IntervalMs:     intervalMs,
		CreatedAt:      time.Now(),
	}

	return sessionID
}

func (r *FDRepo) StopPatternMatching(sessionID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.patternMatchingSessions[sessionID]; !ok {
		return false
	}

	delete(r.patternMatchingSessions, sessionID)

	return true
}

func (r *FDRepo) GetPatternMatchingSession(sessionID string) *PatternMatchingSession {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.patternMatchingSessions[sessionID]
}

func (r *FDRepo) CalculateFraming(
	cameraID string,
	currentPtz *protov1.PTZParameters,
	targetShotType protov1.ShotType,
	targetSubjects []*protov1.DetectedSubject,
	targetAngle protov1.CameraAngle,
) (*protov1.PTZParameters, uint32, bool, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if currentPtz == nil {
		currentPtz = &protov1.PTZParameters{
			Pan:       0.0,
			Tilt:      0.0,
			Zoom:      1.0,
			PanSpeed:  0.0,
			TiltSpeed: 0.0,
			ZoomSpeed: 0.0,
		}
	}

	calculatedPtz := &protov1.PTZParameters{
		Pan:       currentPtz.GetPan() + panAdjustment,
		Tilt:      currentPtz.GetTilt() + tiltAdjustment,
		Zoom:      currentPtz.GetZoom() + zoomAdjustment,
		PanSpeed:  ptzSpeed,
		TiltSpeed: ptzSpeed,
		ZoomSpeed: ptzSpeed,
	}

	return calculatedPtz, estimatedMoveTimeMs, true, ""
}

func (r *FDRepo) SendControlCommand(command *protov1.ControlCommand) *protov1.ControlCommandResult {
	r.mu.Lock()

	commandID := command.GetCommandId()
	if commandID == "" {
		commandID = fmt.Sprintf("cmd-%d", time.Now().UnixNano())
		command.CommandId = commandID
	}

	r.controlCommands[commandID] = command

	resultingPtz := command.GetPtzParameters()
	if resultingPtz == nil {
		resultingPtz = &protov1.PTZParameters{
			Pan:       0.0,
			Tilt:      0.0,
			Zoom:      1.0,
			PanSpeed:  0.0,
			TiltSpeed: 0.0,
			ZoomSpeed: 0.0,
		}
	}

	result := &protov1.ControlCommandResult{
		CommandId:       commandID,
		Success:         true,
		ErrorMessage:    "",
		ResultingPtz:    resultingPtz,
		ExecutionTimeMs: executionTimeMs,
	}

	cameraID := command.GetCameraId()
	if cameraID != "" {
		r.lastPTZEvents[cameraID] = &PTZCommandEvent{
			Command:     command,
			Result:      result,
			TimestampMs: time.Now().UnixMilli(),
		}
	}

	r.mu.Unlock()

	if cameraID != "" {
		r.publishPTZCommand(cameraID, r.lastPTZEvents[cameraID])
	}

	return result
}

func (r *FDRepo) SubscribePTZCommands(cameraID string) <-chan *PTZCommandEvent {
	ch := make(chan *PTZCommandEvent, 100)

	r.ptzSubscribersMu.Lock()
	r.ptzSubscribers[cameraID] = append(r.ptzSubscribers[cameraID], ch)
	r.ptzSubscribersMu.Unlock()

	r.mu.RLock()
	if event, ok := r.lastPTZEvents[cameraID]; ok && event != nil {
		select {
		case ch <- event:
		default:
		}
	}
	r.mu.RUnlock()

	return ch
}

func (r *FDRepo) UnsubscribePTZCommands(cameraID string, ch <-chan *PTZCommandEvent) {
	r.ptzSubscribersMu.Lock()
	defer r.ptzSubscribersMu.Unlock()

	subscribers := r.ptzSubscribers[cameraID]
	for i, subscriber := range subscribers {
		if subscriber == ch {
			r.ptzSubscribers[cameraID] = append(subscribers[:i], subscribers[i+1:]...)
			close(subscriber)
			break
		}
	}

	if len(r.ptzSubscribers[cameraID]) == 0 {
		delete(r.ptzSubscribers, cameraID)
	}
}

func (r *FDRepo) publishPTZCommand(cameraID string, event *PTZCommandEvent) {
	r.ptzSubscribersMu.RLock()
	defer r.ptzSubscribersMu.RUnlock()

	subscribers := r.ptzSubscribers[cameraID]
	for _, ch := range subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (r *FDRepo) GetControlCommand(commandID string) *protov1.ControlCommand {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.controlCommands[commandID]
}

func (r *FDRepo) ReportCameraState(state *protov1.CameraState) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cameraStates[state.GetCameraId()] = state

	return true
}

func (r *FDRepo) GetCameraState(cameraID string) *protov1.CameraState {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.cameraStates[cameraID]
}

func (r *FDRepo) ExecuteCinematography(
	instruction *protov1.CinematographyInstruction,
) *protov1.CinematographyResult {
	r.mu.Lock()
	defer r.mu.Unlock()

	instructionID := instruction.GetInstructionId()
	if instructionID == "" {
		instructionID = fmt.Sprintf("instr-%d", time.Now().UnixNano())
		instruction.InstructionId = instructionID
	}

	r.cinematographyInstructions[instructionID] = instruction

	appliedPtz := instruction.GetPtzParameters()
	if appliedPtz == nil {
		appliedPtz = &protov1.PTZParameters{
			Pan:       0.0,
			Tilt:      0.0,
			Zoom:      1.0,
			PanSpeed:  0.0,
			TiltSpeed: 0.0,
			ZoomSpeed: 0.0,
		}
	}

	return &protov1.CinematographyResult{
		InstructionId: instructionID,
		CameraId:      instruction.GetCameraId(),
		Success:       true,
		ErrorMessage:  "",
		AppliedPtz:    appliedPtz,
		CompletedAtMs: time.Now().UnixMilli(),
	}
}

func (r *FDRepo) GetCinematographyInstruction(sourceFilter []string) *protov1.CinematographyInstruction {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, instr := range r.cinematographyInstructions {
		if len(sourceFilter) == 0 {
			return instr
		}

		for _, source := range sourceFilter {
			if instr.GetInstructionId() == source {
				return instr
			}
		}
	}

	return nil
}
