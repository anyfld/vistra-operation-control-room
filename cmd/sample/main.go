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

type ptzScenarioStep struct {
	description string
	ptz         *protov1.PTZParameters
}

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
	fmt.Println("3. PTZサンプルシナリオを実行")
	fmt.Println("4. 終了")
	fmt.Println()
}

func registerCamera(
	ctx context.Context,
	client protov1connect.CameraServiceClient,
	scanner *bufio.Scanner,
) {
	const webrtcSampleURL = "http://localhost:1984/webrtc.html?src=camera&media=video+audio"

	fmt.Println("\n--- カメラ追加 ---")

	sampleCameras := generateSampleCameras(3)

	fmt.Println("サンプルカメラ:")

	for i, cam := range sampleCameras {
		fmt.Printf("%d. %s (%s)\n", i+1, cam.GetName(), cam.GetConnection().GetType().String())
	}

	fmt.Printf(
		"追加するカメラ番号を選択 (1-%d)\nサンプルでは WebRTC カメラ (%s) を利用します: ",
		len(sampleCameras),
		webrtcSampleURL,
	)

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
	fmt.Println("\n--- PTZサンプルシナリオ実行 ---")

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

	fmt.Println("\n選択したカメラでPTZサンプルシナリオを実行します。")

	getResp, err := cameraClient.GetCamera(
		ctx,
		connect.NewRequest(&protov1.GetCameraRequest{
			CameraId: selectedCamera.GetId(),
		}),
	)
	if err != nil {
		fmt.Printf("エラー: カメラ能力情報の取得に失敗しました: %v\n", err)

		return
	}

	capabilities := getResp.Msg.GetCapabilities()
	if capabilities == nil {
		fmt.Println(
			"警告: カメラ能力情報がないため、デフォルトのPTZシナリオを使用します。",
		)
	}

	scenario := buildPTZScenario(capabilities)

	for i, step := range scenario {
		fmt.Printf("\n%s\n", step.description)

		command := &protov1.ControlCommand{
			CommandId: fmt.Sprintf(
				"cmd-%d-%d",
				time.Now().UnixNano(),
				i+1,
			),
			CameraId:      selectedCamera.GetId(),
			Type:          protov1.ControlCommandType_CONTROL_COMMAND_TYPE_PTZ_ABSOLUTE,
			PtzParameters: step.ptz,
			TimeoutMs:     5000,
		}

		cmdResp, err := fdClient.StreamControlCommands(
			ctx,
			connect.NewRequest(&protov1.StreamControlCommandsRequest{
				Message: &protov1.StreamControlCommandsRequest_Command{
					Command: command,
				},
			}),
		)
		if err != nil {
			fmt.Printf(
				"エラー: シナリオ %d のPTZコマンド送信に失敗しました: %v\n",
				i+1,
				err,
			)

			return
		}

		result := cmdResp.Msg.GetResult()
		if result == nil {
			fmt.Println("エラー: コマンド結果が nil です。")

			return
		}

		fmt.Printf("  コマンドID: %s\n", result.GetCommandId())
		fmt.Printf("  成功: %v\n", result.GetSuccess())

		if result.GetErrorMessage() != "" {
			fmt.Printf(
				"  エラーメッセージ: %s\n",
				result.GetErrorMessage(),
			)
		}

		if result.GetResultingPtz() != nil {
			ptz := result.GetResultingPtz()
			fmt.Printf(
				"  結果PTZ: Pan=%.2f, Tilt=%.2f, Zoom=%.2f\n",
				ptz.GetPan(),
				ptz.GetTilt(),
				ptz.GetZoom(),
			)
		}

		fmt.Printf(
			"  実行時間: %d ms\n",
			result.GetExecutionTimeMs(),
		)

		if i != len(scenario)-1 {
			fmt.Println("  次のステップまで 1 秒待機します...")
			time.Sleep(1 * time.Second)
		}
	}

	fmt.Println("\n✓ PTZサンプルシナリオが完了しました。")
}

func generateSampleCameras(count int) []*protov1.RegisterCameraRequest {
	cameras := []*protov1.RegisterCameraRequest{
		{
			Name:       "Sample WebRTC Camera",
			Mode:       protov1.CameraMode_CAMERA_MODE_AUTONOMOUS,
			MasterMfId: "master-mf-001",
			Connection: &protov1.CameraConnection{
				Type: protov1.ConnectionType_CONNECTION_TYPE_WEBRTC,
				Address: "http://localhost:1984/" +
					"webrtc.html?src=camera&media=video+audio",
			},
			WebrtcConnectionName: "camera",
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

func buildPTZScenario(
	capabilities *protov1.CameraCapabilities,
) []ptzScenarioStep {
	const (
		defaultPanMin  = -180.0
		defaultPanMax  = 180.0
		defaultTiltMin = -45.0
		defaultTiltMax = 45.0
		defaultZoomMin = 1.0
		defaultZoomMax = 5.0
	)

	panMin := defaultPanMin
	panMax := defaultPanMax
	tiltMin := defaultTiltMin
	tiltMax := defaultTiltMax
	zoomMin := defaultZoomMin
	zoomMax := defaultZoomMax

	if capabilities != nil && capabilities.GetSupportsPtz() {
		if capabilities.GetPanMin() != 0 || capabilities.GetPanMax() != 0 {
			panMin = float64(capabilities.GetPanMin())
			panMax = float64(capabilities.GetPanMax())
		}

		if capabilities.GetTiltMin() != 0 || capabilities.GetTiltMax() != 0 {
			tiltMin = float64(capabilities.GetTiltMin())
			tiltMax = float64(capabilities.GetTiltMax())
		}

		if capabilities.GetZoomMin() != 0 || capabilities.GetZoomMax() != 0 {
			zoomMin = float64(capabilities.GetZoomMin())
			zoomMax = float64(capabilities.GetZoomMax())
		}
	}

	panCenter := (panMin + panMax) / 2.0
	tiltCenter := (tiltMin + tiltMax) / 2.0
	zoomRange := zoomMax - zoomMin

	if zoomRange <= 0 {
		zoomRange = defaultZoomMax - defaultZoomMin
		zoomMin = defaultZoomMin
		//nolint: ineffassign,wastedassign // zoomMax may need to be reset if it was modified
		zoomMax = defaultZoomMax
	}

	panRight := panCenter + (panMax-panCenter)*0.5
	panLeft := panCenter + (panMin-panCenter)*0.5
	tiltSlightDown := tiltCenter - (tiltCenter-tiltMin)*0.2

	zoomWide := zoomMin
	zoomMedium := zoomMin + zoomRange*0.5
	zoomClose := zoomMin + zoomRange*0.8

	return []ptzScenarioStep{
		{
			description: "1) ワイドショット (中央・広角)",
			ptz: &protov1.PTZParameters{
				Pan:       float32(panCenter),
				Tilt:      float32(tiltCenter),
				Zoom:      float32(zoomWide),
				PanSpeed:  0.5,
				TiltSpeed: 0.5,
				ZoomSpeed: 0.5,
			},
		},
		{
			description: "2) 右寄りクローズアップ",
			ptz: &protov1.PTZParameters{
				Pan:       float32(panRight),
				Tilt:      float32(tiltSlightDown),
				Zoom:      float32(zoomClose),
				PanSpeed:  0.6,
				TiltSpeed: 0.6,
				ZoomSpeed: 0.6,
			},
		},
		{
			description: "3) 左寄りミディアムショット",
			ptz: &protov1.PTZParameters{
				Pan:       float32(panLeft),
				Tilt:      float32(tiltCenter),
				Zoom:      float32(zoomMedium),
				PanSpeed:  0.5,
				TiltSpeed: 0.5,
				ZoomSpeed: 0.5,
			},
		},
		{
			description: "4) 中央に戻す",
			ptz: &protov1.PTZParameters{
				Pan:       float32(panCenter),
				Tilt:      float32(tiltCenter),
				Zoom:      float32(zoomMedium),
				PanSpeed:  0.5,
				TiltSpeed: 0.5,
				ZoomSpeed: 0.5,
			},
		},
	}
}
