package usecase

import (
	"context"
	"errors"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/infrastructure"
)

type CameraInteractor interface {
	RegisterCamera(ctx context.Context, req *protov1.RegisterCameraRequest) (*protov1.Camera, error)
	UnregisterCamera(ctx context.Context, cameraID string) (bool, error)
	UpdateCamera(ctx context.Context, req *protov1.UpdateCameraRequest) (*protov1.Camera, error)
	GetCamera(
		ctx context.Context,
		cameraID string,
	) (*protov1.Camera, *protov1.CameraConnection, *protov1.CameraCapabilities, error)
	ListCameras(ctx context.Context, req *protov1.ListCamerasRequest) ([]*protov1.Camera, error)
	SwitchCameraMode(ctx context.Context, cameraID string, mode protov1.CameraMode) (bool, error)
	UpdateCameraState(
		ctx context.Context,
		cameraID string,
		ptz *protov1.PTZParameters,
		status protov1.CameraStatus,
	) (bool, error)
	GetConnectionStatus(ctx context.Context, cameraID string) (protov1.CameraStatus, bool, error)
	GetAllConnectionStatuses(ctx context.Context, cameraIDs []string) (map[string]protov1.CameraStatus, error)
	CheckAndUpdateDisconnectedCameras(ctx context.Context) error
}

var ErrCameraNotFound = errors.New("camera not found")

type CameraUsecase struct {
	repo *infrastructure.CameraRepo
}

func NewCameraUsecase(repo *infrastructure.CameraRepo) *CameraUsecase {
	return &CameraUsecase{repo: repo}
}

func (u *CameraUsecase) RegisterCamera(
	ctx context.Context,
	req *protov1.RegisterCameraRequest,
) (*protov1.Camera, error) {
	return u.repo.RegisterCamera(req), nil
}

func (u *CameraUsecase) UnregisterCamera(
	ctx context.Context,
	cameraID string,
) (bool, error) {
	return u.repo.UnregisterCamera(cameraID), nil
}

func (u *CameraUsecase) UpdateCamera(
	ctx context.Context,
	req *protov1.UpdateCameraRequest,
) (*protov1.Camera, error) {
	camera := u.repo.UpdateCamera(req.GetCameraId(), req)
	if camera == nil {
		return nil, ErrCameraNotFound
	}

	return camera, nil
}

func (u *CameraUsecase) GetCamera(
	ctx context.Context,
	cameraID string,
) (*protov1.Camera, *protov1.CameraConnection, *protov1.CameraCapabilities, error) {
	camera := u.repo.GetCamera(cameraID)
	if camera == nil {
		return nil, nil, nil, nil
	}

	connection := u.repo.GetConnection(cameraID)
	capabilities := u.repo.GetCapabilities(cameraID)

	return camera, connection, capabilities, nil
}

func (u *CameraUsecase) ListCameras(
	ctx context.Context,
	req *protov1.ListCamerasRequest,
) ([]*protov1.Camera, error) {
	return u.repo.ListCameras(
		req.GetMasterMfId(),
		req.GetModeFilter(),
		req.GetStatusFilter(),
	), nil
}

func (u *CameraUsecase) SwitchCameraMode(
	ctx context.Context,
	cameraID string,
	mode protov1.CameraMode,
) (bool, error) {
	return u.repo.SwitchCameraMode(cameraID, mode), nil
}

func (u *CameraUsecase) UpdateCameraState(
	ctx context.Context,
	cameraID string,
	ptz *protov1.PTZParameters,
	status protov1.CameraStatus,
) (bool, error) {
	success := u.repo.UpdateCameraState(cameraID, ptz, status)
	if !success {
		return false, ErrCameraNotFound
	}

	return true, nil
}

func (u *CameraUsecase) GetConnectionStatus(
	ctx context.Context,
	cameraID string,
) (protov1.CameraStatus, bool, error) {
	status, ok := u.repo.GetConnectionStatus(cameraID)

	return status, ok, nil
}

func (u *CameraUsecase) GetAllConnectionStatuses(
	ctx context.Context,
	cameraIDs []string,
) (map[string]protov1.CameraStatus, error) {
	if len(cameraIDs) == 0 {
		return u.repo.GetAllConnectionStatuses(), nil
	}

	result := make(map[string]protov1.CameraStatus)

	for _, cameraID := range cameraIDs {
		if status, ok := u.repo.GetConnectionStatus(cameraID); ok {
			result[cameraID] = status
		}
	}

	return result, nil
}

func (u *CameraUsecase) CheckAndUpdateDisconnectedCameras(ctx context.Context) error {
	u.repo.CheckAndUpdateDisconnectedCameras()

	return nil
}
