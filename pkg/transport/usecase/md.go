package usecase

import (
	"context"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/infrastructure"
)

type MDInteractor interface {
	ReceiveCinematographyInstruction(
		ctx context.Context,
		req *protov1.ReceiveCinematographyInstructionRequest,
	) (*protov1.ReceiveCinematographyInstructionResponse, error)
	GetCinematographyInstruction(
		ctx context.Context,
		cameraID string,
	) (*protov1.CinematographyInstruction, error)
	ConfigureVideoOutput(
		ctx context.Context,
		config *protov1.VideoOutputConfig,
	) (*protov1.VideoOutput, error)
	GetVideoOutput(ctx context.Context, outputID string) (*protov1.VideoOutput, error)
	ListVideoOutputs(
		ctx context.Context,
		req *protov1.ListVideoOutputsRequest,
	) ([]*protov1.VideoOutput, error)
	StartStreaming(ctx context.Context, outputID string, sourceCameraID string) (bool, error)
	StopStreaming(ctx context.Context, outputID string) (bool, error)
	SwitchSource(
		ctx context.Context,
		outputID string,
		newSourceCameraID string,
	) (bool, error)
	GetStreamingStatus(ctx context.Context, outputID string) ([]*protov1.VideoOutput, error)
	CreateLLMRequest(
		ctx context.Context,
		prompt string,
		context *protov1.LLMContext,
	) (string, error)
	GetLLMRequest(ctx context.Context, requestID string) (string, *protov1.LLMContext, error)
}

type MDUsecase struct {
	repo *infrastructure.MDRepo
}

func NewMDUsecase(repo *infrastructure.MDRepo) *MDUsecase {
	return &MDUsecase{repo: repo}
}

func (u *MDUsecase) ReceiveCinematographyInstruction(
	ctx context.Context,
	req *protov1.ReceiveCinematographyInstructionRequest,
) (*protov1.ReceiveCinematographyInstructionResponse, error) {
	return u.repo.ReceiveCinematographyInstruction(req.GetInstruction(), req.GetSource())
}

func (u *MDUsecase) GetCinematographyInstruction(
	ctx context.Context,
	cameraID string,
) (*protov1.CinematographyInstruction, error) {
	return u.repo.GetCinematographyInstruction(cameraID), nil
}

func (u *MDUsecase) ConfigureVideoOutput(
	ctx context.Context,
	config *protov1.VideoOutputConfig,
) (*protov1.VideoOutput, error) {
	return u.repo.ConfigureVideoOutput(config), nil
}

func (u *MDUsecase) GetVideoOutput(
	ctx context.Context,
	outputID string,
) (*protov1.VideoOutput, error) {
	return u.repo.GetVideoOutput(outputID), nil
}

func (u *MDUsecase) ListVideoOutputs(
	ctx context.Context,
	req *protov1.ListVideoOutputsRequest,
) ([]*protov1.VideoOutput, error) {
	return u.repo.ListVideoOutputs(req.GetTypeFilter(), req.GetStatusFilter()), nil
}

func (u *MDUsecase) StartStreaming(
	ctx context.Context,
	outputID string,
	sourceCameraID string,
) (bool, error) {
	return u.repo.StartStreaming(outputID, sourceCameraID), nil
}

func (u *MDUsecase) StopStreaming(
	ctx context.Context,
	outputID string,
) (bool, error) {
	return u.repo.StopStreaming(outputID), nil
}

func (u *MDUsecase) SwitchSource(
	ctx context.Context,
	outputID string,
	newSourceCameraID string,
) (bool, error) {
	return u.repo.SwitchSource(outputID, newSourceCameraID), nil
}

func (u *MDUsecase) GetStreamingStatus(
	ctx context.Context,
	outputID string,
) ([]*protov1.VideoOutput, error) {
	if outputID == "" {
		return u.repo.ListVideoOutputs(nil, nil), nil
	}

	output := u.repo.GetVideoOutput(outputID)
	if output == nil {
		return []*protov1.VideoOutput{}, nil
	}

	return []*protov1.VideoOutput{output}, nil
}

func (u *MDUsecase) CreateLLMRequest(
	ctx context.Context,
	prompt string,
	context *protov1.LLMContext,
) (string, error) {
	return u.repo.CreateLLMRequest(prompt, context), nil
}

func (u *MDUsecase) GetLLMRequest(
	ctx context.Context,
	requestID string,
) (string, *protov1.LLMContext, error) {
	llmReq := u.repo.GetLLMRequest(requestID)
	if llmReq == nil {
		return "", nil, nil
	}

	return llmReq.Prompt, llmReq.Context, nil
}
