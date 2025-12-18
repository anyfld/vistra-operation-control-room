package interactor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/anyfld/vistra-operation-control-room/pkg/llm/domain"
	"github.com/anyfld/vistra-operation-control-room/pkg/llm/domain/mock"
	"github.com/anyfld/vistra-operation-control-room/pkg/llm/usecase/interactor"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"
	"google.golang.org/genai"
)

const (
	testSystemPrompt = "You are a helpful assistant"
	testHello        = "Hello"
)

func TestNewLLMInteractor(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockGeminiRepository(ctrl)
	llmInteractor := interactor.NewLLMInteractor(mockRepo)

	if llmInteractor == nil {
		t.Fatal("Expected non-nil interactor")
	}
}

func TestLLMInteractorImpl_SendChatMessage_SuccessWithHistory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		systemPrompt     string
		history          []*domain.Content
		message          string
		functions        []domain.Function
		expectedMessages []*string
		expectedError    error
	}{
		{
			name:         "simple message with history",
			systemPrompt: testSystemPrompt,
			history: []*domain.Content{
				{Role: "user", Parts: []*genai.Part{{Text: "Hello"}}},
				{Role: "model", Parts: []*genai.Part{{Text: "Hi there!"}}},
			},
			message:          "How are you?",
			functions:        []domain.Function{},
			expectedMessages: []*string{lo.ToPtr("I'm doing well, thank you!")},
			expectedError:    nil,
		},
		{
			name:         "message with multiple history entries",
			systemPrompt: testSystemPrompt,
			history: []*domain.Content{
				{Role: "user", Parts: []*genai.Part{{Text: "Hello"}}},
				{Role: "model", Parts: []*genai.Part{{Text: "Hi there!"}}},
				{Role: "user", Parts: []*genai.Part{{Text: "What's up?"}}},
				{Role: "model", Parts: []*genai.Part{{Text: "Not much!"}}},
			},
			message:          "Tell me more",
			functions:        []domain.Function{},
			expectedMessages: []*string{lo.ToPtr("Sure, I can tell you more.")},
			expectedError:    nil,
		},
		{
			name:         "custom system prompt",
			systemPrompt: "You are a technical support agent",
			history: []*domain.Content{
				{Role: "user", Parts: []*genai.Part{{Text: "My app crashes"}}},
			},
			message:          "What should I do?",
			functions:        []domain.Function{},
			expectedMessages: []*string{lo.ToPtr("Have you tried restarting the app?")},
			expectedError:    nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockGeminiRepository(ctrl)
			llmInteractor := interactor.NewLLMInteractor(mockRepo)

			ctx := t.Context()

			mockRepo.EXPECT().
				SendChatMessage(ctx, testCase.systemPrompt, testCase.history, testCase.message, testCase.functions).
				Return(testCase.expectedMessages, testCase.expectedError).
				Times(1)

			messages, err := llmInteractor.SendChatMessage(
				ctx, testCase.systemPrompt, testCase.history, testCase.message, testCase.functions)

			if !errors.Is(err, testCase.expectedError) {
				t.Fatalf("Expected error %v, got %v", testCase.expectedError, err)
			}

			if len(messages) != len(testCase.expectedMessages) {
				t.Fatalf("Expected %d messages, got %d", len(testCase.expectedMessages), len(messages))
			}

			if messages[0] == nil || *messages[0] != *testCase.expectedMessages[0] {
				t.Errorf("Expected message %v, got %v", *testCase.expectedMessages[0], *messages[0])
			}
		})
	}
}

func TestLLMInteractorImpl_SendChatMessage_WithFunctions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		functions        []domain.Function
		expectedMessages []*string
		expectedError    error
	}{
		{
			name: "single function",
			functions: []domain.Function{
				{
					FunctionDeclaration: &genai.FunctionDeclaration{
						Name:        "get_weather",
						Description: "Get current weather",
					},
					Function: func(ctx context.Context, request *genai.FunctionCall) (map[string]any, error) {
						return map[string]any{"temperature": 72}, nil
					},
				},
			},
			expectedMessages: []*string{lo.ToPtr("The weather is 72 degrees.")},
			expectedError:    nil,
		},
		{
			name: "multiple functions",
			functions: []domain.Function{
				{
					FunctionDeclaration: &genai.FunctionDeclaration{
						Name:        "get_weather",
						Description: "Get current weather",
					},
					Function: func(ctx context.Context, request *genai.FunctionCall) (map[string]any, error) {
						return map[string]any{"temperature": 72}, nil
					},
				},
				{
					FunctionDeclaration: &genai.FunctionDeclaration{
						Name:        "get_time",
						Description: "Get current time",
					},
					Function: func(ctx context.Context, request *genai.FunctionCall) (map[string]any, error) {
						return map[string]any{"time": "3:00 PM"}, nil
					},
				},
			},
			expectedMessages: []*string{lo.ToPtr("It is 3:00 PM and the weather is 72 degrees.")},
			expectedError:    nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockGeminiRepository(ctrl)
			llmInteractor := interactor.NewLLMInteractor(mockRepo)

			ctx := t.Context()
			history := []*domain.Content{}
			message := "What's the weather and time?"

			mockRepo.EXPECT().
				SendChatMessage(ctx, testSystemPrompt, history, message, testCase.functions).
				Return(testCase.expectedMessages, testCase.expectedError).
				Times(1)

			messages, err := llmInteractor.SendChatMessage(ctx, testSystemPrompt, history, message, testCase.functions)

			if !errors.Is(err, testCase.expectedError) {
				t.Fatalf("Expected error %v, got %v", testCase.expectedError, err)
			}

			if len(messages) != len(testCase.expectedMessages) {
				t.Fatalf("Expected %d messages, got %d", len(testCase.expectedMessages), len(messages))
			}

			if messages[0] == nil || *messages[0] != *testCase.expectedMessages[0] {
				t.Errorf("Expected message %v, got %v", *testCase.expectedMessages[0], *messages[0])
			}
		})
	}
}

func TestLLMInteractorImpl_SendChatMessage_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedError error
		setupMock     func(*mock.MockGeminiRepository, error)
	}{
		{
			name:          "repository error",
			expectedError: errors.New("failed to send message"),
			setupMock: func(mockRepo *mock.MockGeminiRepository, err error) {
				mockRepo.EXPECT().
					SendChatMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, err).
					Times(1)
			},
		},
		{
			name:          "context cancelled error",
			expectedError: errors.New("context cancelled"),
			setupMock: func(mockRepo *mock.MockGeminiRepository, err error) {
				mockRepo.EXPECT().
					SendChatMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, err).
					Times(1)
			},
		},
		{
			name:          "network error",
			expectedError: errors.New("network error"),
			setupMock: func(mockRepo *mock.MockGeminiRepository, err error) {
				mockRepo.EXPECT().
					SendChatMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, err).
					Times(1)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockGeminiRepository(ctrl)
			testCase.setupMock(mockRepo, testCase.expectedError)

			llmInteractor := interactor.NewLLMInteractor(mockRepo)

			ctx := t.Context()
			history := []*domain.Content{}
			message := testHello
			functions := []domain.Function{}

			messages, err := llmInteractor.SendChatMessage(ctx, testSystemPrompt, history, message, functions)

			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			if err.Error() != testCase.expectedError.Error() {
				t.Errorf("Expected error %v, got %v", testCase.expectedError, err)
			}

			if messages != nil {
				t.Error("Expected nil messages on error")
			}
		})
	}
}

func TestLLMInteractorImpl_SendChatMessage_EmptyHistory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		message          string
		expectedMessages []*string
	}{
		{
			name:             "first message",
			message:          "First message",
			expectedMessages: []*string{lo.ToPtr("Hello! How can I help you?")},
		},
		{
			name:             "greeting",
			message:          "Hi there",
			expectedMessages: []*string{lo.ToPtr("Hello! Nice to meet you.")},
		},
		{
			name:             "question",
			message:          "What can you do?",
			expectedMessages: []*string{lo.ToPtr("I can help you with various tasks.")},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockGeminiRepository(ctrl)
			llmInteractor := interactor.NewLLMInteractor(mockRepo)

			ctx := t.Context()
			history := []*domain.Content{}
			functions := []domain.Function{}

			mockRepo.EXPECT().
				SendChatMessage(ctx, testSystemPrompt, history, testCase.message, functions).
				Return(testCase.expectedMessages, nil).
				Times(1)

			messages, err := llmInteractor.SendChatMessage(ctx, testSystemPrompt, history, testCase.message, functions)

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if len(messages) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(messages))
			}

			if *messages[0] != *testCase.expectedMessages[0] {
				t.Errorf("Expected message %v, got %v", *testCase.expectedMessages[0], *messages[0])
			}
		})
	}
}

func TestLLMInteractorImpl_SendChatMessage_MultipleResponses(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockGeminiRepository(ctrl)
	llmInteractor := interactor.NewLLMInteractor(mockRepo)

	ctx := t.Context()
	history := []*domain.Content{}
	message := testHello
	functions := []domain.Function{}

	expectedMessages := []*string{
		lo.ToPtr("Response 1"),
		lo.ToPtr("Response 2"),
		lo.ToPtr("Response 3"),
	}

	mockRepo.EXPECT().
		SendChatMessage(ctx, testSystemPrompt, history, message, functions).
		Return(expectedMessages, nil).
		Times(1)

	messages, err := llmInteractor.SendChatMessage(ctx, testSystemPrompt, history, message, functions)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(messages) != len(expectedMessages) {
		t.Fatalf("Expected %d messages, got %d", len(expectedMessages), len(messages))
	}

	for i, msg := range messages {
		if *msg != *expectedMessages[i] {
			t.Errorf("Message %d: expected %v, got %v", i, *expectedMessages[i], *msg)
		}
	}
}

func TestLLMInteractorImpl_SendChatMessage_ContextHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectCanceled bool
	}{
		{
			name: "normal context",
			setupContext: func() context.Context {
				return t.Context()
			},
			expectCanceled: false,
		},
		{
			name: "cancelled context",
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()

				return ctx
			},
			expectCanceled: true,
		},
		{
			name: "timeout context",
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()

				return ctx
			},
			expectCanceled: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockGeminiRepository(ctrl)
			llmInteractor := interactor.NewLLMInteractor(mockRepo)

			ctx := testCase.setupContext()
			history := []*domain.Content{}
			message := testHello
			functions := []domain.Function{}

			mockRepo.EXPECT().
				SendChatMessage(ctx, testSystemPrompt, history, message, functions).
				Return(nil, context.Canceled).
				Times(1)

			messages, err := llmInteractor.SendChatMessage(ctx, testSystemPrompt, history, message, functions)

			if testCase.expectCanceled {
				if err == nil {
					t.Error("Expected error for cancelled context")
				}
			}

			if messages != nil && err != nil {
				t.Errorf("Got unexpected result: messages=%v, err=%v", messages, err)
			}
		})
	}
}

func TestLLMInteractorImpl_SendChatMessage_EmptyResponse(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockGeminiRepository(ctrl)
	llmInteractor := interactor.NewLLMInteractor(mockRepo)

	ctx := t.Context()
	history := []*domain.Content{}
	message := testHello
	functions := []domain.Function{}

	mockRepo.EXPECT().
		SendChatMessage(ctx, testSystemPrompt, history, message, functions).
		Return([]*string{}, nil).
		Times(1)

	messages, err := llmInteractor.SendChatMessage(ctx, testSystemPrompt, history, message, functions)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(messages) != 0 {
		t.Fatalf("Expected 0 messages, got %d", len(messages))
	}
}

func TestLLMInteractorImpl_SendChatMessage_VerifyMockCalls(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockGeminiRepository(ctrl)
	llmInteractor := interactor.NewLLMInteractor(mockRepo)

	ctx := t.Context()
	history := []*domain.Content{
		{Role: "user", Parts: []*genai.Part{{Text: "Hi"}}},
	}
	message := "How are you?"
	functions := []domain.Function{}
	expectedMessages := []*string{lo.ToPtr("I'm great!")}

	mockRepo.EXPECT().
		SendChatMessage(ctx, testSystemPrompt, history, message, functions).
		Return(expectedMessages, nil).
		Times(1)

	messages, err := llmInteractor.SendChatMessage(ctx, testSystemPrompt, history, message, functions)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if *messages[0] != *expectedMessages[0] {
		t.Errorf("Expected message %v, got %v", *expectedMessages[0], *messages[0])
	}
}
