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
)

func newFDTestServer(t *testing.T) (*httptest.Server, protov1connect.FDServiceClient) {
	t.Helper()

	mux := http.NewServeMux()
	registerFDService(mux)

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream1 := client.StreamControlCommands(ctx)
	stream2 := client.StreamControlCommands(ctx)

	cameraID := "cam-e2e-1"

	initReq := &protov1.StreamControlCommandsRequest{
		Message: &protov1.StreamControlCommandsRequest_Init{
			Init: &protov1.StreamControlCommandsInit{
				CameraId: cameraID,
			},
		},
	}

	require.NoError(t, stream1.Send(initReq))
	require.NoError(t, stream2.Send(initReq))

	waitForStatus := func(
		stream *connect.BidiStreamForClient[
			protov1.StreamControlCommandsRequest,
			protov1.StreamControlCommandsResponse,
		],
	) *protov1.StreamControlCommandsStatus {
		t.Helper()

		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			msg, err := stream.Receive()
			if err != nil {
				return nil
			}
			if msg == nil {
				continue
			}

			if status := msg.GetStatus(); status != nil {
				return status
			}
		}

		return nil
	}

	status1 := waitForStatus(stream1)
	require.NotNil(t, status1)
	require.True(t, status1.GetConnected())

	status2 := waitForStatus(stream2)
	require.NotNil(t, status2)
	require.True(t, status2.GetConnected())

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

	sendReq := connect.NewRequest(&protov1.SendControlCommandRequest{
		Command: command,
	})
	_, err := client.SendControlCommand(ctx, sendReq)
	require.NoError(t, err)

	resultCh1 := make(chan *protov1.ControlCommandResult, 1)
	commandCh2 := make(chan *protov1.ControlCommand, 1)

	collectResults := func(
		stream *connect.BidiStreamForClient[
			protov1.StreamControlCommandsRequest,
			protov1.StreamControlCommandsResponse,
		],
		gotResult chan<- *protov1.ControlCommandResult,
		gotCommand chan<- *protov1.ControlCommand,
	) {
		for {
			msg, err := stream.Receive()
			if err != nil || msg == nil {
				return
			}

			if res := msg.GetResult(); res != nil && gotResult != nil {
				select {
				case gotResult <- res:
				default:
				}
			}

			if cmd := msg.GetCommand(); cmd != nil && gotCommand != nil {
				select {
				case gotCommand <- cmd:
				default:
				}
			}
		}
	}

	go collectResults(stream1, resultCh1, nil)
	go collectResults(stream2, nil, commandCh2)

	time.Sleep(100 * time.Millisecond)

	select {
	case cmd := <-commandCh2:
		require.NotNil(t, cmd)
		require.Equal(t, cameraID, cmd.GetCameraId())
		require.Equal(t, protov1.ControlCommandType_CONTROL_COMMAND_TYPE_PTZ_ABSOLUTE, cmd.GetType())
		require.NotNil(t, cmd.GetPtzParameters())
		require.Equal(t, float32(10), cmd.GetPtzParameters().GetPan())
		require.Equal(t, float32(5), cmd.GetPtzParameters().GetTilt())
		require.Equal(t, float32(2), cmd.GetPtzParameters().GetZoom())
	case <-ctx.Done():
		t.Fatal("timeout waiting for command on subscriber stream")
	}

	select {
	case res := <-resultCh1:
		require.NotNil(t, res)
		require.True(t, res.GetSuccess())
		require.NotEmpty(t, res.GetCommandId())
	case <-ctx.Done():
		t.Fatal("timeout waiting for result on subscriber stream")
	}
}
