package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/gen/proto/v1/protov1connect"
)

const (
	defaultServerURL = "http://localhost:8080"
)

func main() {
	serverURL := flag.String("server", defaultServerURL, "Server URL")
	count := flag.Int("count", 3, "Number of sample cameras to register")
	flag.Parse()

	ctx := context.Background()
	client := protov1connect.NewCameraServiceClient(
		&http.Client{Timeout: 5 * time.Second},
		*serverURL,
	)

	sampleCameras := generateSampleCameras(*count)

	fmt.Printf("Registering %d sample cameras...\n\n", len(sampleCameras))

	for i, req := range sampleCameras {
		resp, err := client.RegisterCamera(ctx, connect.NewRequest(req))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to register camera %d: %v\n", i+1, err)
			continue
		}

		camera := resp.Msg.GetCamera()
		if camera == nil {
			fmt.Fprintf(os.Stderr, "Error: camera %d registration returned nil\n", i+1)
			continue
		}

		fmt.Printf("✓ Registered camera %d:\n", i+1)
		fmt.Printf("  ID: %s\n", camera.GetId())
		fmt.Printf("  Name: %s\n", camera.GetName())
		fmt.Printf("  Mode: %s\n", camera.GetMode().String())
		fmt.Printf("  Master MF ID: %s\n", camera.GetMasterMfId())
		fmt.Printf("  Status: %s\n", camera.GetStatus().String())
		fmt.Println()
	}

	fmt.Println("Sample camera registration completed!")
}

func generateSampleCameras(count int) []*protov1.RegisterCameraRequest {
	cameras := []*protov1.RegisterCameraRequest{
		{
			Name:       "Sample ONVIF Camera 1",
			Mode:       protov1.CameraMode_CAMERA_MODE_AUTONOMOUS,
			MasterMfId: "master-mf-001",
			Connection: &protov1.CameraConnection{
				Type:    protov1.ConnectionType_CONNECTION_TYPE_ONVIF,
				Address: "192.168.1.100",
				Port:    80,
				Credentials: &protov1.CameraCredentials{
					Username: "admin",
					Password: "password123",
				},
				Parameters: map[string]string{
					"profile": "main",
				},
			},
			Capabilities: &protov1.CameraCapabilities{
				SupportsPtz:         true,
				PanMin:              -180.0,
				PanMax:              180.0,
				TiltMin:             -90.0,
				TiltMax:             90.0,
				ZoomMin:             1.0,
				ZoomMax:             10.0,
				SupportedFramerates: []uint32{15, 30, 60},
				PresetCount:         16,
				SupportsAutofocus:   true,
				SupportsArm:         false,
				AdditionalFeatures:  []string{"night_vision", "motion_detection"},
			},
			Metadata: map[string]string{
				"location": "Studio A",
				"model":    "ONVIF-IP-Camera-Pro",
			},
		},
		{
			Name:       "Sample NDI Camera 1",
			Mode:       protov1.CameraMode_CAMERA_MODE_LIGHTWEIGHT,
			MasterMfId: "master-mf-001",
			Connection: &protov1.CameraConnection{
				Type:    protov1.ConnectionType_CONNECTION_TYPE_NDI,
				Address: "192.168.1.101",
				Port:    5960,
				Parameters: map[string]string{
					"source_name": "NDI-Camera-01",
				},
			},
			Capabilities: &protov1.CameraCapabilities{
				SupportsPtz:         false,
				SupportedFramerates: []uint32{30, 60},
				SupportsAutofocus:   true,
				SupportsArm:         false,
				AdditionalFeatures:  []string{"hdr", "low_latency"},
			},
			Metadata: map[string]string{
				"location": "Studio B",
				"model":    "NDI-Camera-HD",
			},
		},
		{
			Name:       "Sample RTSP Camera 1",
			Mode:       protov1.CameraMode_CAMERA_MODE_AUTONOMOUS,
			MasterMfId: "master-mf-002",
			Connection: &protov1.CameraConnection{
				Type:    protov1.ConnectionType_CONNECTION_TYPE_RTSP,
				Address: "192.168.1.102",
				Port:    554,
				Credentials: &protov1.CameraCredentials{
					Username: "user",
					Password: "pass",
				},
				Parameters: map[string]string{
					"stream": "main",
				},
			},
			Capabilities: &protov1.CameraCapabilities{
				SupportsPtz:         true,
				PanMin:              -170.0,
				PanMax:              170.0,
				TiltMin:             -45.0,
				TiltMax:             45.0,
				ZoomMin:             1.0,
				ZoomMax:             5.0,
				SupportedFramerates: []uint32{15, 30},
				PresetCount:         8,
				SupportsAutofocus:   false,
				SupportsArm:         true,
			},
			Metadata: map[string]string{
				"location": "Outdoor",
				"model":    "RTSP-Outdoor-Cam",
			},
		},
	}

	if count > len(cameras) {
		count = len(cameras)
	}

	return cameras[:count]
}
