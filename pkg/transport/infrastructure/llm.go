package infrastructure

import (
	"context"
	"log/slog"

	"github.com/anyfld/vistra-operation-control-room/pkg/config"
	"github.com/anyfld/vistra-operation-control-room/pkg/llm/domain"
	"github.com/anyfld/vistra-operation-control-room/pkg/llm/infrastructure/gemini"
)

// LLMRepo はLLMサービスのリポジトリです。
type LLMRepo struct {
	geminiRepo domain.GeminiRepository
}

// NewLLMRepo は新しいLLMRepoを作成します。
func NewLLMRepo(ctx context.Context, logger *slog.Logger) (*LLMRepo, error) {
	cfg, err := config.LoadLLMConfig()
	if err != nil {
		return nil, err
	}

	geminiRepo, err := gemini.NewRepository(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}

	return &LLMRepo{
		geminiRepo: geminiRepo,
	}, nil
}

// GetGeminiRepository はGeminiRepositoryを取得します。
func (r *LLMRepo) GetGeminiRepository() domain.GeminiRepository {
	return r.geminiRepo
}
