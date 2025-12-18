package interactor

//go:generate go tool mockgen -package=mock -source=llm.go -destination=mock/llm.go

import (
	"context"

	"github.com/anyfld/vistra-operation-control-room/pkg/llm/domain"
)

type LLMInteractor interface {
	SendChatMessage(
		ctx context.Context,
		systemPrompt string,
		history []*domain.Content,
		message string,
		functions []domain.Function,
	) ([]*string, error)
}

type LLMInteractorImpl struct {
	geminiRepository domain.GeminiRepository
}

func NewLLMInteractor(
	geminiRepository domain.GeminiRepository,
) LLMInteractor {
	return &LLMInteractorImpl{
		geminiRepository: geminiRepository,
	}
}

func (l *LLMInteractorImpl) SendChatMessage(
	ctx context.Context,
	systemPrompt string,
	history []*domain.Content,
	message string,
	functions []domain.Function,
) ([]*string, error) {
	return l.geminiRepository.SendChatMessage(ctx, systemPrompt, history, message, functions)
}
