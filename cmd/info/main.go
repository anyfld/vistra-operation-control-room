package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
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
	jsonOutput := flag.Bool("json", false, "Output in JSON format")
	flag.Parse()

	ctx := context.Background()
	client := protov1connect.NewCameraServiceClient(
		&http.Client{Timeout: 5 * time.Second},
		*serverURL,
	)

	req := connect.NewRequest(&protov1.ListCamerasRequest{})

	resp, err := client.ListCameras(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list cameras: %v\n", err)
		os.Exit(1)
	}

	cameras := resp.Msg.GetCameras()
	if len(cameras) == 0 {
		fmt.Println("No cameras registered.")

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
		Camera     *protov1.Camera
		Connection *protov1.CameraConnection
		Status     protov1.CameraStatus
	}

	cameraInfos := make([]cameraInfo, 0, len(cameras))

	for _, camera := range cameras {
		info := cameraInfo{
			Camera: camera,
			Status: camera.GetStatus(),
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
		fmt.Fprintf(os.Stderr, "Error: failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}

func outputTable(
	cameras []*protov1.Camera,
	client protov1connect.CameraServiceClient,
	ctx context.Context,
) {
	fmt.Printf("Registered Cameras (%d):\n\n", len(cameras))
	fmt.Printf("%-20s %-30s %-20s %-15s %-20s %-15s\n",
		"ID", "Name", "Mode", "Status", "Master MF ID", "Last Seen")
	fmt.Println(strings.Repeat("-", 120))

	for _, camera := range cameras {
		lastSeen := formatTimestamp(camera.GetLastSeenAtMs())
		status := formatCameraStatus(camera.GetStatus())
		fmt.Printf("%-20s %-30s %-20s %-15s %-20s %-15s\n",
			camera.GetId(),
			truncate(camera.GetName(), 30),
			formatCameraMode(camera.GetMode()),
			status,
			truncate(camera.GetMasterMfId(), 20),
			lastSeen,
		)

		req := connect.NewRequest(&protov1.GetCameraRequest{
			CameraId: camera.GetId(),
		})

		resp, err := client.GetCamera(ctx, req)
		if err == nil && resp != nil {
			connection := resp.Msg.GetConnection()
			if connection != nil {
				fmt.Printf("  Connection: %s", formatConnectionType(connection.GetType()))

				if connection.GetAddress() != "" {
					fmt.Printf(" %s", connection.GetAddress())
				}

				if connection.GetPort() > 0 {
					fmt.Printf(":%d", connection.GetPort())
				}

				fmt.Println()
			}
		}

		if camera.GetCurrentPtz() != nil {
			ptz := camera.GetCurrentPtz()
			fmt.Printf("  PTZ: Pan=%.2f, Tilt=%.2f, Zoom=%.2f\n",
				ptz.GetPan(), ptz.GetTilt(), ptz.GetZoom())
		}

		if len(camera.GetMetadata()) > 0 {
			fmt.Print("  Metadata: ")

			first := true
			for k, v := range camera.GetMetadata() {
				if !first {
					fmt.Print(", ")
				}

				fmt.Printf("%s=%s", k, v)

				first = false
			}

			fmt.Println()
		}

		fmt.Println()
	}
}

func formatTimestamp(ms int64) string {
	if ms == 0 {
		return "Never"
	}

	t := time.Unix(0, ms*int64(time.Millisecond))
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return fmt.Sprintf("%ds ago", int(diff.Seconds()))
	}

	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}

	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}

	return t.Format("2006-01-02 15:04:05")
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
