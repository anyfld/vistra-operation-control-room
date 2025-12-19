package infrastructure

import (
	"fmt"
	"sync"
	"time"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
)

type InMemoryRepo struct {
	mu                   sync.RWMutex
	masterMfs            map[string]*protov1.MasterMF
	cameras              map[string]*protov1.Camera
	currentConfiguration *protov1.Configuration
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		mu:                   sync.RWMutex{},
		masterMfs:            make(map[string]*protov1.MasterMF),
		cameras:              make(map[string]*protov1.Camera),
		currentConfiguration: nil,
	}
}

func (r *InMemoryRepo) RegisterMasterMF(req *protov1.RegisterMasterMFRequest) *protov1.MasterMF {
	r.mu.Lock()
	defer r.mu.Unlock()

	masterID := fmt.Sprintf("mf-%d", time.Now().UnixNano())
	masterMF := &protov1.MasterMF{
		Id:                   masterID,
		Name:                 req.GetName(),
		IpAddress:            req.GetIpAddress(),
		Port:                 req.GetPort(),
		Status:               protov1.MasterMFStatus_MASTER_MF_STATUS_ONLINE,
		ConnectedCameraCount: 0,
		LastSeenAtMs:         time.Now().UnixMilli(),
		Metadata:             req.GetMetadata(),
	}
	r.masterMfs[masterID] = masterMF

	return masterMF
}

func (r *InMemoryRepo) UnregisterMasterMF(masterID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.masterMfs[masterID]; !ok {
		return false
	}

	delete(r.masterMfs, masterID)

	return true
}

func (r *InMemoryRepo) ListMasterMFs() []*protov1.MasterMF {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*protov1.MasterMF, 0, len(r.masterMfs))
	for _, v := range r.masterMfs {
		out = append(out, v)
	}

	return out
}

func (r *InMemoryRepo) GetMasterMF(id string) *protov1.MasterMF {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if v, ok := r.masterMfs[id]; ok {
		return v
	}

	return nil
}

func (r *InMemoryRepo) ListAllCameras() []*protov1.Camera {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*protov1.Camera, 0, len(r.cameras))
	for _, v := range r.cameras {
		out = append(out, v)
	}

	return out
}

func (r *InMemoryRepo) GetCamera(id string) *protov1.Camera {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if v, ok := r.cameras[id]; ok {
		return v
	}

	return nil
}

func (r *InMemoryRepo) PushConfiguration(cfg *protov1.Configuration, targetMasterMfIds []string) (bool, []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.currentConfiguration = &protov1.Configuration{
		Id:          cfg.GetId(),
		Version:     cfg.GetVersion(),
		ConfigJson:  cfg.GetConfigJson(),
		CreatedAtMs: time.Now().UnixMilli(),
	}

	return true, nil
}

func (r *InMemoryRepo) GetConfiguration(masterMfId string) *protov1.Configuration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.currentConfiguration
}

func (r *InMemoryRepo) SendCinematographyInstruction(
	req *protov1.SendCinematographyInstructionRequest,
) *protov1.SendCinematographyInstructionResponse {
	return &protov1.SendCinematographyInstructionResponse{
		Accepted:      true,
		InstructionId: fmt.Sprintf("instr-%d", time.Now().UnixNano()),
	}
}
