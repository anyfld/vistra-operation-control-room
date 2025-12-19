package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	protov1 "github.com/anyfld/vistra-operation-control-room/gen/proto/v1"
	"github.com/anyfld/vistra-operation-control-room/gen/proto/v1/protov1connect"
	"github.com/anyfld/vistra-operation-control-room/internal/middleware"
	handlers "github.com/anyfld/vistra-operation-control-room/pkg/transport/handlers"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/infrastructure"
	"github.com/anyfld/vistra-operation-control-room/pkg/transport/usecase"
)

const (
	readHeaderTimeout = 10 * time.Second
	shutdownTimeout   = 30 * time.Second
)

type ExampleServiceHandler struct{}

func (h *ExampleServiceHandler) Ping(
	ctx context.Context,
	req *connect.Request[protov1.PingRequest],
) (*connect.Response[protov1.PingResponse], error) {
	return connect.NewResponse(&protov1.PingResponse{
		Message: "pong",
	}), nil
}

func main() {
	mux := setupHandlers()
	addr := getServerAddress()
	server := createServer(addr, mux)

	setupGracefulShutdown(server)

	log.Printf("Starting server on %s", addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func setupHandlers() *http.ServeMux {
	mux := http.NewServeMux()

	handler := &ExampleServiceHandler{}
	path, httpHandler := protov1connect.NewExampleServiceHandler(handler)
	mux.Handle(path, httpHandler)

	registerMDService(mux)
	registerCameraService(mux)
	registerCRService(mux)
	registerFDService(mux)

	return mux
}

func registerMDService(mux *http.ServeMux) {
	mdRepo := infrastructure.NewMDRepo()

	mdUC := usecase.NewMDUsecase(mdRepo)
	if path, h := protov1connect.NewMDServiceHandler(handlers.NewMDHandler(mdUC)); path != "" {
		mux.Handle(path, h)
	}
}

func registerCameraService(mux *http.ServeMux) {
	cameraRepo := infrastructure.NewCameraRepo()

	cameraUC := usecase.NewCameraUsecase(cameraRepo)
	if path, h := protov1connect.NewCameraServiceHandler(handlers.NewCameraHandler(cameraUC)); path != "" {
		mux.Handle(path, h)
	}
}
func registerCRService(mux *http.ServeMux) {
	uc := usecase.New(infrastructure.NewInMemoryRepo())
	if path, h := protov1connect.NewCRServiceHandler(handlers.NewCRHandler(uc)); path != "" {
		mux.Handle(path, h)
	}
}

func registerFDService(mux *http.ServeMux) {
	fdRepo := infrastructure.NewFDRepo()
	cameraRepo := infrastructure.NewCameraRepo()

	fdUC := usecase.NewFDUsecase(fdRepo)
	cameraUC := usecase.NewCameraUsecase(cameraRepo)
	if path, h := protov1connect.NewFDServiceHandler(handlers.NewFDHandler(fdUC, cameraUC)); path != "" {
		mux.Handle(path, h)
	}
}

func getServerAddress() string {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	return addr
}

func createServer(addr string, mux *http.ServeMux) *http.Server {
	return &http.Server{ //nolint:exhaustruct
		Addr: addr,
		Handler: middleware.Middleware(
			slog.New(slog.NewTextHandler(os.Stdout, nil)),
		)(h2c.NewHandler(mux, new(http2.Server))),
		ReadHeaderTimeout: readHeaderTimeout,
	}
}

func setupGracefulShutdown(server *http.Server) {
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		log.Println("Shutting down server...")

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()
}
