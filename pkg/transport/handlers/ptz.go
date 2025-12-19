package handlers

import (
	"context"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/usecase"
)

// PTZHandler はPTZServiceのハンドラです。
type PTZHandler struct {
	uc usecase.PTZInteractor
}

// NewPTZHandler は新しいPTZHandlerを作成します。
func NewPTZHandler(uc usecase.PTZInteractor) *PTZHandler {
	return &PTZHandler{uc: uc}
}

// Polling はFDからのポーリングリクエストを処理します。
// FDは定期的（デフォルト500ms間隔）にこのエンドポイントを呼び出し、
// タスク完了時は即座に呼び出します。
func (h *PTZHandler) Polling(
	ctx context.Context,
	req *connect.Request[protov1.PollingRequest],
) (*connect.Response[protov1.PollingResponse], error) {
	res, err := h.uc.Polling(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

// SendPTZCommand はEPからのPTZ枠命令を受け付けます。
// PTZ命令は到着時にLayer 2（シネマティック枠）を全て破棄・中断します。
func (h *PTZHandler) SendPTZCommand(
	ctx context.Context,
	req *connect.Request[protov1.SendPTZCommandRequest],
) (*connect.Response[protov1.SendPTZCommandResponse], error) {
	res, err := h.uc.SendPTZCommand(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

// SendCinematicCommand はEPからのシネマティック枠命令を受け付けます。
// シネマティック枠はPTZ枠が空の時のみ実行されます。
func (h *PTZHandler) SendCinematicCommand(
	ctx context.Context,
	req *connect.Request[protov1.SendCinematicCommandRequest],
) (*connect.Response[protov1.SendCinematicCommandResponse], error) {
	res, err := h.uc.SendCinematicCommand(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

// GetQueueStatus はCRのキュー状態を取得します。
func (h *PTZHandler) GetQueueStatus(
	ctx context.Context,
	req *connect.Request[protov1.GetQueueStatusRequest],
) (*connect.Response[protov1.GetQueueStatusResponse], error) {
	res, err := h.uc.GetQueueStatus(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
