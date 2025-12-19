package usecase

import (
	"context"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/llm/domain"
	"github.com/anyfld/vistra-operation-control-room/pkg/llm/usecase/interactor"
	"google.golang.org/genai"
)

// LLMInteractor はLLMサービスのユースケースインターフェースです。
type LLMInteractor interface {
	Chat(
		ctx context.Context,
		req *protov1.ChatRequest,
	) (*protov1.ChatResponse, error)
}

// LLMUsecase はLLMサービスのユースケース実装です。
type LLMUsecase struct {
	llmInteractor interactor.LLMInteractor
}

// NewLLMUsecase は新しいLLMUsecaseを作成します。
func NewLLMUsecase(llmInteractor interactor.LLMInteractor) *LLMUsecase {
	return &LLMUsecase{llmInteractor: llmInteractor}
}

// Chat はLLMとのチャットメッセージを処理します。
func (u *LLMUsecase) Chat(
	ctx context.Context,
	req *protov1.ChatRequest,
) (*protov1.ChatResponse, error) {
	systemPrompt := "You are a helpful assistant."
	if req.GetSystemPrompt() != nil {
		systemPrompt = *req.SystemPrompt
	}

	history := make([]*domain.Content, 0, len(req.GetHistory()))
	for _, msg := range req.GetHistory() {
		history = append(history, &genai.Content{
			Role:  msg.GetRole(),
			Parts: []*genai.Part{{Text: msg.GetContent()}},
		})
	}

	messages, err := u.llmInteractor.SendChatMessage(
		ctx,
		systemPrompt,
		history,
		req.GetMessage(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return &protov1.ChatResponse{
			Message: "",
		}, nil
	}

	return &protov1.ChatResponse{
		Message: *messages[0],
	}, nil
}
