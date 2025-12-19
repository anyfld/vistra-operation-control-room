package main

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/gen/proto/v1/protov1connect"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/infrastructure"
)

func newFDTestServer(t *testing.T) (*httptest.Server, protov1connect.FDServiceClient) {
	t.Helper()

	cameraRepo := infrastructure.NewCameraRepo()
	mux := http.NewServeMux()
	registerFDService(mux, cameraRepo)

	handler := h2c.NewHandler(mux, &http2.Server{})
	server := httptest.NewUnstartedServer(handler)
	server.EnableHTTP2 = false
	server.Start()

	transport := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}

	client := &http.Client{Transport: transport}
	fdClient := protov1connect.NewFDServiceClient(client, server.URL)

	return server, fdClient
}

func TestStreamControlCommands_PTZPubSub(t *testing.T) {
	t.Parallel()

	server, client := newFDTestServer(t)
	defer server.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	cameraID := "cam-e2e-1"

	initReq := connect.NewRequest(&protov1.StreamControlCommandsRequest{
		Message: &protov1.StreamControlCommandsRequest_Init{
			Init: &protov1.StreamControlCommandsInit{
				CameraId: cameraID,
			},
		},
	})

	resp, err := client.StreamControlCommands(ctx, initReq)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Msg)
	require.NotNil(t, resp.Msg.GetStatus())
	require.True(t, resp.Msg.GetStatus().GetConnected())

	command := &protov1.ControlCommand{
		CameraId: cameraID,
		Type: protov1.
			ControlCommandType_CONTROL_COMMAND_TYPE_PTZ_ABSOLUTE,
		PtzParameters: &protov1.PTZParameters{
			Pan:  10,
			Tilt: 5,
			Zoom: 2,
		},
	}

	commandReq := connect.NewRequest(&protov1.StreamControlCommandsRequest{
		Message: &protov1.StreamControlCommandsRequest_Command{
			Command: command,
		},
	})

	commandResp, err := client.StreamControlCommands(ctx, commandReq)
	require.NoError(t, err)
	require.NotNil(t, commandResp)
	require.NotNil(t, commandResp.Msg)
	require.NotNil(t, commandResp.Msg.GetResult())

	deadline := time.Now().Add(2 * time.Second)

	var gotResult *protov1.ControlCommandResult

	for time.Now().Before(deadline) {
		pollReq := connect.NewRequest(&protov1.StreamControlCommandsRequest{
			Message: &protov1.StreamControlCommandsRequest_Init{
				Init: &protov1.StreamControlCommandsInit{
					CameraId: cameraID,
				},
			},
		})

		pollResp, pollErr := client.StreamControlCommands(ctx, pollReq)
		require.NoError(t, pollErr)
		require.NotNil(t, pollResp)
		require.NotNil(t, pollResp.Msg)

		if cmd := pollResp.Msg.GetCommand(); cmd != nil {
			require.Equal(t, cameraID, cmd.GetCameraId())
			require.Equal(t, protov1.ControlCommandType_CONTROL_COMMAND_TYPE_PTZ_ABSOLUTE, cmd.GetType())
			require.NotNil(t, cmd.GetPtzParameters())
			require.InDelta(t, float64(10), float64(cmd.GetPtzParameters().GetPan()), 0.01)
			require.InDelta(t, float64(5), float64(cmd.GetPtzParameters().GetTilt()), 0.01)
			require.InDelta(t, float64(2), float64(cmd.GetPtzParameters().GetZoom()), 0.01)
		}

		if res := pollResp.Msg.GetResult(); res != nil {
			gotResult = res

			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	require.NotNil(t, gotResult)
	require.True(t, gotResult.GetSuccess())
	require.NotEmpty(t, gotResult.GetCommandId())
}
