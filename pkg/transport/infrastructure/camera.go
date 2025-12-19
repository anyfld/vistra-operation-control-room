package infrastructure

import (
	"fmt"
	"sync"
	"time"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
)

type CameraRepo struct {
	mu               sync.RWMutex
	cameras          map[string]*protov1.Camera
	connections      map[string]*protov1.CameraConnection
	capabilities     map[string]*protov1.CameraCapabilities
	connectionStatus map[string]protov1.CameraStatus
}

func NewCameraRepo() *CameraRepo {
	return &CameraRepo{
		mu:               sync.RWMutex{},
		cameras:          make(map[string]*protov1.Camera),
		connections:      make(map[string]*protov1.CameraConnection),
		capabilities:     make(map[string]*protov1.CameraCapabilities),
		connectionStatus: make(map[string]protov1.CameraStatus),
	}
}

func (r *CameraRepo) RegisterCamera(req *protov1.RegisterCameraRequest) *protov1.Camera {
	r.mu.Lock()
	defer r.mu.Unlock()

	cameraID := fmt.Sprintf("cam-%d", time.Now().UnixNano())
	camera := new(protov1.Camera)
	camera.Reset()

	camera.Id = cameraID
	camera.Name = req.GetName()
	camera.Mode = req.GetMode()
	camera.MasterMfId = req.GetMasterMfId()
	camera.Status = protov1.CameraStatus_CAMERA_STATUS_ONLINE
	camera.CurrentPtz = nil
	camera.LastSeenAtMs = time.Now().UnixMilli()
	camera.Metadata = req.GetMetadata()

	r.cameras[cameraID] = camera
	if req.GetConnection() != nil {
		r.connections[cameraID] = req.GetConnection()
	}

	if req.GetCapabilities() != nil {
		r.capabilities[cameraID] = req.GetCapabilities()
	}

	r.connectionStatus[cameraID] = protov1.CameraStatus_CAMERA_STATUS_ONLINE

	return camera
}

func (r *CameraRepo) UnregisterCamera(cameraID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.cameras[cameraID]; !ok {
		return false
	}

	delete(r.cameras, cameraID)
	delete(r.connections, cameraID)
	delete(r.capabilities, cameraID)
	delete(r.connectionStatus, cameraID)

	return true
}

func (r *CameraRepo) UpdateCamera(cameraID string, req *protov1.UpdateCameraRequest) *protov1.Camera {
	r.mu.Lock()
	defer r.mu.Unlock()

	camera, ok := r.cameras[cameraID]
	if !ok {
		return nil
	}

	if req.Name != nil {
		camera.Name = req.GetName()
	}

	if req.GetConnection() != nil {
		r.connections[cameraID] = req.GetConnection()
	}

	if req.Metadata != nil {
		camera.Metadata = req.GetMetadata()
	}

	return camera
}

func (r *CameraRepo) GetCamera(cameraID string) *protov1.Camera {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.cameras[cameraID]
}

func (r *CameraRepo) GetConnection(cameraID string) *protov1.CameraConnection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.connections[cameraID]
}

func (r *CameraRepo) GetCapabilities(cameraID string) *protov1.CameraCapabilities {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.capabilities[cameraID]
}

func (r *CameraRepo) ListCameras(
	masterMfId string,
	modeFilter []protov1.CameraMode,
	statusFilter []protov1.CameraStatus,
) []*protov1.Camera {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*protov1.Camera, 0, len(r.cameras))

	for _, camera := range r.cameras {
		if r.matchesFilters(camera, masterMfId, modeFilter, statusFilter) {
			result = append(result, camera)
		}
	}

	return result
}

func (r *CameraRepo) SwitchCameraMode(cameraID string, mode protov1.CameraMode) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	camera, ok := r.cameras[cameraID]
	if !ok {
		return false
	}

	camera.Mode = mode

	return true
}

func (r *CameraRepo) UpdateHeartbeat(cameraID string, ptz *protov1.PTZParameters, status protov1.CameraStatus) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	camera, ok := r.cameras[cameraID]
	if !ok {
		return false
	}

	camera.LastSeenAtMs = time.Now().UnixMilli()

	if status != protov1.CameraStatus_CAMERA_STATUS_UNSPECIFIED {
		camera.Status = status
		r.connectionStatus[cameraID] = status
	}

	if ptz != nil {
		camera.CurrentPtz = ptz
	}

	return true
}

func (r *CameraRepo) GetConnectionStatus(cameraID string) (protov1.CameraStatus, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status, ok := r.connectionStatus[cameraID]

	return status, ok
}

func (r *CameraRepo) GetAllConnectionStatuses() map[string]protov1.CameraStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]protov1.CameraStatus)
	for k, v := range r.connectionStatus {
		result[k] = v
	}

	return result
}

const (
	heartbeatTimeoutSeconds = 30
	millisecondsPerSecond   = 1000
)

func (r *CameraRepo) CheckAndUpdateDisconnectedCameras() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UnixMilli()
	timeoutMs := int64(heartbeatTimeoutSeconds * millisecondsPerSecond)

	for cameraID, camera := range r.cameras {
		if camera.GetLastSeenAtMs() == 0 {
			continue
		}

		elapsed := now - camera.GetLastSeenAtMs()
		if elapsed >= timeoutMs {
			if camera.GetStatus() != protov1.CameraStatus_CAMERA_STATUS_OFFLINE {
				camera.Status = protov1.CameraStatus_CAMERA_STATUS_OFFLINE
				r.connectionStatus[cameraID] = protov1.CameraStatus_CAMERA_STATUS_OFFLINE
			}
		}
	}
}

func (r *CameraRepo) matchesFilters(
	camera *protov1.Camera,
	masterMfId string,
	modeFilter []protov1.CameraMode,
	statusFilter []protov1.CameraStatus,
) bool {
	if masterMfId != "" && camera.GetMasterMfId() != masterMfId {
		return false
	}

	if len(modeFilter) > 0 && !r.matchesMode(camera.GetMode(), modeFilter) {
		return false
	}

	if len(statusFilter) > 0 && !r.matchesStatus(camera.GetStatus(), statusFilter) {
		return false
	}

	return true
}

func (r *CameraRepo) matchesMode(mode protov1.CameraMode, filter []protov1.CameraMode) bool {
	for _, f := range filter {
		if mode == f {
			return true
		}
	}

	return false
}

func (r *CameraRepo) matchesStatus(status protov1.CameraStatus, filter []protov1.CameraStatus) bool {
	for _, f := range filter {
		if status == f {
			return true
		}
	}

	return false
}
