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

func newCameraTestServer(
	t *testing.T,
) (*httptest.Server, protov1connect.CameraServiceClient) {
	t.Helper()

	mux := http.NewServeMux()
	registerCameraService(mux)

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
	cameraClient := protov1connect.NewCameraServiceClient(client, server.URL)

	return server, cameraClient
}

func TestRegisterAndGetCameraE2E(t *testing.T) {
	t.Parallel()

	server, client := newCameraTestServer(t)
	defer server.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	registerReq := &protov1.RegisterCameraRequest{
		Name:       "e2e-camera-1",
		Mode:       protov1.CameraMode_CAMERA_MODE_AUTONOMOUS,
		MasterMfId: "mf-e2e-1",
		Connection: &protov1.CameraConnection{
			Type:    protov1.ConnectionType_CONNECTION_TYPE_ONVIF,
			Address: "192.0.2.1",
			Port:    80,
		},
		Capabilities: &protov1.CameraCapabilities{
			SupportsPtz: true,
			PanMin:      -180,
			PanMax:      180,
			TiltMin:     -90,
			TiltMax:     90,
			ZoomMin:     1,
			ZoomMax:     10,
		},
		Metadata: map[string]string{
			"location": "studio-a",
		},
	}

	registerResp, err := client.RegisterCamera(ctx, connect.NewRequest(registerReq))
	require.NoError(t, err)
	require.NotNil(t, registerResp)
	require.NotNil(t, registerResp.Msg)
	require.NotNil(t, registerResp.Msg.GetCamera())

	registered := registerResp.Msg.GetCamera()
	require.NotEmpty(t, registered.GetId())
	require.Equal(t, registerReq.GetName(), registered.GetName())
	require.Equal(t, registerReq.GetMode(), registered.GetMode())
	require.Equal(t, registerReq.GetMasterMfId(), registered.GetMasterMfId())
	require.Equal(
		t,
		protov1.CameraStatus_CAMERA_STATUS_ONLINE,
		registered.GetStatus(),
	)
	require.Equal(t, registerReq.GetMetadata(), registered.GetMetadata())

	getResp, err := client.GetCamera(
		ctx,
		connect.NewRequest(&protov1.GetCameraRequest{
			CameraId: registered.GetId(),
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, getResp)
	require.NotNil(t, getResp.Msg)
	require.NotNil(t, getResp.Msg.GetCamera())

	got := getResp.Msg.GetCamera()
	require.Equal(t, registered.GetId(), got.GetId())
	require.Equal(t, registerReq.GetName(), got.GetName())
	require.Equal(t, registerReq.GetMode(), got.GetMode())
	require.Equal(t, registerReq.GetMasterMfId(), got.GetMasterMfId())
	require.Equal(
		t,
		protov1.CameraStatus_CAMERA_STATUS_ONLINE,
		got.GetStatus(),
	)
	require.Equal(t, registerReq.GetMetadata(), got.GetMetadata())

	require.NotNil(t, getResp.Msg.GetConnection())
	require.Equal(
		t,
		registerReq.GetConnection().GetType(),
		getResp.Msg.GetConnection().GetType(),
	)
	require.Equal(
		t,
		registerReq.GetConnection().GetAddress(),
		getResp.Msg.GetConnection().GetAddress(),
	)
	require.Equal(
		t,
		registerReq.GetConnection().GetPort(),
		getResp.Msg.GetConnection().GetPort(),
	)

	require.NotNil(t, getResp.Msg.GetCapabilities())
	require.Equal(
		t,
		registerReq.GetCapabilities().GetSupportsPtz(),
		getResp.Msg.GetCapabilities().GetSupportsPtz(),
	)
	require.Equal(
		t,
		registerReq.GetCapabilities().GetPanMin(),
		getResp.Msg.GetCapabilities().GetPanMin(),
	)
	require.Equal(
		t,
		registerReq.GetCapabilities().GetPanMax(),
		getResp.Msg.GetCapabilities().GetPanMax(),
	)
	require.Equal(
		t,
		registerReq.GetCapabilities().GetTiltMin(),
		getResp.Msg.GetCapabilities().GetTiltMin(),
	)
	require.Equal(
		t,
		registerReq.GetCapabilities().GetTiltMax(),
		getResp.Msg.GetCapabilities().GetTiltMax(),
	)
	require.Equal(
		t,
		registerReq.GetCapabilities().GetZoomMin(),
		getResp.Msg.GetCapabilities().GetZoomMin(),
	)
	require.Equal(
		t,
		registerReq.GetCapabilities().GetZoomMax(),
		getResp.Msg.GetCapabilities().GetZoomMax(),
	)

	listResp, err := client.ListCameras(
		ctx,
		connect.NewRequest(&protov1.ListCamerasRequest{
			MasterMfId:   registerReq.GetMasterMfId(),
			ModeFilter:   []protov1.CameraMode{registerReq.GetMode()},
			StatusFilter: []protov1.CameraStatus{protov1.CameraStatus_CAMERA_STATUS_ONLINE},
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, listResp)
	require.NotNil(t, listResp.Msg)
	require.Len(t, listResp.Msg.GetCameras(), 1)
	require.Equal(
		t,
		registered.GetId(),
		listResp.Msg.GetCameras()[0].GetId(),
	)
}
