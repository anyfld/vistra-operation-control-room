package usecase

import (
	"context"
	"time"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/infrastructure"
)

// PTZInteractor はPTZサービスのユースケースインターフェースです。
type PTZInteractor interface {
	Polling(
		ctx context.Context,
		req *protov1.PollingRequest,
	) (*protov1.PollingResponse, error)
	SendPTZCommand(
		ctx context.Context,
		req *protov1.SendPTZCommandRequest,
	) (*protov1.SendPTZCommandResponse, error)
	SendCinematicCommand(
		ctx context.Context,
		req *protov1.SendCinematicCommandRequest,
	) (*protov1.SendCinematicCommandResponse, error)
	GetQueueStatus(
		ctx context.Context,
		req *protov1.GetQueueStatusRequest,
	) (*protov1.GetQueueStatusResponse, error)
}

// PTZUsecase はPTZサービスのユースケース実装です。
type PTZUsecase struct {
	repo *infrastructure.PTZRepo
}

// NewPTZUsecase は新しいPTZUsecaseを作成します。
func NewPTZUsecase(repo *infrastructure.PTZRepo) *PTZUsecase {
	return &PTZUsecase{repo: repo}
}

// Polling はFDからのポーリングリクエストを処理します。
func (u *PTZUsecase) Polling(
	ctx context.Context,
	req *protov1.PollingRequest,
) (*protov1.PollingResponse, error) {
	currentCommand, nextCommand, interrupt := u.repo.ProcessPolling(
		req.GetCameraId(),
		req.GetCompletedTaskId(),
		req.GetExecutingTaskId(),
		req.GetCurrentPtz(),
		req.GetDeviceStatus(),
		req.GetCameraStatus(),
	)

	return &protov1.PollingResponse{
		CurrentCommand: currentCommand,
		NextCommand:    nextCommand,
		Interrupt:      interrupt,
		TimestampMs:    time.Now().UnixMilli(),
	}, nil
}

// SendPTZCommand はEPからのPTZ命令を受け付けます。
func (u *PTZUsecase) SendPTZCommand(
	ctx context.Context,
	req *protov1.SendPTZCommandRequest,
) (*protov1.SendPTZCommandResponse, error) {
	cameraID := req.GetCameraId()
	command := req.GetCommand()

	if cameraID == "" {
		return &protov1.SendPTZCommandResponse{
			Accepted:     false,
			TaskId:       "",
			ErrorMessage: "camera_id is required",
		}, nil
	}

	if command == nil {
		return &protov1.SendPTZCommandResponse{
			Accepted:     false,
			TaskId:       "",
			ErrorMessage: "command is required",
		}, nil
	}

	taskID, accepted := u.repo.EnqueuePTZCommand(cameraID, command)

	return &protov1.SendPTZCommandResponse{
		Accepted:     accepted,
		TaskId:       taskID,
		ErrorMessage: "",
	}, nil
}

// SendCinematicCommand はEPからのシネマティック命令を受け付けます。
func (u *PTZUsecase) SendCinematicCommand(
	ctx context.Context,
	req *protov1.SendCinematicCommandRequest,
) (*protov1.SendCinematicCommandResponse, error) {
	cameraID := req.GetCameraId()
	command := req.GetCommand()

	if cameraID == "" {
		return &protov1.SendCinematicCommandResponse{
			Accepted:     false,
			TaskId:       "",
			ErrorMessage: "camera_id is required",
		}, nil
	}

	if command == nil {
		return &protov1.SendCinematicCommandResponse{
			Accepted:     false,
			TaskId:       "",
			ErrorMessage: "command is required",
		}, nil
	}

	taskID, accepted := u.repo.EnqueueCinematicCommand(cameraID, command)

	return &protov1.SendCinematicCommandResponse{
		Accepted:     accepted,
		TaskId:       taskID,
		ErrorMessage: "",
	}, nil
}

// GetQueueStatus はキュー状態を取得します。
func (u *PTZUsecase) GetQueueStatus(
	ctx context.Context,
	req *protov1.GetQueueStatusRequest,
) (*protov1.GetQueueStatusResponse, error) {
	cameraID := req.GetCameraId()

	var statuses []*protov1.CameraQueueStatus

	if cameraID != "" {
		status := u.repo.GetQueueStatus(cameraID)
		statuses = []*protov1.CameraQueueStatus{status}
	} else {
		statuses = u.repo.GetAllQueueStatuses()
	}

	return &protov1.GetQueueStatusResponse{
		CameraQueues: statuses,
	}, nil
}
