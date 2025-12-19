package usecase

import (
	"context"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/infrastructure"
)

type CRInteractor interface {
	RegisterMasterMF(ctx context.Context, req *protov1.RegisterMasterMFRequest) (*protov1.MasterMF, error)
	UnregisterMasterMF(ctx context.Context, id string) (bool, error)
	ListMasterMFs(ctx context.Context) ([]*protov1.MasterMF, error)
	GetMasterMF(ctx context.Context, id string) (*protov1.MasterMF, error)
	ListAllCameras(ctx context.Context) ([]*protov1.Camera, error)
	GetCamera(ctx context.Context, id string) (*protov1.Camera, error)
	PushConfiguration(ctx context.Context, cfg *protov1.Configuration, targetMasterMfIds []string) (bool, []string)
	GetConfiguration(ctx context.Context, masterMfId string) *protov1.Configuration
	SendCinematographyInstruction(ctx context.Context,
		req *protov1.SendCinematographyInstructionRequest,
	) (*protov1.SendCinematographyInstructionResponse, error)
}

type CRUsecase struct {
	repo *infrastructure.InMemoryRepo
}

func New(repo *infrastructure.InMemoryRepo) *CRUsecase {
	return &CRUsecase{repo: repo}
}

func (u *CRUsecase) RegisterMasterMF(
	ctx context.Context,
	req *protov1.RegisterMasterMFRequest,
) (*protov1.MasterMF, error) {
	return u.repo.RegisterMasterMF(req), nil
}

func (u *CRUsecase) UnregisterMasterMF(
	ctx context.Context,
	id string,
) (bool, error) {
	return u.repo.UnregisterMasterMF(id), nil
}

func (u *CRUsecase) ListMasterMFs(
	ctx context.Context,
) ([]*protov1.MasterMF, error) {
	return u.repo.ListMasterMFs(), nil
}

func (u *CRUsecase) GetMasterMF(
	ctx context.Context,
	id string,
) (*protov1.MasterMF, error) {
	return u.repo.GetMasterMF(id), nil
}

func (u *CRUsecase) ListAllCameras(
	ctx context.Context,
) ([]*protov1.Camera, error) {
	return u.repo.ListAllCameras(), nil
}

func (u *CRUsecase) GetCamera(
	ctx context.Context,
	id string,
) (*protov1.Camera, error) {
	return u.repo.GetCamera(id), nil
}

func (u *CRUsecase) PushConfiguration(
	ctx context.Context,
	cfg *protov1.Configuration,
	targetMasterMfIds []string,
) (bool, []string) {
	return u.repo.PushConfiguration(cfg, targetMasterMfIds)
}

func (u *CRUsecase) GetConfiguration(
	ctx context.Context,
	masterMfId string,
) *protov1.Configuration {
	return u.repo.GetConfiguration(masterMfId)
}

func (u *CRUsecase) SendCinematographyInstruction(
	ctx context.Context,
	req *protov1.SendCinematographyInstructionRequest,
) (*protov1.SendCinematographyInstructionResponse, error) {
	return u.repo.SendCinematographyInstruction(req), nil
}
