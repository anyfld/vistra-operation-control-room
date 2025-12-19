package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/gen/proto/v1/protov1connect"
)

const (
	defaultServerURL     = "http://localhost:8080"
	httpClientTimeoutSec = 5
)

func main() {
	serverURL := flag.String("server", defaultServerURL, "Server URL")
	jsonOutput := flag.Bool("json", false, "Output in JSON format")
	flag.Parse()

	ctx := context.Background()
	client := protov1connect.NewCameraServiceClient(
		&http.Client{ //nolint:exhaustruct
			Timeout: httpClientTimeoutSec * time.Second,
		},
		*serverURL,
	)

	req := connect.NewRequest(&protov1.ListCamerasRequest{ //nolint:exhaustruct
		// MasterMfId, ModeFilter, StatusFilter, PageSize, PageToken are optional
	})

	resp, err := client.ListCameras(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list cameras: %v\n", err)
		os.Exit(1)
	}

	cameras := resp.Msg.GetCameras()
	if len(cameras) == 0 {
		_, _ = io.WriteString(os.Stdout, "No cameras registered.\n")

		return
	}

	if *jsonOutput {
		outputJSON(cameras, client, ctx)
	} else {
		outputTable(cameras, client, ctx)
	}
}

func outputJSON(
	cameras []*protov1.Camera,
	client protov1connect.CameraServiceClient,
	ctx context.Context,
) {
	type cameraInfo struct {
		Camera     *protov1.Camera           `json:"camera"`
		Connection *protov1.CameraConnection `json:"connection"`
		Status     protov1.CameraStatus      `json:"status"`
	}

	cameraInfos := make([]cameraInfo, 0, len(cameras))

	for _, camera := range cameras {
		info := cameraInfo{ //nolint:exhaustruct
			Camera: camera,
			Status: camera.GetStatus(),
			// Connection will be set below if available
		}

		req := connect.NewRequest(&protov1.GetCameraRequest{
			CameraId: camera.GetId(),
		})

		resp, err := client.GetCamera(ctx, req)
		if err == nil && resp != nil {
			info.Connection = resp.Msg.GetConnection()
		}

		cameraInfos = append(cameraInfos, info)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(cameraInfos); err != nil {
		_, _ = io.WriteString(os.Stderr, "Error: failed to encode JSON: "+err.Error()+"\n")
		os.Exit(1)
	}
}

const (
	maxNameLength       = 30
	maxMasterMfIdLength = 20
	separatorLength     = 120
)

func outputTable(
	cameras []*protov1.Camera,
	client protov1connect.CameraServiceClient,
	ctx context.Context,
) {
	_, _ = fmt.Fprintf(os.Stdout, "Registered Cameras (%d):\n\n", len(cameras))
	_, _ = fmt.Fprintf(os.Stdout, "%-20s %-30s %-20s %-15s %-20s %-15s\n",
		"ID", "Name", "Mode", "Status", "Master MF ID", "Last Seen")
	_, _ = io.WriteString(os.Stdout, strings.Repeat("-", separatorLength)+"\n")

	for _, camera := range cameras {
		outputCameraRow(camera, client, ctx)
	}
}

func outputCameraRow(
	camera *protov1.Camera,
	client protov1connect.CameraServiceClient,
	ctx context.Context,
) {
	lastSeen := formatTimestamp(camera.GetLastSeenAtMs())
	status := formatCameraStatus(camera.GetStatus())
	_, _ = fmt.Fprintf(os.Stdout, "%-20s %-30s %-20s %-15s %-20s %-15s\n",
		camera.GetId(),
		truncate(camera.GetName(), maxNameLength),
		formatCameraMode(camera.GetMode()),
		status,
		truncate(camera.GetMasterMfId(), maxMasterMfIdLength),
		lastSeen,
	)

	outputConnectionInfo(camera, client, ctx)
	outputPTZInfo(camera)
	outputMetadata(camera)

	_, _ = io.WriteString(os.Stdout, "\n")
}

func outputConnectionInfo(
	camera *protov1.Camera,
	client protov1connect.CameraServiceClient,
	ctx context.Context,
) {
	req := connect.NewRequest(&protov1.GetCameraRequest{
		CameraId: camera.GetId(),
	})

	resp, err := client.GetCamera(ctx, req)
	if err != nil || resp == nil {
		return
	}

	connection := resp.Msg.GetConnection()
	if connection == nil {
		return
	}

	_, _ = io.WriteString(os.Stdout, "  Connection: "+formatConnectionType(connection.GetType()))

	if connection.GetAddress() != "" {
		_, _ = io.WriteString(os.Stdout, " "+connection.GetAddress())
	}

	if connection.GetPort() > 0 {
		_, _ = fmt.Fprintf(os.Stdout, ":%d", connection.GetPort())
	}

	_, _ = io.WriteString(os.Stdout, "\n")
}

func outputPTZInfo(camera *protov1.Camera) {
	if camera.GetCurrentPtz() == nil {
		return
	}

	ptz := camera.GetCurrentPtz()
	_, _ = fmt.Fprintf(os.Stdout, "  PTZ: Pan=%.2f, Tilt=%.2f, Zoom=%.2f\n",
		ptz.GetPan(), ptz.GetTilt(), ptz.GetZoom())
}

func outputMetadata(camera *protov1.Camera) {
	if len(camera.GetMetadata()) == 0 {
		return
	}

	_, _ = io.WriteString(os.Stdout, "  Metadata: ")

	first := true
	for k, v := range camera.GetMetadata() {
		if !first {
			_, _ = io.WriteString(os.Stdout, ", ")
		}

		_, _ = fmt.Fprintf(os.Stdout, "%s=%s", k, v)

		first = false
	}

	_, _ = io.WriteString(os.Stdout, "\n")
}

func formatTimestamp(ms int64) string {
	if ms == 0 {
		return "Never"
	}

	timestamp := time.Unix(0, ms*int64(time.Millisecond))
	now := time.Now()
	diff := now.Sub(timestamp)

	if diff < time.Minute {
		return fmt.Sprintf("%ds ago", int(diff.Seconds()))
	}

	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}

	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}

	return timestamp.Format("2006-01-02 15:04:05")
}

func formatCameraMode(mode protov1.CameraMode) string {
	return mode.String()
}

func formatCameraStatus(status protov1.CameraStatus) string {
	return status.String()
}

func formatConnectionType(connType protov1.ConnectionType) string {
	return connType.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}
