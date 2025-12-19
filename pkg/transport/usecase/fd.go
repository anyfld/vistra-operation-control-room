package usecase

import (
	"context"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/infrastructure"
)

type FDInteractor interface {
	ExecuteCinematography(
		ctx context.Context,
		req *protov1.ExecuteCinematographyRequest,
	) (*protov1.CinematographyResult, error)
	GetCinematographyInstruction(
		ctx context.Context,
		sourceFilter []string,
	) (*protov1.CinematographyInstruction, error)
	ProcessImage(
		ctx context.Context,
		req *protov1.ProcessImageRequest,
	) ([]*protov1.DetectedSubject, uint32, error)
	StartPatternMatching(
		ctx context.Context,
		req *protov1.StartPatternMatchingRequest,
	) (string, error)
	StopPatternMatching(ctx context.Context, sessionID string) (bool, error)
	GetPatternMatchingSession(
		ctx context.Context,
		sessionID string,
	) (string, string, []*protov1.Subject, uint32, error)
	CalculateFraming(
		ctx context.Context,
		req *protov1.CalculateFramingRequest,
	) (*protov1.PTZParameters, uint32, bool, string, error)
	SendControlCommand(
		ctx context.Context,
		command *protov1.ControlCommand,
	) (*protov1.ControlCommandResult, error)
	GetControlCommand(
		ctx context.Context,
		cameraID string,
	) (*protov1.ControlCommand, error)
	ReportCameraState(
		ctx context.Context,
		state *protov1.CameraState,
	) (bool, error)
	GetCameraState(
		ctx context.Context,
		cameraID string,
	) (*protov1.CameraState, error)
	SubscribePTZCommands(
		ctx context.Context,
		cameraID string,
	) (<-chan *infrastructure.PTZCommandEvent, error)
	UnsubscribePTZCommands(
		ctx context.Context,
		cameraID string,
		ch <-chan *infrastructure.PTZCommandEvent,
	) error
}

func (u *FDUsecase) SubscribePTZCommands(
	ctx context.Context,
	cameraID string,
) (<-chan *infrastructure.PTZCommandEvent, error) {
	return u.repo.SubscribePTZCommands(cameraID), nil
}

func (u *FDUsecase) UnsubscribePTZCommands(
	ctx context.Context,
	cameraID string,
	ch <-chan *infrastructure.PTZCommandEvent,
) error {
	u.repo.UnsubscribePTZCommands(cameraID, ch)

	return nil
}

type FDUsecase struct {
	repo *infrastructure.FDRepo
}

func NewFDUsecase(repo *infrastructure.FDRepo) *FDUsecase {
	return &FDUsecase{repo: repo}
}

func (u *FDUsecase) ExecuteCinematography(
	ctx context.Context,
	req *protov1.ExecuteCinematographyRequest,
) (*protov1.CinematographyResult, error) {
	return u.repo.ExecuteCinematography(req.GetInstruction()), nil
}

func (u *FDUsecase) GetCinematographyInstruction(
	ctx context.Context,
	sourceFilter []string,
) (*protov1.CinematographyInstruction, error) {
	return u.repo.GetCinematographyInstruction(sourceFilter), nil
}

func (u *FDUsecase) ProcessImage(
	ctx context.Context,
	req *protov1.ProcessImageRequest,
) ([]*protov1.DetectedSubject, uint32, error) {
	detected, processingTime := u.repo.ProcessImage(
		req.GetImage(),
		req.GetTargetSubjects(),
	)

	return detected, processingTime, nil
}

func (u *FDUsecase) StartPatternMatching(
	ctx context.Context,
	req *protov1.StartPatternMatchingRequest,
) (string, error) {
	return u.repo.StartPatternMatching(
		req.GetCameraId(),
		req.GetTargetSubjects(),
		req.GetIntervalMs(),
	), nil
}

func (u *FDUsecase) StopPatternMatching(
	ctx context.Context,
	sessionID string,
) (bool, error) {
	return u.repo.StopPatternMatching(sessionID), nil
}

func (u *FDUsecase) GetPatternMatchingSession(
	ctx context.Context,
	sessionID string,
) (string, string, []*protov1.Subject, uint32, error) {
	session := u.repo.GetPatternMatchingSession(sessionID)
	if session == nil {
		return "", "", nil, 0, nil
	}

	return session.SessionID, session.CameraID, session.TargetSubjects, session.IntervalMs, nil
}

func (u *FDUsecase) CalculateFraming(
	ctx context.Context,
	req *protov1.CalculateFramingRequest,
) (*protov1.PTZParameters, uint32, bool, string, error) {
	ptz, timeMs, success, errMsg := u.repo.CalculateFraming(
		req.GetCameraId(),
		req.GetCurrentPtz(),
		req.GetTargetShotType(),
		req.GetTargetSubjects(),
		req.GetTargetAngle(),
	)

	return ptz, timeMs, success, errMsg, nil
}

func (u *FDUsecase) SendControlCommand(
	ctx context.Context,
	command *protov1.ControlCommand,
) (*protov1.ControlCommandResult, error) {
	return u.repo.SendControlCommand(command), nil
}

func (u *FDUsecase) GetControlCommand(
	ctx context.Context,
	cameraID string,
) (*protov1.ControlCommand, error) {
	return u.repo.GetControlCommand(cameraID), nil
}

func (u *FDUsecase) ReportCameraState(
	ctx context.Context,
	state *protov1.CameraState,
) (bool, error) {
	return u.repo.ReportCameraState(state), nil
}

func (u *FDUsecase) GetCameraState(
	ctx context.Context,
	cameraID string,
) (*protov1.CameraState, error) {
	return u.repo.GetCameraState(cameraID), nil
}
