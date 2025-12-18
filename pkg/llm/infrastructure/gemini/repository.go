package gemini

import (
	"context"
	"log/slog"

	"github.com/anyfld/vistra-operation-control-room/pkg/config"
	"github.com/anyfld/vistra-operation-control-room/pkg/llm/domain"
	"github.com/samber/lo"
	"google.golang.org/genai"
)

const (
	temperature = 1.0
)

type repositoryImpl struct {
	client *genai.Client
	logger *slog.Logger
	model  string
}

func NewRepository(ctx context.Context, cfg config.LLMConfig, logger *slog.Logger) (domain.GeminiRepository, error) {
	client, err := genai.NewClient(
		ctx,
		//nolint:exhaustruct
		&genai.ClientConfig{
			Project:  cfg.ProjectID,
			Location: cfg.Location,
			Backend:  genai.BackendVertexAI,
		},
	)
	if err != nil {
		return nil, err
	}

	return &repositoryImpl{
		client: client,
		logger: logger.With("component", "gemini_repository"),
		model:  cfg.ModelName,
	}, nil
}

func (r *repositoryImpl) SendChatMessage(
	ctx context.Context,
	systemPrompt string,
	history []*domain.Content,
	message string,
	functions []domain.Function,
) ([]*string, error) {
	chat, err := r.client.Chats.Create(ctx, r.model, r.createConfig(systemPrompt, functions), history)
	if err != nil {
		return nil, err
	}

	//nolint:exhaustruct
	res, err := chat.SendMessage(ctx, genai.Part{Text: message})
	if err != nil {
		return nil, err
	}

	messages := make([]*string, 0)

	for len(res.Candidates) != 0 && len(res.Candidates[0].Content.Parts) != 0 {
		candidate := res.Candidates[0]
		firstPart := candidate.Content.Parts[0]

		messages = append(messages, lo.ToPtr(firstPart.Text))

		functionResponses := r.handleFunctionCalls(ctx, candidate.Content.Parts, functions)
		if len(functionResponses) == 0 {
			break
		}

		res, err = chat.SendMessage(ctx, functionResponses...)
		if err != nil {
			r.logger.Error("Failed to send function response", "error", err)

			break
		}
	}

	return messages, nil
}

func (r *repositoryImpl) handleFunctionCalls(
	ctx context.Context,
	parts []*genai.Part,
	functions []domain.Function,
) []genai.Part {
	functionResponses := make([]genai.Part, 0, len(parts))

	for _, part := range parts {
		if part.FunctionCall == nil {
			continue
		}

		matched, found := lo.Find(functions, func(f domain.Function) bool {
			return f.FunctionDeclaration.Name == part.FunctionCall.Name
		})

		var result map[string]any

		if found {
			var err error

			result, err = matched.Function(ctx, part.FunctionCall)
			if err != nil {
				r.logger.Error("Failed to execute function", "error", err)
			}
		}

		//nolint:exhaustruct
		functionResponses = append(functionResponses, genai.Part{
			//nolint:exhaustruct
			FunctionResponse: &genai.FunctionResponse{Name: part.FunctionCall.Name, Response: result},
		})
	}

	return functionResponses
}

func (r *repositoryImpl) createConfig(systemPrompt string, functions []domain.Function) *genai.GenerateContentConfig {
	//nolint:exhaustruct
	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr[float32](temperature),
		SystemInstruction: &genai.Content{
			Role:  "system",
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
	}

	if len(functions) > 0 {
		functionDeclarations := make([]*genai.FunctionDeclaration, len(functions))
		for i, function := range functions {
			functionDeclarations[i] = function.FunctionDeclaration
		}

		config.Tools = []*genai.Tool{{FunctionDeclarations: functionDeclarations}}
	}

	return config
}
