package domain_test

import (
	"context"
	"testing"

	"github.com/anyfld/vistra-operation-control-room/pkg/llm/domain"
	"google.golang.org/genai"
)

const testFunctionName = "test_function"

func TestFunctionStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		function domain.Function
		validate func(*testing.T, domain.Function)
	}{
		{
			name: "function with declaration and handler",
			function: domain.Function{
				FunctionDeclaration: &genai.FunctionDeclaration{
					Name:        testFunctionName,
					Description: "A test function",
				},
				Function: func(ctx context.Context, request *genai.FunctionCall) (map[string]any, error) {
					return map[string]any{"result": "test"}, nil
				},
			},
			validate: func(t *testing.T, function domain.Function) {
				t.Helper()
				if function.FunctionDeclaration == nil {
					t.Error("Expected non-nil FunctionDeclaration")
				}
				if function.FunctionDeclaration.Name != testFunctionName {
					t.Errorf("Expected name '%s', got '%s'", testFunctionName, function.FunctionDeclaration.Name)
				}
				if function.Function == nil {
					t.Error("Expected non-nil Function handler")
				}
			},
		},
		{
			name: "function with complex declaration",
			function: domain.Function{
				FunctionDeclaration: &genai.FunctionDeclaration{
					Name:        "complex_function",
					Description: "A complex function with multiple params",
				},
				Function: func(ctx context.Context, request *genai.FunctionCall) (map[string]any, error) {
					return map[string]any{
						"status":  "success",
						"message": "Operation completed",
						"data": map[string]interface{}{
							"key": "value",
						},
					}, nil
				},
			},
			validate: func(t *testing.T, function domain.Function) {
				t.Helper()
				if function.FunctionDeclaration.Name != "complex_function" {
					t.Errorf("Expected name 'complex_function', got '%s'", function.FunctionDeclaration.Name)
				}
				ctx := t.Context()
				result, err := function.Function(ctx, &genai.FunctionCall{Name: "complex_function"})
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected non-nil result")
				}
				if status, ok := result["status"]; !ok || status != "success" {
					t.Error("Expected success status")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.validate(t, tc.function)
		})
	}
}

func TestGeminiRepositoryInterface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupCall   func() (domain.GeminiRepository, context.Context, string, []*domain.Content, string, []domain.Function)
		expectError bool
	}{
		{
			name: "interface is defined",
			setupCall: func() (domain.GeminiRepository, context.Context, string, []*domain.Content, string, []domain.Function) {
				return nil, t.Context(), "", nil, "", nil
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var _ domain.GeminiRepository
		})
	}
}

func TestContentTypeAlias(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func() *domain.Content
		validations func(*testing.T, *domain.Content)
	}{
		{
			name: "content with user role",
			setupFunc: func() *domain.Content {
				return &domain.Content{
					Role:  "user",
					Parts: []*genai.Part{{Text: "Hello"}},
				}
			},
			validations: func(t *testing.T, content *domain.Content) {
				t.Helper()
				if content.Role != "user" {
					t.Errorf("Expected role 'user', got '%s'", content.Role)
				}
				if len(content.Parts) != 1 {
					t.Errorf("Expected 1 part, got %d", len(content.Parts))
				}
				if content.Parts[0].Text != "Hello" {
					t.Errorf("Expected text 'Hello', got '%s'", content.Parts[0].Text)
				}
			},
		},
		{
			name: "content with model role",
			setupFunc: func() *domain.Content {
				return &domain.Content{
					Role:  "model",
					Parts: []*genai.Part{{Text: "Hi there!"}},
				}
			},
			validations: func(t *testing.T, content *domain.Content) {
				t.Helper()
				if content.Role != "model" {
					t.Errorf("Expected role 'model', got '%s'", content.Role)
				}
				if content.Parts[0].Text != "Hi there!" {
					t.Errorf("Expected text 'Hi there!', got '%s'", content.Parts[0].Text)
				}
			},
		},
		{
			name: "content with multiple parts",
			setupFunc: func() *domain.Content {
				return &domain.Content{
					Role: "user",
					Parts: []*genai.Part{
						{Text: "Part 1"},
						{Text: "Part 2"},
						{Text: "Part 3"},
					},
				}
			},
			validations: func(t *testing.T, content *domain.Content) {
				t.Helper()
				if len(content.Parts) != 3 {
					t.Errorf("Expected 3 parts, got %d", len(content.Parts))
				}
				for i, text := range []string{"Part 1", "Part 2", "Part 3"} {
					if content.Parts[i].Text != text {
						t.Errorf("Part %d: expected '%s', got '%s'", i, text, content.Parts[i].Text)
					}
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			content := tc.setupFunc()
			tc.validations(t, content)
		})
	}
}

func TestPartTypeAlias(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func() *domain.Part
		validations func(*testing.T, *domain.Part)
	}{
		{
			name: "part with text",
			setupFunc: func() *domain.Part {
				return &domain.Part{Text: "Hello, world!"}
			},
			validations: func(t *testing.T, part *domain.Part) {
				t.Helper()
				if part.Text != "Hello, world!" {
					t.Errorf("Expected text 'Hello, world!', got '%s'", part.Text)
				}
			},
		},
		{
			name: "part with empty text",
			setupFunc: func() *domain.Part {
				return &domain.Part{Text: ""}
			},
			validations: func(t *testing.T, part *domain.Part) {
				t.Helper()
				if part.Text != "" {
					t.Errorf("Expected empty text, got '%s'", part.Text)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			part := tc.setupFunc()
			tc.validations(t, part)
		})
	}
}

func TestFunctionCallTypeAlias(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func() *domain.FunctionCall
		validations func(*testing.T, *domain.FunctionCall)
	}{
		{
			name: "function call with name",
			setupFunc: func() *domain.FunctionCall {
				return &domain.FunctionCall{Name: "test_function"}
			},
			validations: func(t *testing.T, functionCall *domain.FunctionCall) {
				t.Helper()
				if functionCall.Name != "test_function" {
					t.Errorf("Expected name 'test_function', got '%s'", functionCall.Name)
				}
			},
		},
		{
			name: "function call with args",
			setupFunc: func() *domain.FunctionCall {
				return &domain.FunctionCall{
					Name: "test_function",
					Args: map[string]any{
						"param1": "value1",
						"param2": 42,
					},
				}
			},
			validations: func(t *testing.T, functionCall *domain.FunctionCall) {
				t.Helper()
				if functionCall.Name != "test_function" {
					t.Errorf("Expected name 'test_function', got '%s'", functionCall.Name)
				}
				if functionCall.Args == nil {
					t.Error("Expected non-nil Args")
				}
				if val, ok := functionCall.Args["param1"]; !ok || val != "value1" {
					t.Error("Expected param1 to be 'value1'")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fc := tc.setupFunc()
			tc.validations(t, fc)
		})
	}
}

func TestFunctionHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		handler      domain.Function
		expectResult bool
		expectError  bool
	}{
		{
			name: "successful function call",
			handler: domain.Function{
				FunctionDeclaration: &genai.FunctionDeclaration{
					Name:        "add",
					Description: "Add two numbers",
				},
				Function: func(ctx context.Context, request *genai.FunctionCall) (map[string]any, error) {
					return map[string]any{"result": 5}, nil
				},
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name: "context aware function",
			handler: domain.Function{
				FunctionDeclaration: &genai.FunctionDeclaration{
					Name:        "context_check",
					Description: "Check context",
				},
				Function: func(ctx context.Context, request *genai.FunctionCall) (map[string]any, error) {
					if ctx == nil {
						return nil, context.Canceled
					}

					return map[string]any{"status": "ok"}, nil
				},
			},
			expectResult: true,
			expectError:  false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			result, err := testCase.handler.Function(ctx, &genai.FunctionCall{Name: testCase.handler.FunctionDeclaration.Name})

			if testCase.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !testCase.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if testCase.expectResult && result == nil {
				t.Error("Expected non-nil result")
			}
		})
	}
}
