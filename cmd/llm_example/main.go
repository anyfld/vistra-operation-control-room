package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/anyfld/vistra-operation-control-room/pkg/config"
	"github.com/anyfld/vistra-operation-control-room/pkg/llm/domain"
	"github.com/anyfld/vistra-operation-control-room/pkg/llm/infrastructure/gemini"
	"github.com/anyfld/vistra-operation-control-room/pkg/llm/usecase/interactor"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"google.golang.org/genai"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	app := fx.New(
		fx.Provide(newLogger),
		fx.Provide(newLLMConfig),
		fx.Provide(newGeminiRepository),
		fx.Provide(newLLMInteractor),
		fx.Provide(backgroundContext),
		fx.Invoke(runExamples),
	)

	ctx := context.Background()

	if err := app.Start(ctx); err != nil {
		log.Fatalf("Error starting app: %v", err)
	}

	if err := app.Stop(ctx); err != nil {
		log.Fatalf("Error stopping app: %v", err)
	}
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func newLLMConfig() (config.LLMConfig, error) {
	return config.LoadLLMConfig()
}

func newGeminiRepository(
	ctx context.Context,
	cfg config.LLMConfig,
	logger *slog.Logger,
) (domain.GeminiRepository, error) {
	return gemini.NewRepository(ctx, cfg, logger)
}

func newLLMInteractor(
	repo domain.GeminiRepository,
) interactor.LLMInteractor {
	return interactor.NewLLMInteractor(repo)
}

func backgroundContext() context.Context {
	return context.Background()
}

func runExamples(
	llmInteractor interactor.LLMInteractor,
	logger *slog.Logger,
) {
	ctx := context.Background()

	simpleMessage(ctx, logger, llmInteractor)
	messageWithHistory(ctx, logger, llmInteractor)
	messageWithCalculation(ctx, logger, llmInteractor)
}

func simpleMessage(
	ctx context.Context,
	logger *slog.Logger,
	llmInteractor interactor.LLMInteractor,
) {
	logger.Info("Example 1: Ask about the assistant")

	messages, err := llmInteractor.SendChatMessage(
		ctx,
		"You are a helpful assistant.",
		nil,
		"Who are you?",
		nil,
	)
	if err != nil {
		logger.Error("Failed to send message", "error", err)

		return
	}

	if len(messages) == 0 {
		logger.Error("No response received")

		return
	}

	logger.Info("Response", "message", *messages[0])
}

func messageWithHistory(
	ctx context.Context,
	logger *slog.Logger,
	llmInteractor interactor.LLMInteractor,
) {
	logger.Info("Example 2: Message with history")

	history := []*genai.Content{{
		Role:  "user",
		Parts: []*genai.Part{{Text: "My favorite color is blue."}},
	}}

	messages, err := llmInteractor.SendChatMessage(
		ctx,
		"You are a helpful assistant with good memory of conversations.",
		history,
		"What did I just tell you about my favorite color?",
		nil,
	)
	if err != nil {
		logger.Error("Failed to send message", "error", err)

		return
	}

	if len(messages) == 0 {
		logger.Error("No response received")

		return
	}

	logger.Info("Response", "message", *messages[0])
}

func messageWithCalculation(
	ctx context.Context,
	logger *slog.Logger,
	llmInteractor interactor.LLMInteractor,
) {
	logger.Info("Example 3: Message requesting calculation")

	messages, err := llmInteractor.SendChatMessage(
		ctx,
		"You are a math helper. Perform calculations when asked.",
		nil,
		"Calculate the area of a rectangle with width 5 and height 10",
		nil,
	)
	if err != nil {
		logger.Error("Failed to send message", "error", err)

		return
	}

	if len(messages) == 0 {
		logger.Error("No response received")

		return
	}

	logger.Info("Response", "message", *messages[0])
}
