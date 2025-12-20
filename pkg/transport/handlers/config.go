package handlers

import (
	"context"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/pkg/config"
)

type ConfigHandler struct{}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

func (h *ConfigHandler) GetGlobalConfig(
	ctx context.Context,
	req *connect.Request[protov1.GetGlobalConfigRequest],
) (*connect.Response[protov1.GetGlobalConfigResponse], error) {
	cfg, err := config.GetGlobalConfig()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&protov1.GetGlobalConfigResponse{
		Config: &protov1.GlobalConfig{
			WebrtcUrlTemplate: cfg.WebrtcUrlTemplate,
		},
	}), nil
}
