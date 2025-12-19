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
		outputJSON(cameras)
	} else {
		outputTable(cameras)
	}
}

func outputJSON(cameras []*protov1.Camera) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cameras); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}

func outputTable(cameras []*protov1.Camera) {
	fmt.Printf("Registered Cameras (%d):\n\n", len(cameras))
	fmt.Printf("%-20s %-30s %-20s %-15s %-20s %-15s\n",
		"ID", "Name", "Mode", "Status", "Master MF ID", "Last Seen")
	fmt.Println(strings.Repeat("-", 120))

	for _, camera := range cameras {
		lastSeen := formatTimestamp(camera.GetLastSeenAtMs())
		fmt.Printf("%-20s %-30s %-20s %-15s %-20s %-15s\n",
			camera.GetId(),
			truncate(camera.GetName(), 30),
			formatCameraMode(camera.GetMode()),
			formatCameraStatus(camera.GetStatus()),
			truncate(camera.GetMasterMfId(), 20),
			lastSeen,
		)

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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
