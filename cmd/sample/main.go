package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	flag.Parse()

	ctx := context.Background()
	cameraClient := protov1connect.NewCameraServiceClient(
		&http.Client{Timeout: 5 * time.Second},
		*serverURL,
	)
	fdClient := protov1connect.NewFDServiceClient(
		&http.Client{Timeout: 5 * time.Second},
		*serverURL,
	)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		printMenu()
		fmt.Print("選択してください: ")

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			registerCamera(ctx, cameraClient, scanner)
		case "2":
			listCameras(ctx, cameraClient)
		case "3":
			sendPTZCommand(ctx, fdClient, cameraClient, scanner)
		case "4":
			fmt.Println("終了します。")
			return
		default:
			fmt.Println("無効な選択です。もう一度お試しください。")
		}

		fmt.Println()
	}
}

func printMenu() {
	fmt.Println("\n=== カメラ操作メニュー ===")
	fmt.Println("1. カメラを追加")
	fmt.Println("2. カメラ一覧を表示")
	fmt.Println("3. PTZコマンドを送信")
	fmt.Println("4. 終了")
	fmt.Println()
}

func registerCamera(
	ctx context.Context,
	client protov1connect.CameraServiceClient,
	scanner *bufio.Scanner,
) {
	fmt.Println("\n--- カメラ追加 ---")

	sampleCameras := generateSampleCameras(3)

	fmt.Println("サンプルカメラ:")
	for i, cam := range sampleCameras {
		fmt.Printf("%d. %s (%s)\n", i+1, cam.Name, cam.Connection.Type.String())
	}
	fmt.Print("追加するカメラ番号を選択 (1-3): ")

	if !scanner.Scan() {
		return
	}

	choiceStr := strings.TrimSpace(scanner.Text())
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(sampleCameras) {
		fmt.Println("無効な選択です。")
		return
	}

	req := sampleCameras[choice-1]

	resp, err := client.RegisterCamera(ctx, connect.NewRequest(req))
	if err != nil {
		fmt.Printf("エラー: カメラの登録に失敗しました: %v\n", err)
		return
	}

	camera := resp.Msg.GetCamera()
	if camera == nil {
		fmt.Println("エラー: カメラの登録結果が nil です。")
		return
	}

	fmt.Println("\n✓ カメラが登録されました:")
	fmt.Printf("  ID: %s\n", camera.GetId())
	fmt.Printf("  名前: %s\n", camera.GetName())
	fmt.Printf("  モード: %s\n", camera.GetMode().String())
	fmt.Printf("  Master MF ID: %s\n", camera.GetMasterMfId())
	fmt.Printf("  ステータス: %s\n", camera.GetStatus().String())
}

func listCameras(ctx context.Context, client protov1connect.CameraServiceClient) {
	fmt.Println("\n--- カメラ一覧 ---")

	resp, err := client.ListCameras(ctx, connect.NewRequest(&protov1.ListCamerasRequest{}))
	if err != nil {
		fmt.Printf("エラー: カメラ一覧の取得に失敗しました: %v\n", err)
		return
	}

	cameras := resp.Msg.GetCameras()
	if len(cameras) == 0 {
		fmt.Println("登録されているカメラがありません。")
		return
	}

	for i, camera := range cameras {
		fmt.Printf("\n%d. %s\n", i+1, camera.GetName())
		fmt.Printf("   ID: %s\n", camera.GetId())
		fmt.Printf("   モード: %s\n", camera.GetMode().String())
		fmt.Printf("   ステータス: %s\n", camera.GetStatus().String())
	}
}

func sendPTZCommand(
	ctx context.Context,
	fdClient protov1connect.FDServiceClient,
	cameraClient protov1connect.CameraServiceClient,
	scanner *bufio.Scanner,
) {
	fmt.Println("\n--- PTZコマンド送信 ---")

	resp, err := cameraClient.ListCameras(ctx, connect.NewRequest(&protov1.ListCamerasRequest{}))
	if err != nil {
		fmt.Printf("エラー: カメラ一覧の取得に失敗しました: %v\n", err)
		return
	}

	cameras := resp.Msg.GetCameras()
	if len(cameras) == 0 {
		fmt.Println("登録されているカメラがありません。")
		return
	}

	fmt.Println("カメラ一覧:")
	for i, camera := range cameras {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, camera.GetName(), camera.GetId())
	}
	fmt.Print("カメラ番号を選択: ")

	if !scanner.Scan() {
		return
	}

	choiceStr := strings.TrimSpace(scanner.Text())
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(cameras) {
		fmt.Println("無効な選択です。")
		return
	}

	selectedCamera := cameras[choice-1]

	fmt.Println("\nPTZコマンドタイプ:")
	fmt.Println("1. 絶対位置指定 (PTZ_ABSOLUTE)")
	fmt.Println("2. 相対位置指定 (PTZ_RELATIVE)")
	fmt.Println("3. 連続移動 (PTZ_CONTINUOUS)")
	fmt.Println("4. 停止 (PTZ_STOP)")
	fmt.Print("コマンドタイプを選択: ")

	if !scanner.Scan() {
		return
	}

	cmdTypeStr := strings.TrimSpace(scanner.Text())
	var cmdType protov1.ControlCommandType

	switch cmdTypeStr {
	case "1":
		cmdType = protov1.ControlCommandType_CONTROL_COMMAND_TYPE_PTZ_ABSOLUTE
	case "2":
		cmdType = protov1.ControlCommandType_CONTROL_COMMAND_TYPE_PTZ_RELATIVE
	case "3":
		cmdType = protov1.ControlCommandType_CONTROL_COMMAND_TYPE_PTZ_CONTINUOUS
	case "4":
		cmdType = protov1.ControlCommandType_CONTROL_COMMAND_TYPE_PTZ_STOP
	default:
		fmt.Println("無効な選択です。")
		return
	}

	var ptz *protov1.PTZParameters

	if cmdType != protov1.ControlCommandType_CONTROL_COMMAND_TYPE_PTZ_STOP {
		ptz = &protov1.PTZParameters{}

		fmt.Print("Pan角度 (-180.0 ~ 180.0): ")
		if scanner.Scan() {
			if panStr := strings.TrimSpace(scanner.Text()); panStr != "" {
				if pan, err := strconv.ParseFloat(panStr, 32); err == nil {
					ptz.Pan = float32(pan)
				}
			}
		}

		fmt.Print("Tilt角度 (-90.0 ~ 90.0): ")
		if scanner.Scan() {
			if tiltStr := strings.TrimSpace(scanner.Text()); tiltStr != "" {
				if tilt, err := strconv.ParseFloat(tiltStr, 32); err == nil {
					ptz.Tilt = float32(tilt)
				}
			}
		}

		fmt.Print("Zoom倍率 (1.0 ~ 10.0): ")
		if scanner.Scan() {
			if zoomStr := strings.TrimSpace(scanner.Text()); zoomStr != "" {
				if zoom, err := strconv.ParseFloat(zoomStr, 32); err == nil {
					ptz.Zoom = float32(zoom)
				}
			}
		}

		fmt.Print("Pan速度 (0.0 ~ 1.0): ")
		if scanner.Scan() {
			if panSpeedStr := strings.TrimSpace(scanner.Text()); panSpeedStr != "" {
				if panSpeed, err := strconv.ParseFloat(panSpeedStr, 32); err == nil {
					ptz.PanSpeed = float32(panSpeed)
				}
			}
		}

		fmt.Print("Tilt速度 (0.0 ~ 1.0): ")
		if scanner.Scan() {
			if tiltSpeedStr := strings.TrimSpace(scanner.Text()); tiltSpeedStr != "" {
				if tiltSpeed, err := strconv.ParseFloat(tiltSpeedStr, 32); err == nil {
					ptz.TiltSpeed = float32(tiltSpeed)
				}
			}
		}

		fmt.Print("Zoom速度 (0.0 ~ 1.0): ")
		if scanner.Scan() {
			if zoomSpeedStr := strings.TrimSpace(scanner.Text()); zoomSpeedStr != "" {
				if zoomSpeed, err := strconv.ParseFloat(zoomSpeedStr, 32); err == nil {
					ptz.ZoomSpeed = float32(zoomSpeed)
				}
			}
		}
	}

	command := &protov1.ControlCommand{
		CommandId:     fmt.Sprintf("cmd-%d", time.Now().UnixNano()),
		CameraId:      selectedCamera.GetId(),
		Type:          cmdType,
		PtzParameters: ptz,
		TimeoutMs:     5000,
	}

	cmdResp, err := fdClient.SendControlCommand(
		ctx,
		connect.NewRequest(&protov1.SendControlCommandRequest{
			Command: command,
		}),
	)
	if err != nil {
		fmt.Printf("エラー: PTZコマンドの送信に失敗しました: %v\n", err)
		return
	}

	result := cmdResp.Msg.GetResult()
	if result == nil {
		fmt.Println("エラー: コマンド結果が nil です。")
		return
	}

	fmt.Println("\n✓ PTZコマンドが送信されました:")
	fmt.Printf("  コマンドID: %s\n", result.GetCommandId())
	fmt.Printf("  成功: %v\n", result.GetSuccess())
	if result.GetErrorMessage() != "" {
		fmt.Printf("  エラーメッセージ: %s\n", result.GetErrorMessage())
	}
	if result.GetResultingPtz() != nil {
		ptz := result.GetResultingPtz()
		fmt.Printf("  結果PTZ: Pan=%.2f, Tilt=%.2f, Zoom=%.2f\n",
			ptz.GetPan(), ptz.GetTilt(), ptz.GetZoom())
	}
	fmt.Printf("  実行時間: %d ms\n", result.GetExecutionTimeMs())
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
