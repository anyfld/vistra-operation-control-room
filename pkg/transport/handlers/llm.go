package handlers

import (
	"context"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/usecase"
)

// LLMHandler はLLMServiceのハンドラです。
type LLMHandler struct {
	uc usecase.LLMInteractor
}

// NewLLMHandler は新しいLLMHandlerを作成します。
func NewLLMHandler(uc usecase.LLMInteractor) *LLMHandler {
	return &LLMHandler{uc: uc}
}

// Chat はLLMとのチャットメッセージを処理します。
func (h *LLMHandler) Chat(
	ctx context.Context,
	req *connect.Request[protov1.ChatRequest],
) (*connect.Response[protov1.ChatResponse], error) {
	res, err := h.uc.Chat(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
